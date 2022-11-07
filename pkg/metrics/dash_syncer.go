package metrics

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"go.uber.org/zap"
)

type DashSyncer struct {
	app *bunapp.App

	templates []*DashboardTpl

	debouncerMapMu sync.Mutex
	debouncerMap   map[uint32]*bunutil.Debouncer

	logger *otelzap.Logger
}

func NewDashSyncer(app *bunapp.App) *DashSyncer {
	templates, err := readDashboardTemplates()
	if err != nil {
		app.Logger.Error("readDashboardTemplates failed", zap.Error(err))
	}

	s := &DashSyncer{
		app:          app,
		templates:    templates,
		debouncerMap: make(map[uint32]*bunutil.Debouncer),
		logger:       app.Logger,
	}

	ctx := app.Context()
	projects := app.Config().Projects
	for i := range projects {
		s.Sync(ctx, projects[i].ID)
	}

	return s
}

func (s *DashSyncer) Sync(ctx context.Context, projectID uint32) {
	s.debouncerMapMu.Lock()
	defer s.debouncerMapMu.Unlock()

	debouncer, ok := s.debouncerMap[projectID]
	if !ok {
		debouncer = bunutil.NewDebouncer()
		s.debouncerMap[projectID] = debouncer
	}

	debouncer.Run(10*time.Second, func() {
		_ = bunotel.RunWithNewRoot(ctx, "sync-dashboards", func(ctx context.Context) error {
			return s.syncDashboards(ctx, projectID)
		})
	})
}

func (s *DashSyncer) syncDashboards(ctx context.Context, projectID uint32) error {
	dashMap, err := SelectDashboardMap(ctx, s.app, projectID)
	if err != nil {
		return fmt.Errorf("SelectDashboardMap failed: %w", err)
	}

	metricMap, err := SelectMetricMap(ctx, s.app, projectID)
	if err != nil {
		return fmt.Errorf("SelectMetricMap failed: %w", err)
	}

	for _, tpl := range s.templates {
		dash, ok := dashMap[tpl.ID]
		if !ok {
			dash = &Dashboard{
				TemplateID: tpl.ID,
				ProjectID:  projectID,
			}
		}

		builder := &DashBuilder{
			metricMap: metricMap,
			dash:      dash,
			logger:    s.logger,
		}
		if err := builder.Build(tpl); err != nil {
			return fmt.Errorf("building dashboard %s failed: %w", tpl.ID, err)
		}

		if dash.ID != 0 {
			if err := DeleteDashboard(ctx, s.app, dash.ID); err != nil {
				return fmt.Errorf("DeleteDashboard failed: %w", err)
			}
		}

		if err := builder.Save(ctx, s.app); err != nil {
			return fmt.Errorf("saving dashboard %s failed: %w", tpl.ID, err)
		}
	}

	return nil
}

type DashBuilder struct {
	metricMap map[string]*Metric
	dash      *Dashboard
	gauges    []*DashGauge
	entries   []*DashEntry
	logger    *otelzap.Logger
}

func (b *DashBuilder) Build(tpl *DashboardTpl) error {
	metrics, err := upql.ParseMetrics(tpl.Table.Metrics)
	if err != nil {
		return err
	}

	b.dash.Name = tpl.Name
	b.dash.Metrics = metrics
	b.dash.Query = strings.Join(tpl.Table.Query, " | ")
	b.dash.Columns = tpl.Table.Columns
	b.dash.IsTable = len(b.dash.Metrics) > 0 && b.dash.Query != ""

	for _, gauge := range tpl.Table.Gauges {
		if err := b.gauge(gauge, DashTable); err != nil {
			return err
		}
	}

	for _, gauge := range tpl.Gauges {
		if err := b.gauge(gauge, DashGrid); err != nil {
			return err
		}
	}

	for _, entry := range tpl.Entries {
		if err := b.entry(entry); err != nil {
			return err
		}
	}

	return nil
}

func (b *DashBuilder) gauge(tpl *DashGaugeTpl, dashKind string) error {
	metrics, err := upql.ParseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		if _, ok := b.metricMap[metric.Name]; !ok {
			b.logger.Debug("metric not found",
				zap.Uint32("project_id", b.dash.ProjectID),
				zap.String("metric", metric.Name))
			return nil
		}
	}

	b.gauges = append(b.gauges, &DashGauge{
		DashKind:    dashKind,
		Name:        tpl.Name,
		Description: tpl.Description,
		Template:    tpl.Template,
		Metrics:     metrics,
		Query:       strings.Join(tpl.Query, " | "),
		Columns:     tpl.Columns,
	})
	return nil
}

func (b *DashBuilder) entry(tpl *DashEntryTpl) error {
	metrics, err := upql.ParseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		if _, ok := b.metricMap[metric.Name]; !ok {
			b.logger.Debug("metric not found",
				zap.Uint32("project_id", b.dash.ProjectID),
				zap.String("metric", metric.Name))
			return nil
		}
	}

	b.entries = append(b.entries, &DashEntry{
		Name:        tpl.Name,
		Description: tpl.Description,
		ChartType:   tpl.ChartType,
		Metrics:     metrics,
		Query:       strings.Join(tpl.Query, " | "),
		Columns:     tpl.Columns,
	})
	return nil
}

func (b *DashBuilder) Save(ctx context.Context, app *bunapp.App) error {
	if !b.dash.IsTable && len(b.entries) == 0 {
		return nil
	}

	if err := b.dash.Validate(); err != nil {
		return err
	}

	for _, metric := range b.dash.Metrics {
		if _, ok := b.metricMap[metric.Name]; !ok {
			b.logger.Debug("metric not found",
				zap.Uint32("project_id", b.dash.ProjectID),
				zap.String("metric", metric.Name))
			return nil
		}
	}

	if err := InsertDashboard(ctx, app, b.dash); err != nil {
		return err
	}

	for i, gauge := range b.gauges {
		gauge.DashID = b.dash.ID
		gauge.ProjectID = b.dash.ProjectID
		gauge.Weight = len(b.gauges) - i
	}

	for _, gauge := range b.gauges {
		if err := gauge.Validate(); err != nil {
			return err
		}
	}

	if err := InsertDashGauges(ctx, app, b.gauges); err != nil {
		return err
	}

	for i, entry := range b.entries {
		entry.DashID = b.dash.ID
		entry.ProjectID = b.dash.ProjectID
		entry.Weight = len(b.entries) - i
	}

	for _, entry := range b.entries {
		if err := entry.Validate(); err != nil {
			return err
		}
	}

	if err := InsertDashEntries(ctx, app, b.entries); err != nil {
		return err
	}

	return nil
}
