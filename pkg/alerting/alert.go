package alerting

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vmihailenco/taskq/v4"
	"github.com/xhit/go-str2duration/v2"
	"go.uber.org/zap"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/madalarm"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/unixtime"
)

var (
	_ org.Alert = (*ErrorAlert)(nil)
	_ org.Alert = (*MetricAlert)(nil)
)

var (
	_ org.AlertEvent = (*ErrorAlertEvent)(nil)
	_ org.AlertEvent = (*MetricAlertEvent)(nil)
)

type ErrorAlert struct {
	org.BaseAlert `bun:",inherit"`

	Event *ErrorAlertEvent `json:"-" bun:"rel:belongs-to,join:event_id=id"`
}

func NewErrorAlert() *ErrorAlert {
	return &ErrorAlert{
		Event: new(ErrorAlertEvent),
	}
}
func (a *ErrorAlert) Base() *org.BaseAlert {
	return &a.BaseAlert
}

func (a *ErrorAlert) GetEvent() org.AlertEvent {
	return a.Event
}

func (a *ErrorAlert) SetEvent(event org.AlertEvent) {
	a.EventID = event.Base().ID
	a.Event = event.(*ErrorAlertEvent)
}

type ErrorAlertEvent struct {
	org.BaseAlertEvent `bun:",inherit"`

	Params org.ErrorAlertParams `json:"params" bun:"type:jsonb,nullzero"`
}

func (e *ErrorAlertEvent) Base() *org.BaseAlertEvent {
	return &e.BaseAlertEvent
}

func (e *ErrorAlertEvent) Clone() org.AlertEvent {
	clone := *e
	clone.ID = 0
	clone.UserID = 0
	return &clone
}

type MetricAlert struct {
	org.BaseAlert `bun:",inherit"`

	Event *MetricAlertEvent `json:"-" bun:"rel:belongs-to,join:event_id=id"`
}

func NewMetricAlert() *MetricAlert {
	return &MetricAlert{
		Event: new(MetricAlertEvent),
	}
}

func (a *MetricAlert) Base() *org.BaseAlert {
	return &a.BaseAlert
}

func (a *MetricAlert) GetEvent() org.AlertEvent {
	return a.Event
}

func (a *MetricAlert) SetEvent(event org.AlertEvent) {
	a.EventID = event.Base().ID
	a.Event = event.(*MetricAlertEvent)
}

func (a *MetricAlert) ShortSummary() string {
	bounds := a.Event.Params.Bounds
	unit := a.Event.Params.Monitor.ColumnUnit

	switch a.Event.Params.Firing {
	case -1:
		current := bunconv.Format(a.Event.Params.CurrentValue, unit)
		min := bunconv.Format(bounds.Min.Float64, unit)
		return fmt.Sprintf("%s is smaller than %s", current, min)
	case 1:
		current := bunconv.Format(a.Event.Params.CurrentValue, unit)
		max := bunconv.Format(bounds.Max.Float64, unit)
		return fmt.Sprintf("%s is greater than %s", current, max)
	default:
		return ""
	}
}

func (a *MetricAlert) LongSummary(sep string) string {
	bounds := a.Event.Params.Bounds
	unit := a.Event.Params.Monitor.ColumnUnit
	min := formatNull(bounds.Min, unit, "-Inf")
	max := formatNull(bounds.Max, unit, "+Inf")

	var msg []string
	msg = append(msg, fmt.Sprintf(
		"You specified that value should be between %s and %s.",
		min, max,
	))

	var symbol string
	switch a.Event.Params.Firing {
	case -1:
		symbol = "smaller"
	case 1:
		symbol = "greater"
	}

	currentValue := bunconv.Format(a.Event.Params.CurrentValue, unit)
	currentValueVerbose := bunconv.FormatFloatVerbose(a.Event.Params.CurrentValue)
	groupingInterval := a.Event.Params.Monitor.GroupingInterval.Duration()
	duration := str2duration.String(time.Duration(a.Event.Params.NumPointFiring) * groupingInterval)
	msg = append(msg, fmt.Sprintf(
		"The actual value of %s (%s) has been %s than this range for at least %s.",
		currentValue, currentValueVerbose, symbol, duration,
	))

	return strings.Join(msg, sep)
}

