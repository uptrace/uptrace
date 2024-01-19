package metrics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/org"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type DashSyncer struct {
	app *bunapp.App

	templates []*DashboardTpl

	logger *otelzap.Logger
}

func NewDashSyncer(app *bunapp.App) *DashSyncer {
	templates, err := readDashboardTemplates()
	if err != nil {
		app.Logger.Error("readDashboardTemplates failed", zap.Error(err))
	}

	return &DashSyncer{
		app:       app,
		templates: templates,
		logger:    app.Logger,
	}
}

func (s *DashSyncer) CreateDashboardsHandler(ctx context.Context, projectID uint32) error {
	ctx, span := bunotel.Tracer.Start(ctx, "create-dashboards")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("project_id", int64(projectID)),
	)

	return s.app.PG.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var locked bool
		if err := tx.NewRaw("SELECT pg_try_advisory_xact_lock(?)", projectID).
			Scan(ctx, &locked); err != nil {
			return err
		}
		if !locked {
			fmt.Println("not locked")
			return nil
		}
		return s.createDashboards(ctx, projectID)
	})
}

func (s *DashSyncer) createDashboards(ctx context.Context, projectID uint32) error {
	metricMap, err := SelectMetricMap(ctx, s.app, projectID)
	if err != nil {
		return fmt.Errorf("SelectMetricMap failed: %w", err)
	}

	var dashboards []*Dashboard

	if err := s.app.PG.NewSelect().
		Model(&dashboards).
		Where("project_id = ?", projectID).
		Where("template_id IS NOT NULL").
		Scan(ctx); err != nil {
		return fmt.Errorf("SelectDashboardMap failed: %w", err)
	}

	dashMap := make(map[string]*Dashboard, len(dashboards))
	for _, dash := range dashboards {
		dashMap[dash.TemplateID] = dash
	}

	for _, tpl := range s.templates {
		existingDash, ok := dashMap[tpl.ID]
		if ok && s.isDashChanged(ctx, existingDash) {
			continue
		}

		builder := NewDashBuilder(projectID, metricMap)

		if err := builder.Build(tpl); err != nil {
			return fmt.Errorf("building dashboard %s failed: %w", tpl.ID, err)
		}

		if builder.IsEmpty() {
			if existingDash != nil {
				if err := DeleteDashboard(ctx, s.app.PG, existingDash.ID); err != nil {
					return fmt.Errorf("DeleteDashboard failed: %w", err)
				}
			}
			continue
		}

		if err := s.app.PG.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			return builder.Save(ctx, tx, existingDash, true)
		}); err != nil {
			return fmt.Errorf("saving dashboard %s failed: %w", tpl.ID, err)
		}
	}

	return nil
}

func (s *DashSyncer) isDashChanged(ctx context.Context, dash *Dashboard) bool {
	if !dash.CreatedAt.Equal(dash.UpdatedAt) {
		return true
	}

	n, err := s.app.PG.NewSelect().
		Model((*BaseGridItem)(nil)).
		Where("dash_id = ?", dash.ID).
		Where("updated_at != created_at").
		Count(ctx)
	if err != nil {
		return true
	}
	if n > 0 {
		return true
	}

	return false
}

//------------------------------------------------------------------------------

type DashBuilder struct {
	projectID uint32
	metricMap map[string]*Metric

	dash       *Dashboard
	tableItems []GridItem
	gridRows   []*GridRow
	monitors   []*org.MetricMonitor
}

func NewDashBuilder(projectID uint32, metricMap map[string]*Metric) *DashBuilder {
	return &DashBuilder{
		projectID: projectID,
		metricMap: metricMap,
	}
}

func (b *DashBuilder) Build(tpl *DashboardTpl) error {
	dash, err := b.dashboard(tpl)
	if err != nil {
		return fmt.Errorf("invalid dashboard: %w", err)
	}
	b.dash = dash

	for _, tpl := range tpl.Table.GridItems {
		gridItem, err := b.gridItem(tpl.Value)
		if err != nil {
			return fmt.Errorf("invalid table item: %w", err)
		}
		if b.hasGridItemMetrics(gridItem) {
			b.tableItems = append(b.tableItems, gridItem)
		}
	}

	for _, tpl := range tpl.GridRows {
		gridRow, err := b.gridRow(tpl)
		if err != nil {
			return fmt.Errorf("invalid grid row: %w", err)
		}
		if len(gridRow.Items) > 0 {
			b.gridRows = append(b.gridRows, gridRow)
		}
	}

	for _, tpl := range tpl.Monitors {
		monitor, err := b.monitor(tpl)
		if err != nil {
			return fmt.Errorf("invalid monitor: %w", err)
		}
		if b.hasMetrics(monitor.Params.Metrics) {
			b.monitors = append(b.monitors, monitor)
		}
	}

	return nil
}

