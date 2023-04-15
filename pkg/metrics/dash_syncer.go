package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
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
		dash, ok := dashMap[tpl.ID]
		if ok {
			if s.isDashChanged(ctx, dash) {
				continue
			}
		}

		builder := NewDashBuilder(projectID, dash)

		if err := builder.Build(tpl); err != nil {
			return fmt.Errorf("building dashboard %s failed: %w", tpl.ID, err)
		}

		if !builder.HasMetrics(metricMap) {
			continue
		}

		if builder.oldDash != nil {
			// Preserve some fields.
			builder.dash.Name = builder.oldDash.Name
			builder.dash.GridQuery = builder.oldDash.GridQuery
		}

		if err := builder.Save(ctx, s.app); err != nil {
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
		Model((*BaseGridColumn)(nil)).
		Where("dash_id = ?", dash.ID).
		Where("updated_at != created_at").
		Count(ctx)
	if err != nil {
		s.logger.Ctx(ctx).Error("countChangedColumns failed", zap.Error(err))
		return true
	}
	if n > 0 {
		return true
	}

	n, err = s.app.PG.NewSelect().
		Model((*DashGauge)(nil)).
		Where("dash_id = ?", dash.ID).
		Where("updated_at != created_at").
		Count(ctx)
	if err != nil {
		s.logger.Ctx(ctx).Error("countChangedGauges failed", zap.Error(err))
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
	oldDash   *Dashboard
	dash      *Dashboard
	gauges    []*DashGauge
	grid      []GridColumn
}

func NewDashBuilder(projectID uint32, oldDash *Dashboard) *DashBuilder {
	return &DashBuilder{
		projectID: projectID,
		oldDash:   oldDash,
	}
}

func (b *DashBuilder) Build(tpl *DashboardTpl) error {
	now := time.Now()
	b.dash = &Dashboard{
		ProjectID: b.projectID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := b.dash.FromTemplate(tpl); err != nil {
		return err
	}
	if err := b.dash.Validate(); err != nil {
		return err
	}

	for index, gauge := range tpl.Table.Gauges {
		if err := b.gauge(index, gauge, DashTable); err != nil {
			return err
		}
	}

	for index, gauge := range tpl.GridGauges {
		if err := b.gauge(index, gauge, DashGrid); err != nil {
			return err
		}
	}

	for _, tpl := range tpl.Grid {
		if err := b.gridColumn(tpl); err != nil {
			return err
		}
	}

	return nil
}

func (b *DashBuilder) gauge(index int, tpl *DashGaugeTpl, dashKind DashKind) error {
	gauge := &DashGauge{
		ProjectID: b.projectID,
		DashKind:  dashKind,
		Index:     sql.NullInt64{Int64: int64(index), Valid: true},
		CreatedAt: b.dash.CreatedAt,
		UpdatedAt: b.dash.UpdatedAt,
	}
	if err := gauge.FromTemplate(tpl); err != nil {
		return err
	}
	if err := gauge.Validate(); err != nil {
		return err
	}

	b.gauges = append(b.gauges, gauge)
	return nil
}

func (b *DashBuilder) gridColumn(tpl GridColumnTpl) error {
	switch tpl := tpl.(type) {
	case *ChartGridColumnTpl:
		col := &ChartGridColumn{
			BaseGridColumn: &BaseGridColumn{
				ProjectID: b.projectID,
				Type:      GridColumnChart,
				CreatedAt: b.dash.CreatedAt,
				UpdatedAt: b.dash.UpdatedAt,
			},
		}
		col.BaseGridColumn.Params.Any = &col.Params

		if err := col.FromTemplate(tpl); err != nil {
			return err
		}
		if err := col.Validate(); err != nil {
			return err
		}

		b.grid = append(b.grid, col)
		return nil
	case *TableGridColumnTpl:
		col := &TableGridColumn{
			BaseGridColumn: &BaseGridColumn{
				ProjectID: b.projectID,
				Type:      GridColumnTable,
				CreatedAt: b.dash.CreatedAt,
				UpdatedAt: b.dash.UpdatedAt,
			},
		}
		col.BaseGridColumn.Params.Any = &col.Params

		if err := col.FromTemplate(tpl); err != nil {
			return err
		}
		if err := col.Validate(); err != nil {
			return err
		}

		b.grid = append(b.grid, col)
		return nil
	case *HeatmapGridColumnTpl:
		col := &HeatmapGridColumn{
			BaseGridColumn: &BaseGridColumn{
				ProjectID: b.projectID,
				Type:      GridColumnHeatmap,
				CreatedAt: b.dash.CreatedAt,
				UpdatedAt: b.dash.UpdatedAt,
			},
		}
		col.BaseGridColumn.Params.Any = &col.Params

		if err := col.FromTemplate(tpl); err != nil {
			return err
		}
		if err := col.Validate(); err != nil {
			return err
		}

		b.grid = append(b.grid, col)
		return nil
	default:
		return fmt.Errorf("unsupported grid column template type: %T", tpl)
	}
}

func (b *DashBuilder) HasMetrics(metricMap map[string]*Metric) bool {
	for _, metric := range b.dash.TableMetrics {
		if _, ok := metricMap[metric.Name]; !ok {
			return false
		}
	}

gauge_loop:
	for i := len(b.gauges) - 1; i >= 0; i-- {
		gauge := b.gauges[i]
		for _, metric := range gauge.Metrics {
			if _, ok := metricMap[metric.Name]; !ok {

				b.gauges = append(b.gauges[:i], b.gauges[i+1:]...)
				continue gauge_loop
			}
		}
	}

grid_loop:
	for i := len(b.grid) - 1; i >= 0; i-- {
		col := b.grid[i]
		switch col := col.(type) {
		case *ChartGridColumn:
			for _, metric := range col.Params.Metrics {
				if _, ok := metricMap[metric.Name]; !ok {
					b.grid = append(b.grid[:i], b.grid[i+1:]...)
					continue grid_loop
				}
			}

		case *TableGridColumn:
			for _, metric := range col.Params.Metrics {
				if _, ok := metricMap[metric.Name]; !ok {
					b.grid = append(b.grid[:i], b.grid[i+1:]...)
					continue grid_loop
				}
			}

		case *HeatmapGridColumn:
			if _, ok := metricMap[col.Params.Metric]; !ok {
				b.grid = append(b.grid[:i], b.grid[i+1:]...)
				continue grid_loop
			}

		default:
			panic(fmt.Errorf("unsupported grid column type: %T", col))
		}

	}

	return len(b.grid) > 0
}

func (b *DashBuilder) Save(ctx context.Context, app *bunapp.App) error {
	if b.oldDash != nil {
		if err := DeleteDashboard(ctx, app, b.oldDash.ID); err != nil {
			return fmt.Errorf("DeleteDashboard failed: %w", err)
		}

		b.dash.ID = b.oldDash.ID
	}

	if err := InsertDashboard(ctx, app, b.dash); err != nil {
		return fmt.Errorf("InsertDashboard failed: %w", err)
	}

	for _, gauge := range b.gauges {
		gauge.DashID = b.dash.ID
	}

	baseCols := make([]*BaseGridColumn, len(b.grid))
	for i, col := range b.grid {
		baseCol := col.Base()
		baseCol.DashID = b.dash.ID
		baseCols[i] = baseCol
	}

	if len(b.gauges) > 0 {
		if err := InsertDashGauges(ctx, app, b.gauges); err != nil {
			return err
		}
	}

	if len(baseCols) > 0 {
		if err := InsertGridColumns(ctx, app, baseCols); err != nil {
			return err
		}
	}

	return nil
}
