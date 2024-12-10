package metrics

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/fx"
	"gopkg.in/yaml.v3"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/unixtime"
)

type DashHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	Conf   *bunconf.Config
	PG     *bun.DB
}
type DashHandler struct {
	*DashHandlerParams
}

func NewDashHandler(p DashHandlerParams) *DashHandler {
	return &DashHandler{&p}
}

func registerDashHandler(h *DashHandler, p bunapp.RouterParams, m *Middleware) {
	p.RouterInternalV1.
		Use(m.UserAndProject).
		WithGroup("/metrics/:project_id/dashboards", func(g *bunrouter.Group) {
			g.POST("", h.Create)
			g.GET("", h.List)
			g.POST("/yaml", h.CreateFromYAML)

			g = g.NewGroup("/:dash_id").Use(m.Dashboard)

			g.GET("", h.Show)
			g.GET("/yaml", h.ShowYAML)
			g.POST("", h.Clone)
			g.PUT("", h.Update)
			g.PUT("/yaml", h.UpdateYAML)
			g.PUT("/table", h.UpdateTable)
			g.PUT("/grid", h.UpdateGrid)
			g.PUT("/reset", h.Reset)
			g.DELETE("", h.Delete)
			g.PUT("/pinned", h.Pin)
			g.PUT("/unpinned", h.Unpin)
		})
}

type DashboardIn struct {
	Name         string          `json:"name"`
	MinInterval  unixtime.Millis `json:"minInterval"`
	TimeOffset   unixtime.Millis `json:"timeOffset"`
	GridMaxWidth int             `json:"gridMaxWidth"`
}

func (in *DashboardIn) Populate(dash *Dashboard) error {
	dash.Name = in.Name
	dash.MinInterval = in.MinInterval
	dash.TimeOffset = in.TimeOffset
	dash.GridMaxWidth = in.GridMaxWidth
	return nil
}

func (h *DashHandler) Create(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	var in DashboardIn
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	dash := new(Dashboard)
	if err := in.Populate(dash); err != nil {
		return err
	}

	dash.ProjectID = project.ID
	if err := dash.Validate(); err != nil {
		return err
	}

	if err := InsertDashboard(ctx, h.PG, dash); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard": dash,
	})
}

