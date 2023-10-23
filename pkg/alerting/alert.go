package alerting

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/madalarm"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/org"
	"go.uber.org/zap"
)

const maxRecentAlertDuration = 8 * time.Hour

var (
	_ org.Alert = (*ErrorAlert)(nil)
	_ org.Alert = (*MetricAlert)(nil)
)

type ErrorAlert struct {
	*org.BaseAlert `bun:",inherit"`

	Params org.ErrorAlertParams `json:"params" bun:",scanonly"`
}

func NewErrorAlert() *ErrorAlert {
	return NewErrorAlertBase(org.NewBaseAlert(org.AlertError))
}

func NewErrorAlertBase(base *org.BaseAlert) *ErrorAlert {
	alert := &ErrorAlert{
		BaseAlert: base,
	}
	alert.BaseAlert.Event.Params.Any = &alert.Params
	return alert
}

func (a *ErrorAlert) Base() *org.BaseAlert {
	return a.BaseAlert
}

func (a *ErrorAlert) Summary() string {
	if a.Params.SpanCount > 1 {
		return fmt.Sprintf("The error has %d occurrences", a.Params.SpanCount)
	}
	return "A new error has just occurred"
}

type MetricAlert struct {
	*org.BaseAlert `bun:",inherit"`

	Params MetricAlertParams `json:"params" bun:",scanonly"`
}

func NewMetricAlert() *MetricAlert {
	return NewMetricAlertBase(org.NewBaseAlert(org.AlertMetric))
}

func NewMetricAlertBase(base *org.BaseAlert) *MetricAlert {
	alert := &MetricAlert{
		BaseAlert: base,
	}
	alert.BaseAlert.Event.Params.Any = &alert.Params
	return alert
}

func (a *MetricAlert) Base() *org.BaseAlert {
	return a.BaseAlert
}

func (a *MetricAlert) ShortSummary() string {
	bounds := a.Params.Bounds
	unit := a.Params.Monitor.ColumnUnit

	switch a.Params.Firing {
	case -1:
		current := bunconv.Format(a.Params.CurrentValue, unit)
		min := bunconv.Format(bounds.Min.Float64, unit)
		return fmt.Sprintf("%s is smaller than %s", current, min)
	case 1:
		current := bunconv.Format(a.Params.CurrentValue, unit)
		max := bunconv.Format(bounds.Max.Float64, unit)
		return fmt.Sprintf("%s is greater than %s", current, max)
	default:
		return ""
	}
}

func (a *MetricAlert) LongSummary(sep string) string {
	bounds := a.Params.Bounds
	unit := a.Params.Monitor.ColumnUnit
	min := formatNull(bounds.Min, unit, "-Inf")
	max := formatNull(bounds.Max, unit, "+Inf")

	var msg []string

	msg = append(msg, fmt.Sprintf(
		"You specified that value should be between %s and %s.",
		min, max,
	))

	var symbol string
	switch a.Params.Firing {
	case -1:
		symbol = "smaller"
	case 1:
		symbol = "greater"
	}

	currentValue := bunconv.Format(a.Params.CurrentValue, unit)
	currentValueVerbose := bunconv.FormatFloatVerbose(a.Params.CurrentValue)
	duration := bunconv.ShortDuration(
		time.Duration(a.Params.NumPointFiring) * time.Minute,
	)
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

type MetricAlertParams struct {
	Firing       int     `json:"firing"`
	InitialValue float64 `json:"initialValue"`
	CurrentValue float64 `json:"currentValue"`
	NormalValue  float64 `json:"normalValue"`

	NumPointFiring int             `json:"numPointFiring"`
	Bounds         madalarm.Bounds `json:"bounds"`

	Monitor struct {
		Metrics    []mql.MetricAlias `json:"metrics"`
		Query      string            `json:"query"`
		Column     string            `json:"column"`
		ColumnUnit string            `json:"columnUnit"`
	} `json:"monitor"`
}

func (params *MetricAlertParams) Update(checkRes *madalarm.CheckResult) {
	params.Firing = checkRes.Firing
	params.InitialValue = checkRes.FirstValue
	params.CurrentValue = checkRes.FirstValue
	params.NormalValue = 0
	params.NumPointFiring = checkRes.FiringFor
	params.Bounds = checkRes.Bounds
}

func selectAlertWithEvent(ctx context.Context, app *bunapp.App, eventID uint64) (org.Alert, error) {
	event := new(org.AlertEvent)

	if err := app.PG.NewSelect().
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
		user, err := org.SelectUser(ctx, app, event.UserID)
		if err != nil {
			return nil, err
		}
		event.User = user
	}

	return decodeAlert(alert)
}

