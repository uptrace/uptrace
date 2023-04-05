package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"github.com/uptrace/uptrace/pkg/org"
)

type DashGaugeHandler struct {
	*bunapp.App
}

func NewDashGaugeHandler(app *bunapp.App) *DashGaugeHandler {
	return &DashGaugeHandler{
		App: app,
	}
}

func (h *DashGaugeHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	query := req.URL.Query()
	dashKind := DashKind(query.Get("dash_kind"))

	gauges, err := SelectDashGauges(ctx, h.App, dash.ID, dashKind)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"gauges": gauges,
	})
}

type DashGaugeIn struct {
	DashKind DashKind `json:"dashKind"`

	Name        string `json:"name"`
	Description string `json:"description"`
	Template    string `json:"template"`

	Metrics   []upql.MetricAlias       `json:"metrics"`
	Query     string                   `json:"query"`
	ColumnMap map[string]*MetricColumn `json:"columnMap"`
}

func (in *DashGaugeIn) Validate(
	ctx context.Context, gauge *DashGauge,
) error {
	gauge.DashKind = in.DashKind

	gauge.Name = in.Name
	gauge.Description = in.Description
	gauge.Template = in.Template

	gauge.Query = in.Query
	gauge.Metrics = in.Metrics
	gauge.ColumnMap = in.ColumnMap

	gauge.UpdatedAt = time.Now()

	return gauge.Validate()
}

func (h *DashGaugeHandler) Create(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)
	dash := dashFromContext(ctx)

	in := new(DashGaugeIn)
	if err := httputil.UnmarshalJSON(w, req, in, 10<<10); err != nil {
		return err
	}

	dashGauge := new(DashGauge)
	dashGauge.ProjectID = project.ID
	dashGauge.DashID = dash.ID
	if err := in.Validate(ctx, dashGauge); err != nil {
		return httperror.Wrap(err)
	}

	if _, err := h.PG.NewInsert().
		Model(dashGauge).
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"gauge": dashGauge,
	})
}

func (h *DashGaugeHandler) Update(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dashGauge := dashGaugeFromContext(ctx)

	in := new(DashGaugeIn)
	if err := httputil.UnmarshalJSON(w, req, in, 10<<10); err != nil {
		return err
	}

	if err := in.Validate(ctx, dashGauge); err != nil {
		return httperror.Wrap(err)
	}

	if _, err := h.PG.NewUpdate().
		Model(dashGauge).
		Column("name", "description", "template", "query", "metrics", "column_map", "updated_at").
		Where("id = ?", dashGauge.ID).
		Where("project_id = ?", dashGauge.ProjectID).
		Where("dash_id = ?", dashGauge.DashID).
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"gauge": dashGauge,
	})
}

func (h *DashGaugeHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dashGauge := dashGaugeFromContext(ctx)

	if _, err := h.PG.
		NewDelete().
		Model(dashGauge).
		Where("id = ?", dashGauge.ID).
		Where("project_id = ?", dashGauge.ProjectID).
		Where("dash_id = ?", dashGauge.DashID).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"gauge": dashGauge,
	})
}

func (h *DashGaugeHandler) UpdateOrder(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)
	dash := dashFromContext(ctx)

	var gauges []DashGauge
	if err := httputil.UnmarshalJSON(w, req, &gauges, 10<<10); err != nil {
		return err
	}

	if _, err := h.PG.NewUpdate().
		With("_data", h.PG.NewValues(&gauges)).
		Model(&gauges).
		TableExpr("_data").
		Set("index = _data.index").
		Where("e.id = _data.id").
		Where("e.dash_id = ?", dash.ID).
		Where("e.project_id = ?", project.ID).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"gauges": gauges,
	})
}
