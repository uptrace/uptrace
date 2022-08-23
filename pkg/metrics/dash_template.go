package metrics

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type DashSyncer struct {
	app *bunapp.App

	httpClient *http.Client

	mu            sync.Mutex
	dashboards    []*bunconf.Dashboard
	lastUpdatedAt time.Time

	debouncerMapMu sync.Mutex
	debouncerMap   map[uint32]*bunutil.Debouncer

	logger otelzap.Logger
}

func NewDashSyncer(app *bunapp.App) *DashSyncer {
	s := &DashSyncer{
		app: app,
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
		debouncerMap: make(map[uint32]*bunutil.Debouncer),
	}

	ctx := app.Context()
	projects := app.Config().Projects
	for i := range projects {
		s.Sync(ctx, projects[i].ID)
	}

	return s
}

func (s *DashSyncer) dashboardTemplates(ctx context.Context) []*bunconf.Dashboard {
	s.mu.Lock()
	defer s.mu.Unlock()

	if time.Since(s.lastUpdatedAt) < 24*time.Hour {
		return s.dashboards
	}

	dashboards, err := s.resolveDashboardTemplates(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "resolveDashboardTemplates failed", zap.Error(err))
		return s.dashboards
	}

	s.dashboards = dashboards
	s.lastUpdatedAt = time.Now()

	return s.dashboards
}

func (s *DashSyncer) resolveDashboardTemplates(ctx context.Context) ([]*bunconf.Dashboard, error) {
	conf := s.app.Config()

	var dashboards []*bunconf.Dashboard

	for _, url := range conf.Dashboards {
		got, err := s.getDashboards(ctx, url)
		if err != nil {
			return nil, err
		}
		dashboards = append(dashboards, got...)
	}

	return dashboards, nil
}

func (s *DashSyncer) getDashboards(ctx context.Context, url string) ([]*bunconf.Dashboard, error) {
	var b []byte
	var err error

	for i := 0; i < 3; i++ {
		if i > 0 {
			time.Sleep(time.Second)
		}

		b, err = s.httpGet(ctx, url)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	var dashboards []*bunconf.Dashboard

	dec := yaml.NewDecoder(bytes.NewReader(b))
	for {
		dashboard := new(bunconf.Dashboard)
		if err := dec.Decode(&dashboard); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		dashboards = append(dashboards, dashboard)
	}

	return dashboards, nil
}

func (s *DashSyncer) httpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
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
	templates := s.dashboardTemplates(ctx)

	dashMap, err := SelectDashboardMap(ctx, s.app, projectID)
	if err != nil {
		return fmt.Errorf("SelectDashboardMap failed: %w", err)
	}

	metricMap, err := SelectMetricMap(ctx, s.app, projectID)
	if err != nil {
		return fmt.Errorf("SelectMetricMap failed: %w", err)
	}

	for _, tpl := range templates {
		if err := tpl.Validate(); err != nil {
			return err
		}

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
	entries   []*DashEntry
}

func (b *DashBuilder) Build(tpl *bunconf.Dashboard) error {
	metrics, err := upql.ParseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	b.dash.Name = tpl.Name
	b.dash.Metrics = metrics
	b.dash.Query = strings.Join(tpl.Query, " | ")
	b.dash.Columns = tpl.Columns
	b.dash.IsTable = len(b.dash.Metrics) > 0 && b.dash.Query != ""

	for _, entry := range tpl.Entries {
		if err := b.entry(entry); err != nil {
			return err
		}
	}

	if err := b.dash.Validate(); err != nil {
		return err
	}
	for _, entry := range b.entries {
		if err := entry.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (b *DashBuilder) entry(tpl *bunconf.DashEntry) error {
	metrics, err := upql.ParseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		if _, ok := b.metricMap[metric.Name]; !ok {
			return nil
		}
	}

	b.entries = append(b.entries, &DashEntry{
		Name:    tpl.Name,
		Metrics: metrics,
		Query:   tpl.Query,
		Columns: tpl.Columns,
	})
	return nil
}

func (b *DashBuilder) Save(ctx context.Context, app *bunapp.App) error {
	for _, metric := range b.dash.Metrics {
		if _, ok := b.metricMap[metric.Name]; !ok {
			return nil
		}
	}

	if err := InsertDashboard(ctx, app, b.dash); err != nil {
		return err
	}

	for i, entry := range b.entries {
		entry.DashID = b.dash.ID
		entry.ProjectID = b.dash.ProjectID
		entry.Weight = len(b.entries) - i
	}

	if err := InsertDashEntries(ctx, app, b.entries); err != nil {
		return err
	}

	return nil
}