func (h *DashHandler) Update(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	var in DashboardIn
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if err := in.Populate(dash); err != nil {
		return err
	}
	if err := dash.Validate(); err != nil {
		return err
	}

	// No need to update updated_at column, because we know how to preserve these columns.

	if _, err := h.PG.NewUpdate().
		Model(dash).
		Column("name", "min_interval", "time_offset", "grid_max_width").
		Where("id = ?", dash.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (h *DashHandler) UpdateGrid(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	var in struct {
		GridQuery string `json:"gridQuery"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	dash.GridQuery = in.GridQuery
	dash.UpdatedAt = time.Now()

	if err := dash.Validate(); err != nil {
		return err
	}

	if _, err := h.PG.NewUpdate().
		Column("grid_query", "updated_at").
		Model(dash).
		Where("id = ?", dash.ID).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (h *DashHandler) UpdateTable(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	var in struct {
		Name           string                  `json:"name"`
		TableMetrics   []mql.MetricAlias       `json:"tableMetrics"`
		TableQuery     string                  `json:"tableQuery"`
		TableColumnMap map[string]*TableColumn `json:"tableColumnMap"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	dash.Name = in.Name
	dash.TableMetrics = in.TableMetrics
	dash.TableQuery = in.TableQuery
	dash.TableColumnMap = in.TableColumnMap
	dash.UpdatedAt = time.Now()

	if err := dash.Validate(); err != nil {
		return err
	}

	if _, err := h.PG.NewUpdate().
		Column(
			"name",
			"table_metrics",
			"table_query",
			"table_grouping",
			"table_column_map",
			"updated_at",
		).
		Model(dash).
		Where("id = ?", dash.ID).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (h *DashHandler) Clone(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	tableItems, gridRows, err := h.selectDashItems(ctx, dash)
	if err != nil {
		return err
	}

	if err := h.PG.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		dash.ID = 0
		dash.Name += " clone"
		dash.TemplateID = ""

		if err := InsertDashboard(ctx, tx, dash); err != nil {
			return err
		}

		if len(tableItems) > 0 {
			for _, gridItem := range tableItems {
				base := gridItem.Base()
				base.ID = 0
				base.DashID = dash.ID
			}

			if err := InsertGridItems(ctx, tx, tableItems); err != nil {
				return err
			}
		}

		for _, gridRow := range gridRows {
			gridRow.ID = 0
			gridRow.DashID = dash.ID

			if err := InsertGridRow(ctx, tx, gridRow); err != nil {
				return err
			}

			for _, gridItem := range gridRow.Items {
				base := gridItem.Base()
				base.ID = 0
				base.DashID = dash.ID
				base.RowID = gridRow.ID
			}

			if err := InsertGridItems(ctx, tx, gridRow.Items); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard": dash,
	})
}

func (h *DashHandler) Reset(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	if dash.TemplateID == "" {
		return h.resetLayout(ctx, dash)
	}
	return h.resetFromTemplate(ctx, dash)
}

func (h *DashHandler) resetLayout(ctx context.Context, dash *Dashboard) error {
	tableItems, gridRows, err := h.selectDashItems(ctx, dash)
	if err != nil {
		return err
	}

	if len(tableItems) > 0 {
		if err := h.resetGridLayout(ctx, tableItems); err != nil {
			return err
		}
	}

	for _, gridRow := range gridRows {
		if len(gridRow.Items) == 0 {
			continue
		}
		if err := h.resetGridLayout(ctx, gridRow.Items); err != nil {
			return err
		}
	}

	return nil
}

func (h *DashHandler) resetGridLayout(ctx context.Context, gridItems []GridItem) error {
	if err := resetGridLayout(gridItems, true); err != nil {
		return err
	}

	if _, err := h.PG.NewUpdate().
		With("_data", h.PG.NewValues(&gridItems)).
		Model((*BaseGridItem)(nil)).
		TableExpr("_data").
		Set("width = _data.width").
		Set("height = _data.height").
		Set("x_axis = _data.x_axis").
		Set("y_axis = _data.y_axis").
		Where("g.id = _data.id").
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func resetGridLayout(gridItems []GridItem, force bool) error {
	var xAxis int
	var yAxis int
	var rowHeight int
	for _, gridItem := range gridItems {
		base := gridItem.Base()

		if force {
			base.Width = 0
			base.Height = 0
		}

		if err := gridItem.Validate(); err != nil {
			return err
		}

		// Should be after the Validate call which defines Width and Height.
		if xAxis+base.Width > 24 {
			yAxis += rowHeight
			xAxis = 0
			rowHeight = 0
		}

		base.XAxis = xAxis
		base.YAxis = yAxis

		xAxis += base.Width
		rowHeight = max(rowHeight, base.Height)
	}
	return nil
}

func (h *DashHandler) resetFromTemplate(ctx context.Context, dash *Dashboard) error {
	tpls, err := readDashboardTemplates()
	if err != nil {
		return err
	}

	var tpl *DashboardTpl
	for _, el := range tpls {
		if el.ID == dash.TemplateID {
			tpl = el
			break
		}
	}

	if tpl == nil {
		return httperror.NotFound("can't find dashboard template %q", dash.TemplateID)
	}

	metricMap, err := SelectMetricMap(ctx, h.PG, dash.ProjectID)
	if err != nil {
		return err
	}

	builder := NewDashBuilder(tpl, dash.ProjectID, metricMap)

	if err := builder.Build(); err != nil {
		return err
	}

	if err := h.PG.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return builder.Save(ctx, tx, dash, false)
	}); err != nil {
		return err
	}

	return nil
}

func (h *DashHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	if err := DeleteDashboard(ctx, h.PG, dash.ID); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard": dash,
	})
}