func (b *DashBuilder) dashboard(tpl *DashboardTpl) (*Dashboard, error) {
	dash := &Dashboard{
		ProjectID: b.projectID,
	}
	if err := tpl.Populate(dash); err != nil {
		return nil, err
	}

	if b.metricMap != nil {
		tableQueryParts := mql.SplitQuery(dash.TableQuery)
		for i := len(dash.TableMetrics) - 1; i >= 0; i-- {
			metric := dash.TableMetrics[i]

			if b.hasMetric(metric.Name) {
				continue
			}

			dash.TableMetrics = append(dash.TableMetrics[:i], dash.TableMetrics[i+1:]...)

			for i := len(tableQueryParts) - 1; i >= 0; i-- {
				part := tableQueryParts[i]
				if strings.Contains(part, "$"+metric.Alias) {
					tableQueryParts = append(tableQueryParts[:i], tableQueryParts[i+1:]...)
				}
			}
		}
		dash.TableQuery = mql.JoinQuery(tableQueryParts)

		tableQuery, err := mql.ParseQueryError(dash.TableQuery)
		if err != nil {
			return nil, err
		}

		tableAttrMap := make(map[string]bool)
		for _, metric := range dash.TableMetrics {
			metric, ok := b.metricMap[metric.Name]
			if !ok {
				continue
			}
			for _, attr := range metric.AttrKeys {
				tableAttrMap[attr] = true
			}
		}

		for i := len(tableQuery.Parts) - 1; i >= 0; i-- {
			part := tableQuery.Parts[i]

			grouping, ok := part.AST.(*ast.Grouping)
			if !ok {
				continue
			}

			if len(grouping.Elems) != 1 {
				return nil, fmt.Errorf("group by with multiple elems: %q", part.Query)
			}

			elem := grouping.Elems[0]
			if !tableAttrMap[elem.Alias] {
				tableQuery.Parts = append(tableQuery.Parts[:i], tableQuery.Parts[i+1:]...)
			}
		}

		dash.TableQuery = tableQuery.String()
	}

	if len(dash.TableMetrics) > 0 {
		columnMap, err := mql.ParseColumns(dash.TableQuery)
		if err != nil {
			return nil, err
		}

		for name, col := range dash.TableColumnMap {
			expr, ok := columnMap[name]
			if !ok {
				delete(dash.TableColumnMap, name)
				continue
			}

			if col.AggFunc == "" {
				col.AggFunc = mql.TableFuncName(expr)
			}
		}
	}

	return dash, nil
}

func (b *DashBuilder) gridRow(tpl *GridRowTpl) (*GridRow, error) {
	row := new(GridRow)

	if err := tpl.Populate(row); err != nil {
		return nil, err
	}

	for _, itemTpl := range tpl.Items {
		gridItem, err := b.gridItem(itemTpl.Value)
		if err != nil {
			return nil, err
		}
		if b.hasGridItemMetrics(gridItem) {
			row.Items = append(row.Items, gridItem)
		}
	}

	return row, nil
}

func (b *DashBuilder) gridItem(tpl any) (GridItem, error) {
	switch tpl := tpl.(type) {
	case *ChartGridItemTpl:
		gridItem := NewChartGridItem()
		if err := tpl.Populate(gridItem); err != nil {
			return nil, err
		}
		return gridItem, nil

	case *TableGridItemTpl:
		gridItem := NewTableGridItem()
		if err := tpl.Populate(gridItem); err != nil {
			return nil, err
		}
		return gridItem, nil

	case *HeatmapGridItemTpl:
		gridItem := NewHeatmapGridItem()
		if err := tpl.Populate(gridItem); err != nil {
			return nil, err
		}
		return gridItem, nil

	case *GaugeGridItemTpl:
		gridItem := NewGaugeGridItem()
		if err := tpl.Populate(gridItem); err != nil {
			return nil, err
		}

		columnMap, err := mql.ParseColumns(gridItem.Params.Query)
		if err != nil {
			return nil, err
		}

		for name, col := range gridItem.Params.ColumnMap {
			expr, ok := columnMap[name]
			if !ok {
				delete(gridItem.Params.ColumnMap, name)
				continue
			}

			if col.AggFunc == "" {
				col.AggFunc = mql.TableFuncName(expr)
			}
		}

		return gridItem, nil

	case *GridItemTpl:
		return b.gridItem(tpl.Value)

	default:
		return nil, fmt.Errorf("unsupported grid column template type: %T", tpl)
	}
}