func formatNull(v bunutil.NullFloat64, unit string, nullValue string) string {
	if !v.Valid {
		return nullValue
	}
	return bunconv.Format(v.Float64, unit)
}

type MetricAlertEvent struct {
	org.BaseAlertEvent `bun:",inherit"`

	Params MetricAlertParams `json:"params" bun:"type:jsonb,nullzero"`
}

func (e *MetricAlertEvent) Base() *org.BaseAlertEvent {
	return &e.BaseAlertEvent
}

func (e *MetricAlertEvent) Clone() org.AlertEvent {
	clone := *e
	clone.ID = 0
	clone.UserID = 0
	return &clone
}

type MetricAlertParams struct {
	Firing       int     `json:"firing"`
	InitialValue float64 `json:"initialValue"`
	CurrentValue float64 `json:"currentValue"`
	NormalValue  float64 `json:"normalValue"`

	NumPointFiring int    `json:"numPointFiring"`
	WhereQuery     string `json:"whereQuery"`

	Bounds madalarm.Bounds `json:"bounds"`

	Monitor struct {
		Metrics          []mql.MetricAlias `json:"metrics"`
		Query            string            `json:"query"`
		Column           string            `json:"column"`
		ColumnUnit       string            `json:"columnUnit"`
		GroupingInterval unixtime.Millis   `json:"groupingInterval"`
	} `json:"monitor"`
}

func (params *MetricAlertParams) Update(
	monitor *org.MetricMonitor, checkRes *madalarm.CheckResult,
) {
	params.UpdateMonitor(monitor)
	params.UpdateCheckResult(checkRes)
}

func (params *MetricAlertParams) UpdateMonitor(monitor *org.MetricMonitor) {
	params.Monitor.Metrics = monitor.Params.Metrics
	params.Monitor.Query = monitor.Params.Query
	params.Monitor.Column = monitor.Params.Column
	params.Monitor.ColumnUnit = monitor.Params.ColumnUnit
	//params.Monitor.GroupingInterval = monitor.Params.GroupingInterval
}

func (params *MetricAlertParams) UpdateCheckResult(checkRes *madalarm.CheckResult) {
	params.Firing = checkRes.Firing
	params.InitialValue = checkRes.FirstValue
	params.CurrentValue = checkRes.FirstValue
	params.NormalValue = 0
	params.NumPointFiring = checkRes.FiringFor
	params.Bounds = checkRes.Bounds
}

func selectAlertWithEvent(
	ctx context.Context,
	pg *bun.DB,
	users *org.UserGateway,
	eventID uint64,
) (org.Alert, error) {
	event := new(org.BaseAlertEvent)

	if err := pg.NewSelect().
		Model(event).
		Relation("Alert").
		Where("e.id = ?", eventID).
		Where("alert.id IS NOT NULL").
		Scan(ctx); err != nil {
		return nil, err
	}

	alert := event.Alert
	alert.Event = event

	if event.UserID != 0 {
		user, err := users.SelectByID(ctx, event.UserID)
		if err != nil {
			return nil, err
		}
		event.User = user
	}

	return decodeAlert(alert)
}

func SelectAlert(ctx context.Context, pg *bun.DB, id uint64) (org.Alert, error) {
	alert, err := SelectBaseAlert(ctx, pg, id)
	if err != nil {
		return nil, err
	}
	return decodeAlert(alert)
}

func SelectBaseAlert(ctx context.Context, pg *bun.DB, alertID uint64) (*org.BaseAlert, error) {
	alert := new(org.BaseAlert)
	if err := pg.NewSelect().
		Model(alert).
		Relation("Event").
		Where("a.id = ?", alertID).
		Where("event.id IS NOT NULL").
		Scan(ctx); err != nil {
		return nil, err
	}
	return alert, nil
}

func decodeAlert(base *org.BaseAlert) (org.Alert, error) {
	if base.Event == nil {
		return nil, errors.New("alert event can't be nil")
	}

	switch base.Type {
	case org.AlertError:
		alert := NewErrorAlert()
		alert.BaseAlert = *base
		alert.Event.BaseAlertEvent = *base.Event
		if err := base.Event.Params.Decode(&alert.Event.Params); err != nil {
			return nil, err
		}
		return alert, nil
	case org.AlertMetric:
		alert := NewMetricAlert()
		alert.BaseAlert = *base
		alert.Event.BaseAlertEvent = *base.Event
		if err := base.Event.Params.Decode(&alert.Event.Params); err != nil {
			return nil, err
		}
		return alert, nil
	default:
		return nil, fmt.Errorf("unsupported alert type: %s", base.Type)
	}
}

