package alerting

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

const (
	facetKeyState = "alert.state"
	facetKeyType  = "alert.type"
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

func (h *AlertHandler) Close(w http.ResponseWriter, req bunrouter.Request) error {
	return h.updateAlertsState(w, req, org.AlertClosed)
}

func (h *AlertHandler) Open(w http.ResponseWriter, req bunrouter.Request) error {
	return h.updateAlertsState(w, req, org.AlertOpen)
}

func (h *AlertHandler) updateAlertsState(
	w http.ResponseWriter, req bunrouter.Request, state org.AlertState,
) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)
	project := org.ProjectFromContext(ctx)

	var in struct {
		AlertIDs []string `json:"alertIds"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if len(in.AlertIDs) == 0 {
		return errors.New("at least one alert is required")
	}
	if len(in.AlertIDs) > 100 {
		return fmt.Errorf("got %d alerts, wanted <= 100", len(in.AlertIDs))
	}

	for _, alertID := range in.AlertIDs {
		alertID, err := strconv.ParseUint(alertID, 10, 64)
		if err != nil {
			return err
		}

		if err := h.updateAlertState(ctx, user, project.ID, alertID, state); err != nil {
			return err
		}
	}

	return httputil.JSON(w, bunrouter.H{})
}

func (h *AlertHandler) updateAlertState(
	ctx context.Context, user *org.User, projectID uint32, alertID uint64, state org.AlertState,
) error {
	alert, err := SelectAlert(ctx, h.App, alertID)
	if err != nil {
		return err
	}
	baseAlert := alert.Base()

	if baseAlert.ProjectID != projectID {
		return httperror.Forbidden("you don't have enough permissions to update this alert")
	}
	if baseAlert.State == state {
		return nil
	}

	if err := updateAlertState(
		ctx,
		h.App,
		alert,
		state,
		user.ID,
	); err != nil {
		return err
	}

	return nil
}

func (h *AlertHandler) CloseAll(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	if _, err := h.PG.NewUpdate().
		Model((*org.BaseAlert)(nil)).
		Set("state = ?", org.AlertClosed).
		Where("project_id = ?", project.ID).
		Where("state = ?", org.AlertOpen).
		Exec(ctx); err != nil {
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

	alerts, count, err := SelectAlerts(ctx, h.App, f)
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

func selectAlertFacets(
	ctx context.Context, app *bunapp.App, f *AlertFilter,
) ([]*org.Facet, error) {
	facetMap, err := selectAlertFacetMap(ctx, app, f)
	if err != nil {
		return nil, err
	}

	delete(facetMap, facetKeyType)

	stateFacet, err := selectAlertStateFacet(ctx, app, f)
	if err != nil {
		return nil, err
	}

	typeFacet, err := selectAlertTypeFacet(ctx, app, f)
	if err != nil {
		return nil, err
	}

	var facets []*org.Facet
	if len(stateFacet.Items) > 0 {
		facets = append(facets, stateFacet)
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
			app.Zap(ctx).Error("_selectAlertFacetsForAttr failed", zap.Error(err))
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

func selectAlertStateFacet(
	ctx context.Context, app *bunapp.App, f *AlertFilter,
) (*org.Facet, error) {
	f = f.Clone()
	f.State = nil

	var items []*org.FacetItem

	if err := app.PG.NewSelect().
		ColumnExpr("? AS key", facetKeyState).
		ColumnExpr("state AS value").
		ColumnExpr("count(*) AS count").
		Model((*org.BaseAlert)(nil)).
		Apply(f.WhereClause).
		GroupExpr("state").
		OrderExpr("value ASC").
		Scan(ctx, &items); err != nil {
		return nil, err
	}

	return &org.Facet{
		Key:   facetKeyState,
		Items: items,
	}, nil
}

func hasOpenState(items []*org.FacetItem) bool {
	for _, item := range items {
		if item.Value == string(org.AlertOpen) {
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
		ColumnExpr("? AS key", facetKeyType).
		ColumnExpr("type AS value").
		ColumnExpr("count(*) AS count").
		Model((*org.BaseAlert)(nil)).
		Apply(f.WhereClause).
		GroupExpr("type").
		OrderExpr("value ASC").
		Scan(ctx, &items); err != nil {
		return nil, err
	}

	return &org.Facet{
		Key:   facetKeyType,
		Items: items,
	}, nil
}
