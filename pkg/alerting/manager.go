package alerting

import (
	"context"
	"database/sql"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/vmihailenco/taskq/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/madalarm"
	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/unixtime"
)

const noDataMinutesThreshold = 60

type ManagerParams struct {
	fx.In

	Logger    *otelzap.Logger
	PG        *bun.DB
	CH        *ch.DB
	MainQueue taskq.Queue
}

type Manager struct {
	//conf *ManagerConfig
	*ManagerParams

	stop func()
	exit <-chan struct{}
	done <-chan struct{}
}

type ManagerConfig struct {
	//Monitors []*org.MetricMonitor
	Logger *zap.Logger
}

// func NewManager(app *bunapp.App, conf *ManagerConfig) *Manager {
func NewManager(p ManagerParams) *Manager {
	return &Manager{ManagerParams: &p}
}

func runManager(lc fx.Lifecycle, man *Manager) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go man.Run()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			man.Stop()
			return nil
		},
	})
}

func (m *Manager) Stop() {
	m.stop()
	<-m.done
}

func (m *Manager) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	m.stop = cancel
	m.exit = ctx.Done()

	done := make(chan struct{})
	m.done = done
	defer close(done)

	nextCheck := time.Now().
		Add(time.Minute).
		Truncate(time.Minute).
		Add(30 * time.Second)

	for {
		select {
		case <-time.After(time.Until(nextCheck)):
		case <-m.exit:
			return
		}

		m.tick(ctx, nextCheck.Truncate(time.Minute))
		nextCheck = nextCheck.Add(time.Minute)
	}
}

func (m *Manager) tick(ctx context.Context, tm time.Time) {
	monitors, err := m.selectMetricMonitors(ctx)
	if err != nil {
		m.Logger.Error("SelectMonitors failed", zap.Error(err))
		return
	}

	for _, monitor := range monitors {
		if err := bunotel.RunWithNewRoot(ctx, "alerting.monitor", func(ctx context.Context) error {
			span := trace.SpanFromContext(ctx)

			span.SetAttributes(
				attribute.String("monitor.name", monitor.Name),
				attribute.String("monitor.query", monitor.Params.Query),
			)

			return m.monitor(ctx, monitor, tm)
		}); err != nil {
			m.Logger.Error("monitor failed",
				zap.Error(err),
				zap.String("monitor.name", monitor.Name),
				zap.String("monitor.query", monitor.Params.Query))
		}
	}
}

func (m *Manager) selectMetricMonitors(ctx context.Context) ([]*org.MetricMonitor, error) {
	monitors := make([]*org.MetricMonitor, 0)
	if err := m.PG.NewSelect().
		Model(&monitors).
		Where("type = ?", org.MonitorMetric).
		Where("state NOT IN (?)", org.MonitorPaused, org.MonitorFailed).
		Scan(ctx); err != nil {
		return nil, err
	}
	return monitors, nil
}