func updateAlertEvent(
	ctx context.Context, db bun.IDB, alert org.Alert, event org.AlertEvent,
) error {
	baseAlert := alert.Base()
	baseEvent := event.Base()

	if baseEvent.ID == 0 {
		return errors.New("event id can't be zero")
	}

	res, err := db.NewUpdate().
		Model(alert).
		Set("name = ?", baseAlert.Name).
		Set("event_id = ?", baseEvent.ID).
		Where("id = ?", baseAlert.ID).
		Apply(func(q *bun.UpdateQuery) *bun.UpdateQuery {
			if baseAlert.EventID != 0 {
				return q.Where("event_id = ?", baseAlert.EventID)
			}
			return q.Where("event_id IS NULL")
		}).
		Exec(ctx)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return errors.New("transaction conflict")
	}

	alert.SetEvent(event)
	return nil
}

func selectMatchingAlert(
	ctx context.Context,
	pg *bun.DB,
	alert *org.BaseAlert,
	dest org.Alert,
) error {
	return pg.NewSelect().
		Model(dest).
		ExcludeColumn().
		Relation("Event").
		Where("a.type = ?", alert.Type).
		Where("a.attrs_hash = ?", alert.AttrsHash).
		Where("event.id IS NOT NULL").
		OrderExpr("event.created_at DESC").
		Limit(1).
		Apply(func(q *bun.SelectQuery) *bun.SelectQuery {
			if alert.SpanGroupID != 0 {
				q = q.Where("a.span_group_id = ?", alert.SpanGroupID)
			}
			if alert.MonitorID != 0 {
				q = q.Where("a.monitor_id = ?", alert.MonitorID)
			}
			return q
		}).
		Scan(ctx)
}

func createAlert(
	ctx context.Context,
	logger *otelzap.Logger,
	pg *bun.DB,
	ch *ch.DB,
	mainQueue taskq.Queue,
	alert org.Alert,
) error {
	baseAlert := alert.Base()

	event := alert.GetEvent()
	baseEvent := event.Base()
	baseEvent.ProjectID = baseAlert.ProjectID

	switch baseEvent.Name {
	case "":
		baseEvent.Name = org.AlertEventCreated
	case org.AlertEventCreated:
		// okay
	default:
		return fmt.Errorf("unexpected event name: %q", baseEvent.Name)
	}

	return tryAlertInTx(ctx, logger, pg, ch, mainQueue, alert, func(tx bun.Tx) error {
		inserted, err := org.InsertAlert(ctx, tx, alert)
		if err != nil {
			return err
		}
		if !inserted {
			return errors.New("transaction conflict")
		}

		baseEvent.AlertID = baseAlert.ID
		if err := org.InsertAlertEvent(ctx, tx, event); err != nil {
			return err
		}

		if err := updateAlertEvent(ctx, tx, alert, event); err != nil {
			return err
		}

		return nil
	})
}

func tryAlertInTx(
	ctx context.Context,
	logger *otelzap.Logger,
	pg *bun.DB,
	ch *ch.DB,
	mainQueue taskq.Queue,
	alert org.Alert,
	fn func(tx bun.Tx) error,
) error {
	if err := pg.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(tx)
	}); err != nil {
		return err
	}

	if alert.Muted() {
		return nil
	}

	baseAlert := alert.Base()
	if baseAlert.EventID == 0 {
		return errors.New("alert's event id can't be zero")
	}

	switch alert := alert.(type) {
	case *ErrorAlert:
		if err := scheduleNotifyOnErrorAlert(ctx, logger, pg, ch, mainQueue, alert); err != nil {
			logger.Error("scheduleNotifyOnErrorAlert failed", zap.Error(err))
		}
		return nil
	case *MetricAlert:
		if err := scheduleNotifyOnMetricAlert(ctx, pg, mainQueue, alert); err != nil {
			logger.Error("scheduleNotifyOnMetricAlert failed", zap.Error(err))
		}
		return nil
	default:
		return fmt.Errorf("unsupported alert type: %T", alert)
	}
}
