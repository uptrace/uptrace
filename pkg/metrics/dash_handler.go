package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/unixtime"
	"gopkg.in/yaml.v3"
)

type DashHandler struct {
	*bunapp.App
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

	if err := dash.Validate(); err != nil {
		return err
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

	var in struct {
		Name        *string          `json:"name"`
		GridQuery   *string          `json:"gridQuery"`
		MinInterval *unixtime.Millis `json:"minInterval"`
		TimeOffset  *unixtime.Millis `json:"timeOffset"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	q := h.PG.NewUpdate().
		Model(dash).
		Where("id = ?", dash.ID).
		Returning("*")

	if in.Name != nil {
		dash.Name = *in.Name
		q = q.Column("name")
	}
	if in.GridQuery != nil {
		dash.GridQuery = *in.GridQuery
		q = q.Column("grid_query")
	}
	if in.MinInterval != nil {
		dash.MinInterval = *in.MinInterval
		q = q.Column("min_interval")
	}
	if in.TimeOffset != nil {
		dash.TimeOffset = *in.TimeOffset
		q = q.Column("time_offset")
	}

	if err := dash.Validate(); err != nil {
		return err
	}

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard": dash,
	})
}

func (h *DashHandler) UpdateTable(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	var in struct {
		TableMetrics   []mql.MetricAlias        `json:"tableMetrics"`
		TableQuery     string                   `json:"tableQuery"`
		TableColumnMap map[string]*MetricColumn `json:"tableColumnMap"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	dash.TableMetrics = in.TableMetrics
	dash.TableQuery = in.TableQuery
	dash.TableColumnMap = in.TableColumnMap
	dash.UpdatedAt = time.Now()

	if err := dash.Validate(); err != nil {
		return err
	}

	if _, err := h.PG.NewUpdate().
		Column("table_metrics", "table_query", "table_grouping", "table_column_map", "updated_at").
		Model(dash).
		Where("id = ?", dash.ID).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard": dash,
	})
}

func (h *DashHandler) Clone(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	gauges, err := SelectDashGauges(ctx, h.App, dash.ID, "")
	if err != nil {
		return err
	}

	grid, err := SelectBaseGridColumns(ctx, h.App, dash.ID)
	if err != nil {
		return err
	}

	dash.ID = 0
	dash.Name += " clone"
	dash.TemplateID = ""

	if err := InsertDashboard(ctx, h.App, dash); err != nil {
		return err
	}

	if len(gauges) > 0 {
		for _, gauge := range gauges {
			gauge.ID = 0
			gauge.DashID = dash.ID
		}

		if err := InsertDashGauges(ctx, h.App, gauges); err != nil {
			return err
		}
	}

	if len(grid) > 0 {
		for _, col := range grid {
			col.ID = 0
			col.DashID = dash.ID
		}

		if err := InsertGridColumns(ctx, h.App, grid); err != nil {
			return err
		}
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

	if err := h.PG.NewSelect().
		Model(&dashboards).
		Where("project_id = ?", project.ID).
		OrderExpr("pinned DESC, name ASC").
		Limit(1000).
		Scan(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboards": dashboards,
	})
}

func (h *DashHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	grid, err := SelectBaseGridColumns(ctx, h.App, dash.ID)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard": dash,
		"grid":      grid,
		"yamlUrl": h.SiteURL(
			fmt.Sprintf("/internal/v1/metrics/%d/dashboards/%d/yaml", dash.ProjectID, dash.ID),
		),
	})
}

func (h *DashHandler) ShowYAML(w http.ResponseWriter, req bunrouter.Request) error {
	tpl, err := h.dashboardTpl(req)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(tpl)
	if err != nil {
		return err
	}

	header := w.Header()
	header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.yaml", tpl.ID))
	header.Set("Content-Type", "text/yaml")

	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}

func (h *DashHandler) dashboardTpl(req bunrouter.Request) (*DashboardTpl, error) {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	grid, err := SelectGridColumns(ctx, h.App, dash.ID)
	if err != nil {
		return nil, err
	}

	tableGauges, err := SelectDashGauges(ctx, h.App, dash.ID, DashTable)
	if err != nil {
		return nil, err
	}

	gridGauges, err := SelectDashGauges(ctx, h.App, dash.ID, DashTable)
	if err != nil {
		return nil, err
	}

	tpl, err := NewDashboardTpl(dash, grid, tableGauges, gridGauges)
	if err != nil {
		return nil, err
	}

	return tpl, nil
}

func (h *DashHandler) FromYAML(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	tpl := new(DashboardTpl)

	dec := yaml.NewDecoder(req.Body)
	if err := yamlUnmarshalDashboardTpl(dec, tpl); err != nil {
		return err
	}

	builder := NewDashBuilder(dash.ProjectID, dash)
	if err := builder.Build(tpl); err != nil {
		return err
	}
	if err := builder.Save(ctx, h.App); err != nil {
		return err
	}

	return nil
}

func (h *DashHandler) Pin(w http.ResponseWriter, req bunrouter.Request) error {
	return h.updatePinned(w, req, true)
}

func (h *DashHandler) Unpin(w http.ResponseWriter, req bunrouter.Request) error {
	return h.updatePinned(w, req, false)
}

func (h *DashHandler) updatePinned(
	w http.ResponseWriter, req bunrouter.Request, pinned bool,
) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	if _, err := h.PG.NewUpdate().
		Model((*Dashboard)(nil)).
		Where("id = ?", dash.ID).
		Set("pinned = ?", pinned).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}