func (m *Manager) monitor(ctx context.Context, monitor *org.MetricMonitor, timeLT time.Time) error {
	metricMap, err := m.selectMetricMap(ctx, monitor)
	if err != nil {
		if err == sql.ErrNoRows { // one of the metrics does not exist
			return org.UpdateMonitorState(
				ctx, m.PG, monitor.ID, monitor.State, org.MonitorFailed,
			)
		}
		return err
	}

	result, err := m.selectTimeseries(ctx, monitor, metricMap, timeLT)
	if err != nil {
		return err
	}

	options, err := monitor.MadalarmOptions()
	if err != nil {
		return err
	}

	checker, err := madalarm.NewChecker(options...)
	if err != nil {
		return err
	}

	var firing bool

	var buf []byte
	for _, ts := range result.Timeseries {
		buf = ts.Attrs.Bytes(buf[:0], nil)
		attrsHash := xxhash.Sum64(buf)

		alert, err := m.checkTimeseries(ctx, monitor, checker, ts, attrsHash)
		if err != nil {
			return err
		}

		if alert != nil && alert.Event.Status != org.AlertStatusClosed {
			firing = true
		}
	}

	switch {
	case len(result.Timeseries) == 0:
		if err := org.UpdateMonitorState(
			ctx, m.PG, monitor.ID, monitor.State, org.MonitorNoData,
		); err != nil {
			return err
		}

	case firing && monitor.State != org.MonitorFiring:
		if err := org.UpdateMonitorState(
			ctx, m.PG, monitor.ID, monitor.State, org.MonitorFiring,
		); err != nil {
			return err
		}

	case !firing && monitor.State != org.MonitorActive:
		if err := org.UpdateMonitorState(
			ctx, m.PG, monitor.ID, monitor.State, org.MonitorActive,
		); err != nil {
			return err
		}
	}

	if _, err := m.PG.NewUpdate().
		Model(monitor).
		Set("updated_at = now()").
		Where("id = ?", monitor.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (m *Manager) selectMetricMap(
	ctx context.Context,
	monitor *org.MetricMonitor,
) (map[string]*metrics.Metric, error) {
	metricMap := make(map[string]*metrics.Metric)
	for _, ma := range monitor.Params.Metrics {
		metric, err := metrics.SelectMetricByName(ctx, m.PG, monitor.ProjectID, ma.Name)
		if err != nil {
			return nil, err
		}
		metricMap["$"+ma.Alias] = metric
	}
	return metricMap, nil
}

func (m *Manager) selectTimeseries(
	ctx context.Context,
	monitor *org.MetricMonitor,
	metricMap map[string]*metrics.Metric,
	timeLT time.Time,
) (*mql.Result, error) {
	storage := metrics.NewCHStorage(ctx, m.CH, &metrics.CHStorageConfig{
		ProjectID: monitor.ProjectID,
		MetricMap: metricMap,
		TableName: "datapoint_minutes",
	})
	engine := mql.NewEngine(
		storage,
		unixtime.ToSeconds(timeLT.Add(-noDataMinutesThreshold*time.Minute)),
		unixtime.ToSeconds(timeLT),
		time.Minute,
	)

	query := mql.ParseQuery(monitor.Params.Query)
	result := engine.Run(query.Parts)

	for _, part := range query.Parts {
		if part.Error.Wrapped != nil {
			return nil, part.Error.Wrapped
		}
	}

	return result, nil
}

func (m *Manager) checkTimeseries(
	ctx context.Context,
	monitor *org.MetricMonitor,
	checker *madalarm.Checker,
	ts *mql.Timeseries,
	attrsHash uint64,
) (*MetricAlert, error) {
	baseAlert := &org.BaseAlert{
		ProjectID: monitor.ProjectID,
		MonitorID: monitor.ID,
		Type:      org.AlertMetric,
		Attrs:     ts.Attrs.Map(),
		AttrsHash: attrsHash,
	}

	alert, err := m.selectMetricAlert(ctx, baseAlert)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if alert != nil && alert.Event.Status == org.AlertStatusOpen {
		return m.checkOpenAlert(ctx, monitor, checker, ts, alert)
	}

	ts.TrimNaNLeft()
	checkRes, err := checker.Check(ts.Value, nil)
	if err != nil {
		return nil, err
	}

	if checkRes.Firing == 0 {
		return nil, nil
	}

	alertTime := ts.Time[len(ts.Time)-checkRes.FiringFor].Time()

	if alert != nil && alertTime.Sub(alert.Event.CreatedAt) < 8*time.Hour {
		// There is already a recent closed alert. Reuse it instead of creating a new one.
		if err := tryAlertInTx(ctx, m.Logger, m.PG, m.CH, m.MainQueue, alert, func(tx bun.Tx) error {
			event := alert.GetEvent().Clone().(*MetricAlertEvent)
			event.Params.Update(monitor, checkRes)

			baseEvent := event.Base()
			baseEvent.Name = org.AlertEventStatusChanged
			baseEvent.Status = org.AlertStatusOpen
			baseEvent.Time = alertTime
			baseEvent.CreatedAt = alertTime

			if err := org.InsertAlertEvent(ctx, tx, event); err != nil {
				return err
			}
			if err := updateAlertEvent(ctx, tx, alert, event); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return nil, err
		}
		return alert, nil
	}

	baseAlert.Name = monitor.Name + ": " + ts.Name()
	baseAlert.CreatedAt = alertTime

	alert = &MetricAlert{
		BaseAlert: *baseAlert,
		Event:     new(MetricAlertEvent),
	}
	alert.Event.Status = org.AlertStatusOpen
	alert.Event.Time = alertTime

	params := &alert.Event.Params
	params.WhereQuery = ts.WhereQuery()
	params.Update(monitor, checkRes)

	if err := createAlert(ctx, m.Logger, m.PG, m.CH, m.MainQueue, alert); err != nil {
		return nil, err
	}
	return alert, nil
}

func (m *Manager) selectMetricAlert(
	ctx context.Context, alert *org.BaseAlert,
) (*MetricAlert, error) {
	dest := NewMetricAlert()
	if err := selectMatchingAlert(ctx, m.PG, alert, dest); err != nil {
		return nil, err
	}
	return dest, nil
}

func (m *Manager) checkOpenAlert(
	ctx context.Context,
	monitor *org.MetricMonitor,
	checker *madalarm.Checker,
	ts *mql.Timeseries,
	alert *MetricAlert,
) (*MetricAlert, error) {
	checkNumPoint := monitor.Params.CheckNumPoint
	bounds := alert.Event.Params.Bounds

	checkRes, err := checker.Check(ts.Value, &bounds)
	if err != nil {
		return nil, err
	}

	if checkRes.Firing == 0 && checkRes.FiringFor == 0 {
		if err := tryAlertInTx(ctx, m.Logger, m.PG, m.CH, m.MainQueue, alert, func(tx bun.Tx) error {
			event := alert.Event.Clone().(*MetricAlertEvent)
			event.Params.NormalValue = ts.Value[len(ts.Value)-checkNumPoint]
			event.Params.UpdateMonitor(monitor)

			baseEvent := event.Base()
			baseEvent.Name = org.AlertEventStatusChanged
			baseEvent.Status = org.AlertStatusClosed
			baseEvent.CreatedAt = ts.Time[len(ts.Time)-checkNumPoint].Time()

			if err := org.InsertAlertEvent(ctx, tx, event); err != nil {
				return err
			}
			if err := updateAlertEvent(ctx, tx, alert, event); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return alert, nil
	}

	if checkRes.FiringFor < checkNumPoint {
		return alert, nil
	}

	if alertTime := ts.Time[len(ts.Time)-1].Time(); alertTime.Sub(alert.Event.CreatedAt) >= time.Hour {
		if err := tryAlertInTx(ctx, m.Logger, m.PG, m.CH, m.MainQueue, alert, func(tx bun.Tx) error {
			event := alert.Event.Clone().(*MetricAlertEvent)
			event.Params.Firing = checkRes.Firing
			event.Params.CurrentValue = ts.Value[len(ts.Value)-1]
			event.Params.UpdateMonitor(monitor)

			baseEvent := event.Base()
			baseEvent.Name = org.AlertEventRecurring
			baseEvent.CreatedAt = alertTime

			if err := org.InsertAlertEvent(ctx, tx, event); err != nil {
				return err
			}
			if err := updateAlertEvent(ctx, tx, alert, event); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return nil, err
		}
	}

	return alert, nil
}
