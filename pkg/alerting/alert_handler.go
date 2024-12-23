package alerting

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/pkg/clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/vmihailenco/taskq/v4"
)

type AlertHandlerParams struct {
	fx.In

	Logger    *otelzap.Logger
	Conf      *bunconf.Config
	PG        *bun.DB
	CH        *ch.DB
	MainQueue taskq.Queue
}

type AlertHandler struct {
	*AlertHandlerParams
}

func NewAlertHandler(p AlertHandlerParams) *AlertHandler {
	return &AlertHandler{&p}
}

func registerAlertHandler(h *AlertHandler, p bunapp.RouterParams, m *Middleware) {
	p.RouterInternalV1.NewGroup("/projects/:project_id",
		bunrouter.WithMiddleware(m.UserAndProject),
		bunrouter.WithGroup(func(g *bunrouter.Group) {
			g.GET("/alerts", h.List)
			g.GET("/alerts/:alert_id", h.Show)
			g.PUT("/alerts/closed", h.Close)
			g.PUT("/alerts/open", h.Open)
			g.DELETE("/alerts", h.Delete)
		}))
}

func (h *AlertHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	alertID, err := req.Params().Uint64("alert_id")
	if err != nil {
		return err
	}

	alert, err := SelectAlert(ctx, h.PG, alertID)
	if err != nil {
		return err
	}

	if alert.Base().ProjectID != project.ID {
		return org.ErrAccessDenied
	}

	return httputil.JSON(w, bunrouter.H{
		"alert": alert,
	})
}

