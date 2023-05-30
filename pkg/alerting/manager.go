package alerting

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/cespare/xxhash"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/bunutil"
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
	Monitors []*MetricMonitor
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
	monitors, err := SelectMetricMonitors(ctx, m.app)
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

func (m *Manager) monitor(ctx context.Context, monitor *MetricMonitor, timeLT time.Time) error {
	metricMap, err := m.selectMetricMap(ctx, monitor)
	if err != nil {
		if err == sql.ErrNoRows { // one of the metrics does not exist
			return UpdateMonitorState(
				ctx, m.app, monitor.ID, monitor.State, MonitorFailed,
			)
		}
		return err
	}

	result, err := m.selectTimeseries(ctx, monitor, metricMap, timeLT)
	if err != nil {
		return err
	}

	if _, err := monitor.MadalarmOptions(); err != nil {
		return UpdateMonitorState(
			ctx, m.app, monitor.ID, monitor.State, MonitorFailed,
		)
		return err
	}

	var firing bool

	var buf []byte
	for i := range result.Timeseries {
		ts := &result.Timeseries[i]

		buf = ts.Attrs.Bytes(buf[:0])
		attrsHash := xxhash.Sum64(buf)

		alert, err := m.monitorTimeseries(ctx, monitor, ts, timeLT, attrsHash)
		if err != nil {
			return err
		}

		if alert != nil && alert.State != org.AlertClosed {
			firing = true
		}
	}

	switch {
	case len(result.Timeseries) == 0:
		if err := UpdateMonitorState(
			ctx, m.app, monitor.ID, monitor.State, MonitorNoData,
		); err != nil {
			return err
		}

	case firing && monitor.State != MonitorFiring:
		if err := UpdateMonitorState(
			ctx, m.app, monitor.ID, monitor.State, MonitorFiring,
		); err != nil {
			return err
		}

	case !firing && monitor.State != MonitorActive:
		if err := UpdateMonitorState(
			ctx, m.app, monitor.ID, monitor.State, MonitorActive,
		); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) selectMetricMap(
	ctx context.Context,
	monitor *MetricMonitor,
) (map[string]*metrics.Metric, error) {
	metricMap := make(map[string]*metrics.Metric)
	for _, ma := range monitor.Params.Metrics {
		metric, err := metrics.SelectMetricByName(ctx, m.app, monitor.ProjectID, ma.Name)
		if err != nil {
			return nil, err
		}
		metricMap[ma.Alias] = metric
	}
	return metricMap, nil
}

func (m *Manager) selectTimeseries(
	ctx context.Context,
	monitor *MetricMonitor,
	metricMap map[string]*metrics.Metric,
	timeLT time.Time,
) (*mql.Result, error) {

	storageConf := &metrics.CHStorageConfig{
		ProjectID:      monitor.ProjectID,
		MetricMap:      metricMap,
		TableName:      m.app.DistTable("measure_minutes_buffer"),
		GroupingPeriod: time.Minute,
	}
	storageConf.TimeFilter = org.TimeFilter{
		TimeGTE: timeLT.Add(-noDataMinutesThreshold * time.Minute),
		TimeLT:  timeLT,
	}

	storage := metrics.NewCHStorage(ctx, m.app.CH, storageConf)
	engine := mql.NewEngine(storage)

	query := mql.Parse(monitor.Params.Query)
	result := engine.Run(query.Parts)

	for _, part := range query.Parts {
		if part.Error.Wrapped != nil {
			return nil, part.Error.Wrapped
		}
	}

	return result, nil
}

func (m *Manager) monitorTimeseries(
	ctx context.Context,
	monitor *MetricMonitor,
	ts *mql.Timeseries,
	tm time.Time,
	attrsHash uint64,
) (*MetricAlert, error) {
	baseAlert := &org.BaseAlert{
		ProjectID: monitor.ProjectID,
		MonitorID: monitor.ID,
		Type:      org.AlertMetric,
		Attrs:     ts.Attrs.Map(),
		AttrsHash: attrsHash,
	}

	alert, err := selectRecentMetricAlert(ctx, m.app, baseAlert)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	var noDataCount int
	for i := len(ts.Value) - 1; i >= 0; i-- {
		n := ts.Value[i]
		if !math.IsNaN(n) {
			break
		}
		noDataCount++
		if noDataCount >= noDataMinutesThreshold {
			if alert != nil && alert.State == org.AlertOpen {
				if err := closeAlert(ctx, m.app, alert); err != nil {
					return nil, err
				}
			}
			return alert, nil
		}
	}

	options, err := monitor.MadalarmOptions()
	if err != nil {
		return nil, err
	}

	var index int
	for i, n := range ts.Value {
		if !math.IsNaN(n) {
			break
		}
		index = i
	}
	input := ts.Value[index:]
	timeCol := ts.Time[index:]

	checkRes, err := madalarm.Check(input, options...)
	if err != nil {
		return nil, err
	}
	recoverMinutesThreshold := int(monitor.Params.ForDuration) + 5

	if checkRes.Firing == 0 {
		if alert == nil {
			return nil, nil
		}

		// Wait until the timeseries fully recovers.
		values := lastN(checkRes.Input, recoverMinutesThreshold)
		for i := len(values) - 1; i >= 0; i-- {
			x := values[i]
			if checkRes.IsOutlier(x) != 0 {
				return alert, nil
			}
		}

		if alert.State != org.AlertClosed {
			if err := closeAlert(ctx, m.app, alert); err != nil {
				return nil, err
			}
			return alert, nil
		}

		return alert, nil
	}

	alertTime := timeCol[len(timeCol)-checkRes.FiringFor]

	if alert != nil {
		if alert.State == org.AlertOpen && time.Since(alert.UpdatedAt) < 3*time.Hour {
			return alert, nil
		}

		alert.Params.Update(checkRes)
		alert.UpdatedAt = alertTime

		if err := m.openMetricAlert(ctx, alert); err != nil {
			return nil, err
		}
		return alert, nil
	}

	baseAlert.DedupHash = monitor.ID * baseAlert.AttrsHash * timeSlot(maxRecentAlertDuration)
	baseAlert.Name = monitor.Name + ": " + ts.Name()
	baseAlert.State = org.AlertOpen
	baseAlert.CreatedAt = alertTime
	baseAlert.UpdatedAt = alertTime

	alert = &MetricAlert{
		BaseAlert: baseAlert,
	}
	alert.BaseAlert.Params.Any = &alert.Params

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

func (m *Manager) openMetricAlert(
	ctx context.Context, alert *MetricAlert,
) error {
	res, err := m.app.PG.NewUpdate().
		Model(alert).
		Set("state = ?", org.AlertOpen).
		Set("params = ?", alert.Params).
		Set("updated_at = ?", alert.UpdatedAt).
		Where("id = ?", alert.ID).
		Where("state = ?", alert.State).
		Returning("state").
		Exec(ctx)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return nil
	}

	if err := createAlertEvent(ctx, m.app, alert, &org.AlertEvent{
		ProjectID: alert.ProjectID,
		AlertID:   alert.ID,
		Name:      org.AlertEventStateChanged,
		Params:    bunutil.Params{Any: alert.Params},
	}); err != nil {
		return err
	}

	return nil
}

func lastN[T any](s []T, n int) []T {
	if n > len(s) {
		n = len(s)
	}
	return s[len(s)-n:]
}

func timeSlot(period time.Duration) uint64 {
	if period <= 0 {
		return 0
	}
	return uint64(time.Now().UnixNano()) / uint64(period)
}