func SelectAlert(ctx context.Context, app *bunapp.App, id uint64) (org.Alert, error) {
	alert, err := SelectBaseAlert(ctx, app, id)
	if err != nil {
		return nil, err
	}
	return decodeAlert(alert)
}

func SelectBaseAlert(ctx context.Context, app *bunapp.App, alertID uint64) (*org.BaseAlert, error) {
	alert := new(org.BaseAlert)
	if err := app.PG.NewSelect().
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
	switch base.Type {
	case org.AlertError:
		alert := &ErrorAlert{
			BaseAlert: base,
		}
		if err := base.Event.Params.Decode(&alert.Params); err != nil {
			return nil, err
		}
		return alert, nil
	case org.AlertMetric:
		alert := &MetricAlert{
			BaseAlert: base,
		}
		if err := base.Event.Params.Decode(&alert.Params); err != nil {
			return nil, err
		}
		return alert, nil
	default:
		return nil, fmt.Errorf("unknown alert type: %s", base.Type)
	}
}

func changeAlertStatus(
	ctx context.Context,
	app *bunapp.App,
	alert org.Alert,
	status org.AlertStatus,
	userID uint64,
) error {
	return createAlertEvent(ctx, app, alert, func(tx bun.Tx) error {
		baseAlert := alert.Base()

		event := baseAlert.Event.Clone()
		event.UserID = userID
		event.Name = org.AlertEventStatusChanged
		event.Status = status

		if err := org.InsertAlertEvent(ctx, tx, event); err != nil {
			return err
		}

		if err := updateAlertEvent(ctx, tx, baseAlert, event); err != nil {
			return err
		}

		return nil
	})
}

func updateAlertEvent(
	ctx context.Context, db bun.IDB, baseAlert *org.BaseAlert, event *org.AlertEvent,
) error {
	res, err := db.NewUpdate().
		Model(baseAlert).
		Set("event_id = ?", event.ID).
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

	baseAlert.EventID = event.ID
	baseAlert.Event = event

	return nil
}

func selectMatchingAlert(
	ctx context.Context,
	app *bunapp.App,
	alert *org.BaseAlert,
	dest org.Alert,
) error {
	return app.PG.NewSelect().
		Model(dest).
		ExcludeColumn().
		Relation("Event", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.ExcludeColumn("params")
		}).
		ColumnExpr("event.params").
		Where("a.type = ?", alert.Type).
		Where("a.attrs_hash = ?", alert.AttrsHash).
		Where("event.id IS NOT NULL").
		OrderExpr("event.created_at DESC").
		Limit(1).
		Apply(func(q *bun.SelectQuery) *bun.SelectQuery {
			if alert.MonitorID != 0 {
				q = q.Where("a.monitor_id = ?", alert.MonitorID)
			}
			if alert.TrackableModel != "" && alert.TrackableID != 0 {
				q = q.Where("a.trackable_model = ?", alert.TrackableModel).
					Where("a.trackable_id = ?", alert.TrackableID)
			}
			return q
		}).
		Scan(ctx)
}

func createAlert(ctx context.Context, app *bunapp.App, alert org.Alert) error {
	baseAlert := alert.Base()

	event := baseAlert.Event
	event.ProjectID = baseAlert.ProjectID

	switch event.Name {
	case "":
		event.Name = org.AlertEventCreated
	case org.AlertEventCreated:
		// okay
	default:
		return fmt.Errorf("unexpected event name: %q", event.Name)
	}

	return createAlertEvent(ctx, app, alert, func(tx bun.Tx) error {
		inserted, err := org.InsertAlert(ctx, app, tx, baseAlert)
		if err != nil {
			return err
		}
		if !inserted {
			return nil
		}

		event.AlertID = baseAlert.ID
		if err := org.InsertAlertEvent(ctx, tx, event); err != nil {
			return err
		}

		if err := updateAlertEvent(ctx, tx, baseAlert, event); err != nil {
			return err
		}

		return nil
	})
}

func createAlertEvent(
	ctx context.Context, app *bunapp.App, alert org.Alert, fn func(tx bun.Tx) error,
) error {
	if err := app.PG.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(tx)
	}); err != nil {
		return err
	}

	switch alert := alert.(type) {
	case *ErrorAlert:
		if err := scheduleNotifyOnErrorAlert(ctx, app, alert); err != nil {
			app.Zap(ctx).Error("scheduleNotifyOnErrorAlert failed", zap.Error(err))
		}
		return nil
	case *MetricAlert:
		if err := scheduleNotifyOnMetricAlert(ctx, app, alert); err != nil {
			app.Zap(ctx).Error("scheduleNotifyOnMetricAlert failed", zap.Error(err))
		}
		return nil
	default:
		return fmt.Errorf("unknown alert type: %T", alert)
	}
}