func (h *AlertHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	var in struct {
		AlertIDs []uint64 `json:"alertIds"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if _, err := h.PG.NewDelete().
		Model((*org.BaseAlert)(nil)).
		Where("id IN (?)", bun.In(in.AlertIDs)).
		Where("project_id = ?", project.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (h *AlertHandler) Close(w http.ResponseWriter, req bunrouter.Request) error {
	return h.updateAlertsStatus(w, req, org.AlertStatusClosed)
}

func (h *AlertHandler) Open(w http.ResponseWriter, req bunrouter.Request) error {
	return h.updateAlertsStatus(w, req, org.AlertStatusOpen)
}

func (h *AlertHandler) updateAlertsStatus(
	w http.ResponseWriter, req bunrouter.Request, status org.AlertStatus,
) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)
	project := org.ProjectFromContext(ctx)

	var in struct {
		AlertIDs []uint64 `json:"alertIds"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if len(in.AlertIDs) == 0 {
		return errors.New("at least one alert is required")
	}
	if len(in.AlertIDs) > 1000 {
		in.AlertIDs = in.AlertIDs[:1000]
	}

	for _, alertID := range in.AlertIDs {
		if err := h.changeAlertStatus(ctx, user.ID, project.ID, alertID, status); err != nil {
			return err
		}
	}

	return httputil.JSON(w, bunrouter.H{})
}

func (h *AlertHandler) changeAlertStatus(
	ctx context.Context, userID uint64, projectID uint32, alertID uint64, status org.AlertStatus,
) error {
	alert, err := SelectAlert(ctx, h.PG, alertID)
	if err != nil {
		return err
	}
	baseAlert := alert.Base()

	if baseAlert.ProjectID != projectID {
		return httperror.Forbidden("you don't have enough permissions to update this alert")
	}
	if baseAlert.Event.Status == status {
		return nil
	}

	if err := tryAlertInTx(ctx, h.Logger, h.PG, h.CH, h.MainQueue, alert, func(tx bun.Tx) error {
		event := alert.GetEvent().Clone()
		baseEvent := event.Base()
		baseEvent.UserID = userID
		baseEvent.Name = org.AlertEventStatusChanged
		baseEvent.Status = status

		if err := org.InsertAlertEvent(ctx, tx, event); err != nil {
			return err
		}
		if err := updateAlertEvent(ctx, tx, alert, event); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (h *AlertHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := new(AlertFilter)
	if err := DecodeAlertFilter(req, f); err != nil {
		return err
	}

	alerts, count, err := h.selectAlerts(ctx, f)
	if err != nil {
		return err
	}

	facets, err := selectAlertFacets(ctx, h.Logger, h.PG, f)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"alerts": alerts,
		"count":  count,
		"facets": facets,
	})
}

func (h *AlertHandler) selectAlerts(
	ctx context.Context, f *AlertFilter,
) ([]*org.BaseAlert, int, error) {
	alerts := make([]*org.BaseAlert, 0)
	count, err := h.PG.NewSelect().
		Model(&alerts).
		Relation("Event").
		Where("event.id IS NOT NULL").
		Apply(f.WhereClause).
		Apply(f.PGOrder).
		Limit(1000).
		ScanAndCount(ctx)
	if err != nil {
		return nil, count, err
	}
	return alerts, count, nil
}

func selectAlertFacets(
	ctx context.Context,
	logger *otelzap.Logger,
	pg *bun.DB,
	f *AlertFilter,
) ([]*org.Facet, error) {
	facetMap, err := selectAlertFacetMap(ctx, logger, pg, f)
	if err != nil {
		return nil, err
	}

	delete(facetMap, attrkey.AlertType)

	statusFacet, err := selectAlertStatusFacet(ctx, pg, f)
	if err != nil {
		return nil, err
	}

	typeFacet, err := selectAlertTypeFacet(ctx, pg, f)
	if err != nil {
		return nil, err
	}

	var facets []*org.Facet
	if len(statusFacet.Items) > 0 {
		facets = append(facets, statusFacet)
	}
	if len(typeFacet.Items) > 0 {
		facets = append(facets, typeFacet)
	}
	facets = append(facets, org.FacetMapToList(facetMap)...)

	return facets, nil
}

func selectAlertFacetMap(
	ctx context.Context,
	logger *otelzap.Logger,
	pg *bun.DB,
	f *AlertFilter,
) (map[string]*org.Facet, error) {
	start := time.Now()

	facetMap, err := selectAlertFacetsForAttr(ctx, pg, f, "")
	if err != nil {
		return nil, err
	}

	for attrKey := range f.Attrs {
		if time.Since(start) > 5*time.Second {
			break
		}

		f := f.Clone()

		f.Attrs = maps.Clone(f.Attrs)
		delete(f.Attrs, attrKey)

		newFacetMap, err := selectAlertFacetsForAttr(ctx, pg, f, attrKey)
		if err != nil {
			logger.Error("selectAlertFacetsForAttr failed", zap.Error(err))
			continue
		}

		for key, src := range newFacetMap {
			facetMap[key] = src
		}
	}

	return facetMap, nil
}

func selectAlertFacetsForAttr(
	ctx context.Context, pg *bun.DB, f *AlertFilter, attrKey string,
) (map[string]*org.Facet, error) {
	searchq := pg.NewSelect().
		Model((*org.BaseAlert)(nil)).
		Join("JOIN alert_events AS event ON event.id = a.event_id").
		ColumnExpr("tsv").
		Apply(f.WhereClause).
		Limit(100e3)

	facetq := pg.NewSelect().
		ColumnExpr("split_part(word, '~~', 2) AS key").
		ColumnExpr("split_part(word, '~~', 3) AS value").
		ColumnExpr("ndoc AS count").
		ColumnExpr("row_number() OVER ("+
			"PARTITION BY split_part(word, '~~', 2) "+
			"ORDER BY ndoc DESC"+
			") AS rank").
		TableExpr("ts_stat($$ ? $$)", searchq).
		Where("starts_with(word, '~~')")

	q := pg.NewSelect().
		With("q", facetq).
		ColumnExpr("key, value, count").
		TableExpr("q").
		OrderExpr("value ASC")

	if attrKey != "" {
		q = q.Where("key = ?", attrKey)
	}

	var items []*org.FacetItem

	if err := q.Scan(ctx, &items); err != nil {
		return nil, err
	}

	return org.BuildFacetMap(items), nil
}

func selectAlertStatusFacet(
	ctx context.Context, pg *bun.DB, f *AlertFilter,
) (*org.Facet, error) {
	f = f.Clone()
	f.Status = nil

	var items []*org.FacetItem

	if err := pg.NewSelect().
		Model((*org.BaseAlert)(nil)).
		Join("JOIN alert_events AS event ON event.id = a.event_id").
		ColumnExpr("? AS key", attrkey.AlertStatus).
		ColumnExpr("event.status AS value").
		ColumnExpr("count(*) AS count").
		Apply(f.WhereClause).
		GroupExpr("event.status").
		OrderExpr("value ASC").
		Scan(ctx, &items); err != nil {
		return nil, err
	}

	if len(items) > 0 && !hasOpenStatus(items) {
		items = append(items, &org.FacetItem{
			Key:   attrkey.AlertStatus,
			Value: string(org.AlertStatusOpen),
			Count: 0,
		})
	}

	return &org.Facet{
		Key:   attrkey.AlertStatus,
		Items: items,
	}, nil
}

func hasOpenStatus(items []*org.FacetItem) bool {
	for _, item := range items {
		if item.Value == string(org.AlertStatusOpen) {
			return true
		}
	}
	return false
}

func selectAlertTypeFacet(
	ctx context.Context, pg *bun.DB, f *AlertFilter,
) (*org.Facet, error) {
	f = f.Clone()
	f.Type = nil

	var items []*org.FacetItem

	if err := pg.NewSelect().
		Model((*org.BaseAlert)(nil)).
		Join("JOIN alert_events AS event ON event.id = a.event_id").
		ColumnExpr("? AS key", attrkey.AlertType).
		ColumnExpr("type AS value").
		ColumnExpr("count(*) AS count").
		Apply(f.WhereClause).
		GroupExpr("type").
		OrderExpr("value ASC").
		Scan(ctx, &items); err != nil {
		return nil, err
	}

	return &org.Facet{
		Key:   attrkey.AlertType,
		Items: items,
	}, nil
}