func (h *DashHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := new(DashFilter)
	if err := DecodeDashFilter(req, f); err != nil {
		return err
	}

	type DashboardOut struct {
		bun.BaseModel `bun:"dashboards,alias:d"`

		ID         uint64 `json:"id" bun:",pk,autoincrement"`
		ProjectID  uint32 `json:"projectId"`
		TemplateID string `json:"templateId" bun:",nullzero"`

		Name   string `json:"name"`
		Pinned bool   `json:"pinned"`

		CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
		UpdatedAt time.Time `json:"updatedAt" bun:",nullzero"`
	}
	dashboards := make([]*DashboardOut, 0)

	if err := h.PG.NewSelect().
		Model(&dashboards).
		Apply(f.WhereClause).
		Apply(f.PGOrder).
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

	tableItems, gridRows, err := h.selectDashItems(ctx, dash)
	if err != nil {
		return err
	}

	gridMetrics := make([]string, 0)
	seenMetrics := make(map[string]bool)
	for _, row := range gridRows {
		for _, item := range row.Items {
			for _, metricName := range item.Metrics() {
				if seenMetrics[metricName] {
					continue
				}
				seenMetrics[metricName] = true
				gridMetrics = append(gridMetrics, metricName)
			}
		}
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard":   dash,
		"tableItems":  tableItems,
		"gridRows":    gridRows,
		"gridMetrics": gridMetrics,
		"yamlUrl":     h.Conf.SiteURL("/api/v1/metrics/%d/dashboards/%d/yaml", dash.ProjectID, dash.ID),
	})
}

func (h *DashHandler) selectDashItems(
	ctx context.Context, dash *Dashboard,
) ([]GridItem, []*GridRow, error) {
	gridItems, err := SelectGridItems(ctx, h.PG, dash.ID)
	if err != nil {
		return nil, nil, err
	}

	gridRows, err := SelectGridRows(ctx, h.PG, dash.ID)
	if err != nil {
		return nil, nil, err
	}

	rowMap := make(map[uint64]*GridRow)
	for _, row := range gridRows {
		row.Items = make([]GridItem, 0)
		rowMap[row.ID] = row
	}

	tableItems := make([]GridItem, 0)

	for _, gridItem := range gridItems {
		base := gridItem.Base()

		if base.DashKind == DashKindTable {
			tableItems = append(tableItems, gridItem)
			continue
		}
		if base.RowID == 0 {
			continue
		}

		if row, ok := rowMap[base.RowID]; ok {
			row.Items = append(row.Items, gridItem)
		}
	}

	return tableItems, gridRows, nil
}

func (h *DashHandler) ShowYAML(w http.ResponseWriter, req bunrouter.Request) error {
	tpl, err := h.dashboardTpl(req)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	if err := enc.Encode(tpl); err != nil {
		return err
	}

	header := w.Header()
	header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.yml", tpl.ID))
	header.Set("Content-Type", "text/yaml")

	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (h *DashHandler) dashboardTpl(req bunrouter.Request) (*DashboardTpl, error) {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	tableItems, gridRows, err := h.selectDashItems(ctx, dash)
	if err != nil {
		return nil, err
	}

	tpl, err := NewDashboardTpl(dash, tableItems, gridRows)
	if err != nil {
		return nil, err
	}

	return tpl, nil
}

func (h *DashHandler) CreateFromYAML(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	tpl := new(DashboardTpl)

	dec := yaml.NewDecoder(req.Body)
	if err := dec.Decode(tpl); err != nil {
		return err
	}

	// Template id can't be set by a client.
	tpl.ID = ""

	builder := NewDashBuilder(tpl, project.ID, nil)

	if err := builder.Build(); err != nil {
		return err
	}

	if err := h.PG.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return builder.Save(ctx, tx, nil, false)
	}); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"dashboard": builder.dash,
	})
}

func (h *DashHandler) UpdateYAML(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	tpl := new(DashboardTpl)

	dec := yaml.NewDecoder(req.Body)
	if err := dec.Decode(tpl); err != nil {
		return err
	}

	builder := NewDashBuilder(tpl, dash.ProjectID, nil)

	if err := builder.Build(); err != nil {
		return err
	}

	if err := h.PG.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return builder.Save(ctx, tx, dash, false)
	}); err != nil {
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
