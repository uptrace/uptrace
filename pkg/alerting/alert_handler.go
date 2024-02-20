package alerting

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

type AlertHandler struct {
	*bunapp.App
}

func NewAlertHandler(app *bunapp.App) *AlertHandler {
	return &AlertHandler{
		App: app,
	}
}

func (h *AlertHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	alertID, err := req.Params().Uint64("alert_id")
	if err != nil {
		return err
	}

	alert, err := SelectAlert(ctx, h.App, alertID)
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
	alert, err := SelectAlert(ctx, h.App, alertID)
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

	if err := changeAlertStatus(
		ctx,
		h.App,
		alert,
		status,
		userID,
	); err != nil {
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

	facets, err := selectAlertFacets(ctx, h.App, f)
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
	ctx context.Context, app *bunapp.App, f *AlertFilter,
) ([]*org.Facet, error) {
	facetMap, err := selectAlertFacetMap(ctx, app, f)
	if err != nil {
		return nil, err
	}

	delete(facetMap, attrkey.AlertType)

	statusFacet, err := selectAlertStatusFacet(ctx, app, f)
	if err != nil {
		return nil, err
	}

	typeFacet, err := selectAlertTypeFacet(ctx, app, f)
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
	ctx context.Context, app *bunapp.App, f *AlertFilter,
) (map[string]*org.Facet, error) {
	start := time.Now()

	facetMap, err := selectAlertFacetsForAttr(ctx, app, f, "")
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

		newFacetMap, err := selectAlertFacetsForAttr(ctx, app, f, attrKey)
		if err != nil {
			app.Zap(ctx).Error("selectAlertFacetsForAttr failed", zap.Error(err))
			continue
		}

		for key, src := range newFacetMap {
			facetMap[key] = src
		}
	}

	return facetMap, nil
}

func selectAlertFacetsForAttr(
	ctx context.Context, app *bunapp.App, f *AlertFilter, attrKey string,
) (map[string]*org.Facet, error) {
	searchq := app.PG.NewSelect().
		Model((*org.BaseAlert)(nil)).
		Join("JOIN alert_events AS event ON event.id = a.event_id").
		ColumnExpr("tsv").
		Apply(f.WhereClause).
		Limit(100e3)

	facetq := app.PG.NewSelect().
		ColumnExpr("split_part(word, '~~', 2) AS key").
		ColumnExpr("split_part(word, '~~', 3) AS value").
		ColumnExpr("ndoc AS count").
		ColumnExpr("row_number() OVER ("+
			"PARTITION BY split_part(word, '~~', 2) "+
			"ORDER BY ndoc DESC"+
			") AS rank").
		TableExpr("ts_stat($$ ? $$)", searchq).
		Where("starts_with(word, '~~')")

	q := app.PG.NewSelect().
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
	ctx context.Context, app *bunapp.App, f *AlertFilter,
) (*org.Facet, error) {
	f = f.Clone()
	f.Status = nil

	var items []*org.FacetItem

	if err := app.PG.NewSelect().
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
	ctx context.Context, app *bunapp.App, f *AlertFilter,
) (*org.Facet, error) {
	f = f.Clone()
	f.Type = nil

	var items []*org.FacetItem

	if err := app.PG.NewSelect().
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
