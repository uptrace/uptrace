package alerting

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/cespare/xxhash"
	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/madalarm"
	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/org"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const noDataMinutesThreshold = 60

type Manager struct {
	app  *bunapp.App
	conf *ManagerConfig

	stop func()
	exit <-chan struct{}
	done <-chan struct{}
}

type ManagerConfig struct {
	Monitors []*org.MetricMonitor
	Logger   *zap.Logger
}

func NewManager(app *bunapp.App, conf *ManagerConfig) *Manager {
	return &Manager{
		app:  app,
		conf: conf,
	}
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
	defer func() {
		close(done)
	}()

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
		m.conf.Logger.Error("SelectMonitors failed",
			zap.Error(err))
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
			m.conf.Logger.Error("monitor failed",
				zap.Error(err),
				zap.String("monitor.name", monitor.Name),
				zap.String("monitor.query", monitor.Params.Query))
		}
	}
}

func (m *Manager) selectMetricMonitors(ctx context.Context) ([]*org.MetricMonitor, error) {
	monitors := make([]*org.MetricMonitor, 0)
	if err := m.app.PG.NewSelect().
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
				ctx, m.app, monitor.ID, monitor.State, org.MonitorFailed,
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
	for i := range result.Timeseries {
		ts := &result.Timeseries[i]

		buf = ts.Attrs.Bytes(buf[:0])
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
			ctx, m.app, monitor.ID, monitor.State, org.MonitorNoData,
		); err != nil {
			return err
		}

	case firing && monitor.State != org.MonitorFiring:
		if err := org.UpdateMonitorState(
			ctx, m.app, monitor.ID, monitor.State, org.MonitorFiring,
		); err != nil {
			return err
		}

	case !firing && monitor.State != org.MonitorActive:
		if err := org.UpdateMonitorState(
			ctx, m.app, monitor.ID, monitor.State, org.MonitorActive,
		); err != nil {
			return err
		}
	}

	if _, err := m.app.PG.NewUpdate().
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
		metric, err := metrics.SelectMetricByName(ctx, m.app, monitor.ProjectID, ma.Name)
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

	storageConf := &metrics.CHStorageConfig{
		ProjectID:        monitor.ProjectID,
		MetricMap:        metricMap,
		TableName:        "datapoint_minutes",
		GroupingInterval: time.Minute,
	}
	storageConf.TimeFilter = org.TimeFilter{
		TimeGTE: timeLT.Add(-noDataMinutesThreshold * time.Minute),
		TimeLT:  timeLT,
	}

	storage := metrics.NewCHStorage(ctx, m.app.CH, storageConf)
	engine := mql.NewEngine(storage)

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

	var index int
	for i, n := range ts.Value {
		if !math.IsNaN(n) {
			break
		}
		index = i
	}
	input := ts.Value[index:]

	checkRes, err := checker.Check(input, nil)
	if err != nil {
		return nil, err
	}

	if checkRes.Firing == 0 {
		return nil, nil
	}

	alertTime := ts.Time[len(ts.Time)-checkRes.FiringFor]

	if alert != nil && alertTime.Sub(alert.Event.CreatedAt) < 8*time.Hour {
		// There is already a recent closed alert. Reuse it instead of creating one.
		alert.Params.Update(checkRes)
		if err := m.reopenAlert(ctx, alert, alertTime); err != nil {
			return nil, err
		}
		return alert, nil
	}

	baseAlert.Name = monitor.Name + ": " + ts.Name()
	baseAlert.Event = &org.AlertEvent{
		Status: org.AlertStatusOpen,
		Time:   alertTime,
	}
	baseAlert.CreatedAt = alertTime

	alert = NewMetricAlertBase(baseAlert)
	params := &alert.Params
	params.Update(checkRes)

	params.Monitor.Metrics = monitor.Params.Metrics
	params.Monitor.Query = monitor.Params.Query
	if len(ts.Attrs) > 0 {
		params.Monitor.Query += " | " + ts.WhereQuery()
	}
	params.Monitor.Column = monitor.Params.Column
	params.Monitor.ColumnUnit = monitor.Params.ColumnUnit

	if err := createAlert(ctx, m.app, alert); err != nil {
		return nil, err
	}

	return alert, nil
}

func (m *Manager) selectMetricAlert(
	ctx context.Context, alert *org.BaseAlert,
) (*MetricAlert, error) {
	dest := NewMetricAlert()
	if err := selectMatchingAlert(ctx, m.app, alert, dest); err != nil {
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
) (*MetricAlert, error,
) {
	checkNumPoint := monitor.Params.CheckNumPoint
	bounds := checker.Bounds()

	checkRes, err := checker.Check(ts.Value, bounds)
	if err != nil {
		return nil, err
	}

	if checkRes.Firing == 0 && checkRes.FiringFor == 0 {
		alert.Params.NormalValue = ts.Value[len(ts.Value)-checkNumPoint]
		tm := ts.Time[len(ts.Time)-checkNumPoint]
		if err := m.closeAlert(ctx, alert, tm); err != nil {
			return nil, err
		}
		return alert, nil
	}

	tm := ts.Time[len(ts.Time)-1]
	if checkRes.FiringFor == checkNumPoint && tm.Sub(alert.Event.CreatedAt) >= time.Hour {
		// Remind only when the timeseries is fully firing.
		// Otherwise, chances are we will remind about a timeseries that is about to recover.
		alert.Params.CurrentValue = ts.Value[len(ts.Value)-1]
		if err := m.remindAboutAlert(ctx, alert, tm); err != nil {
			return nil, err
		}
	}
	return alert, nil
}

func (m *Manager) remindAboutAlert(
	ctx context.Context, alert *MetricAlert, alertTime time.Time,
) error {
	return createAlertEvent(ctx, m.app, alert, func(tx bun.Tx) error {
		event := alert.Event.Clone()
		event.Name = org.AlertEventRecurring
		event.CreatedAt = alertTime

		if err := org.InsertAlertEvent(ctx, tx, event); err != nil {
			return err
		}

		if err := updateAlertEvent(ctx, tx, alert.Base(), event); err != nil {
			return err
		}

		return nil
	})
}

func (m *Manager) closeAlert(
	ctx context.Context, alert *MetricAlert, alertTime time.Time,
) error {
	// Keep the time as is.
	alert.Event.CreatedAt = alertTime

	return changeAlertStatus(
		ctx,
		m.app,
		alert,
		org.AlertStatusClosed,
		0,
	)
}

func (m *Manager) reopenAlert(
	ctx context.Context, alert *MetricAlert, alertTime time.Time,
) error {
	alert.Event.Time = alertTime
	alert.Event.CreatedAt = alertTime

	return changeAlertStatus(
		ctx,
		m.app,
		alert,
		org.AlertStatusOpen,
		0,
	)
}
