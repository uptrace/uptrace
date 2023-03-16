package metrics

import (
	"errors"
	"net/http"

	"github.com/uptrace/bunrouter"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"github.com/uptrace/uptrace/pkg/org"
)

var errPrebuiltDashboard = errors.New("you can't edit pre-built dashboards (clone instead)")

type DashHandler struct {
	App *bunapp.App
}

func NewDashHandler(app *bunapp.App) *DashHandler {
	return &DashHandler{
		App: app,
	}
}

func (h *DashHandler) Create(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	var in struct {
		Name string `json:"name"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	dash := new(Dashboard)
	dash.Name = in.Name
	dash.ProjectID = project.ID
	if dash.Columns == nil {
		dash.Columns = make(map[string]*MetricColumn)
	}

	if err := InsertDashboard(ctx, h.App, dash); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard": dash,
	})
}

func (h *DashHandler) Update(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	if dash.TemplateID != "" {
		return errPrebuiltDashboard
	}

	var in struct {
		Name      *string `json:"name"`
		BaseQuery *string `json:"baseQuery"`

		IsTable *bool                    `json:"isTable"`
		Metrics []upql.Metric            `json:"metrics"`
		Query   *string                  `json:"query"`
		Columns map[string]*MetricColumn `json:"columnMap"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	q := h.App.DB.NewUpdate().
		Model(dash).
		Where("id = ?", dash.ID).
		Returning("*")

	if in.Name != nil {
		dash.Name = *in.Name
		q = q.Column("name")
	}
	if in.BaseQuery != nil {
		dash.BaseQuery = *in.BaseQuery
		q = q.Column("base_query")
	}
	if in.IsTable != nil {
		dash.IsTable = *in.IsTable
		q = q.Column("is_table")
	}
	if in.Metrics != nil {
		dash.Metrics = in.Metrics
		q = q.Column("metrics")
	}
	if in.Query != nil {
		dash.Query = *in.Query
		q = q.Column("query")
	}
	if in.Columns != nil {
		dash.Columns = in.Columns
		q = q.Column("columns")
	}

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard": dash,
	})
}

func (h *DashHandler) Clone(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	gauges, err := SelectDashGauges(ctx, h.App, dash.ID)
	if err != nil {
		return err
	}

	entries, err := SelectDashEntries(ctx, h.App, dash)
	if err != nil {
		return err
	}

	dash.ID = 0
	dash.Name += " clone"
	dash.TemplateID = ""

	if err := InsertDashboard(ctx, h.App, dash); err != nil {
		return err
	}

	for _, gauge := range gauges {
		gauge.ID = 0
		gauge.DashID = dash.ID
	}

	if err := InsertDashGauges(ctx, h.App, gauges); err != nil {
		return err
	}

	for _, entry := range entries {
		entry.ID = 0
		entry.DashID = dash.ID
	}

	if err := InsertDashEntries(ctx, h.App, entries); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard": dash,
	})
}

func (h *DashHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	if err := DeleteDashboard(ctx, h.App, dash.ID); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard": dash,
	})
}

func (h *DashHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	dashboards := make([]*Dashboard, 0)

	if err := h.App.DB.NewSelect().
		Model(&dashboards).
		Where("project_id = ?", project.ID).
		OrderExpr("name ASC").
		Scan(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboards": dashboards,
	})
}

func (h *DashHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dashboard := dashFromContext(ctx)

	tableGauges, gridGauges, err := SelectTableGridGauges(ctx, h.App, dashboard.ID)
	if err != nil {
		return err
	}

	entries, err := SelectDashEntries(ctx, h.App, dashboard)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard":   dashboard,
		"tableGauges": tableGauges,
		"gridGauges":  gridGauges,
		"entries":     entries,
	})
}