func (b *DashBuilder) monitor(tpl *MetricMonitorTpl) (*org.MetricMonitor, error) {
	monitor := org.NewMetricMonitor()
	if err := tpl.Populate(monitor); err != nil {
		return nil, err
	}

	monitor.ProjectID = b.projectID
	monitor.NotifyEveryoneByEmail = true

	if err := monitor.Validate(); err != nil {
		return nil, err
	}

	return monitor, nil
}

func (b *DashBuilder) hasGridItemMetrics(gridItem GridItem) bool {
	switch gridItem := gridItem.(type) {
	case *ChartGridItem:
		return b.hasMetrics(gridItem.Params.Metrics)
	case *TableGridItem:
		return b.hasMetrics(gridItem.Params.Metrics)
	case *HeatmapGridItem:
		return b.hasMetric(gridItem.Params.Metric)
	case *GaugeGridItem:
		return b.hasMetrics(gridItem.Params.Metrics)
	default:
		panic(fmt.Errorf("unsupported grid item type: %T", gridItem))
	}
}

func (b *DashBuilder) hasMetrics(metrics []mql.MetricAlias) bool {
	for _, metric := range metrics {
		if !b.hasMetric(metric.Name) {
			return false
		}
	}
	return true
}

func (b *DashBuilder) hasMetric(name string) bool {
	if b.metricMap == nil {
		return true
	}
	_, ok := b.metricMap[name]
	return ok
}

func (b *DashBuilder) IsEmpty() bool {
	return len(b.dash.TableMetrics) == 0 && len(b.gridRows) == 0
}

func (b *DashBuilder) Validate() error {
	if err := b.dash.Validate(); err != nil {
		return err
	}
	return nil
}

func (b *DashBuilder) Save(
	ctx context.Context, tx bun.Tx, existingDash *Dashboard, withMonitors bool,
) error {
	if existingDash != nil {
		if err := DeleteDashboard(ctx, tx, existingDash.ID); err != nil {
			return fmt.Errorf("DeleteDashboard failed: %w", err)
		}

		b.dash.ID = existingDash.ID
		b.dash.Name = existingDash.Name
		b.dash.Pinned = existingDash.Pinned
		b.dash.MinInterval = existingDash.MinInterval
		b.dash.TimeOffset = existingDash.TimeOffset
		b.dash.GridQuery = existingDash.GridQuery
	}

	now := time.Now()
	b.dash.CreatedAt = now
	b.dash.UpdatedAt = now

	if err := InsertDashboard(ctx, tx, b.dash); err != nil {
		return fmt.Errorf("InsertDashboard failed: %w", err)
	}

	if len(b.tableItems) > 0 {
		for _, gridItem := range b.tableItems {
			base := gridItem.Base()

			base.DashID = b.dash.ID
			base.DashKind = DashKindTable
			base.CreatedAt = now
			base.UpdatedAt = now

			if err := gridItem.Validate(); err != nil {
				return err
			}
		}

		if err := InsertGridItems(ctx, tx, b.tableItems); err != nil {
			return err
		}
	}

	for _, gridRow := range b.gridRows {
		gridRow.DashID = b.dash.ID
		gridRow.CreatedAt = now
		gridRow.UpdatedAt = now

		if err := InsertGridRow(ctx, tx, gridRow); err != nil {
			return err
		}

		for _, gridItem := range gridRow.Items {
			base := gridItem.Base()

			base.DashID = b.dash.ID
			base.DashKind = DashKindGrid
			base.RowID = gridRow.ID
			base.CreatedAt = now
			base.UpdatedAt = now

			if err := gridItem.Validate(); err != nil {
				return err
			}
		}

		if err := resetGridLayout(gridRow.Items, false); err != nil {
			return err
		}
		if err := InsertGridItems(ctx, tx, gridRow.Items); err != nil {
			return err
		}
	}

	if withMonitors && existingDash == nil {
		for _, monitor := range b.monitors {
			monitor.ProjectID = b.projectID

			if err := org.InsertMonitor(ctx, tx, monitor); err != nil {
				return err
			}
		}
	}

	return nil
}
