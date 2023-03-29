package alerting

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bununit"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/madalarm"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
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
	Params         org.ErrorAlertParams `json:"params"`
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
	Params         MetricAlertParams `json:"params"`
}

func (a *MetricAlert) Base() *org.BaseAlert {
	return a.BaseAlert
}

func (a *MetricAlert) Summary() string {
	switch a.Params.Firing {
	case -1:
		outlier := bununit.Format(a.Params.Outlier, a.Params.Monitor.ColumnUnit)
		min := bununit.Format(a.Params.Bounds.Min.Float64, a.Params.Monitor.ColumnUnit)
		return fmt.Sprintf("%s is less than %s", outlier, min)
	case 1:
		outlier := bununit.Format(a.Params.Outlier, a.Params.Monitor.ColumnUnit)
		max := bununit.Format(a.Params.Bounds.Max.Float64, a.Params.Monitor.ColumnUnit)
		return fmt.Sprintf("%s is greater than %s", outlier, max)
	default:
		return ""
	}
}

type MetricAlertParams struct {
	Firing  int             `json:"firing"`
	Outlier float64         `json:"outlier"`
	Minutes int             `json:"minutes"`
	Bounds  madalarm.Bounds `json:"bounds"`

	Monitor struct {
		Metrics    []upql.MetricAlias `json:"metrics"`
		Query      string             `json:"query"`
		Column     string             `json:"column"`
		ColumnUnit string             `json:"columnUnit"`
	} `json:"monitor"`
}

func (params *MetricAlertParams) Update(checkRes *madalarm.CheckResult) {
	params.Firing = checkRes.Firing
	params.Outlier = checkRes.Outlier
	params.Minutes = checkRes.FiringFor
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

func SelectBaseAlert(ctx context.Context, app *bunapp.App, id uint64) (*org.BaseAlert, error) {
	alert := new(org.BaseAlert)
	if err := app.PG.NewSelect().
		Model(alert).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return alert, nil
}

func SelectAlerts(
	ctx context.Context, app *bunapp.App, f *AlertFilter,
) ([]org.Alert, int, error) {
	baseAlerts, count, err := SelectBaseAlerts(ctx, app, f)
	if err != nil {
		return nil, 0, err
	}

	alerts := make([]org.Alert, len(baseAlerts))
	for i, baseAlert := range baseAlerts {
		alerts[i], err = decodeAlert(baseAlert)
		if err != nil {
			return nil, 0, err
		}
	}
	return alerts, count, nil
}

func SelectBaseAlerts(
	ctx context.Context, app *bunapp.App, f *AlertFilter,
) ([]*org.BaseAlert, int, error) {
	alerts := make([]*org.BaseAlert, 0)
	count, err := app.PG.NewSelect().
		Model(&alerts).
		Apply(f.WhereClause).
		Apply(f.PGOrder).
		Limit(f.Pager.GetLimit()).
		Offset(f.Pager.GetOffset()).
		ScanAndCount(ctx)
	if err != nil {
		return nil, count, err
	}
	return alerts, count, nil
}

func decodeAlert(base *org.BaseAlert) (org.Alert, error) {
	switch base.Type {
	case org.AlertError:
		alert := &ErrorAlert{
			BaseAlert: base,
		}
		if err := base.Params.Decode(&alert.Params); err != nil {
			return nil, err
		}
		return alert, nil
	case org.AlertMetric:
		alert := &MetricAlert{
			BaseAlert: base,
		}
		if err := base.Params.Decode(&alert.Params); err != nil {
			return nil, err
		}
		return alert, nil
	default:
		return nil, fmt.Errorf("unknown alert type: %s", base.Type)
	}
}

func closeAlert(ctx context.Context, app *bunapp.App, alert org.Alert) error {
	return updateAlertState(ctx, app, alert, org.AlertClosed, 0)
}

func updateAlertState(
	ctx context.Context,
	app *bunapp.App,
	alert org.Alert,
	state org.AlertState,
	userID uint64,
) error {
	baseAlert := alert.Base()
	oldState := baseAlert.State

	res, err := app.PG.NewUpdate().
		Model(baseAlert).
		Set("state = ?", state).
		Where("id = ?", baseAlert.ID).
		Where("state = ?", oldState).
		Returning("state").
		Exec(ctx)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 1 {
		if err := createAlertEvent(ctx, app, alert, &org.AlertEvent{
			UserID:    userID,
			ProjectID: baseAlert.ProjectID,
			AlertID:   baseAlert.ID,
			Name:      org.AlertEventStateChanged,
			Params: bunutil.Params{
				Any: map[string]string{
					"state":    string(baseAlert.State),
					"oldState": string(oldState),
				},
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

func selectErrorAlert(
	ctx context.Context, app *bunapp.App, alert *org.BaseAlert,
) (*ErrorAlert, error) {
	dest := new(ErrorAlert)
	if err := selectMatchingAlert(ctx, app, alert, dest, nil); err != nil {
		return nil, err
	}
	dest.BaseAlert.Params.Any = &dest.Params
	return dest, nil
}

func selectRecentMetricAlert(
	ctx context.Context, app *bunapp.App, alert *org.BaseAlert,
) (*MetricAlert, error) {
	dest := new(MetricAlert)
	if err := selectMatchingAlert(
		ctx,
		app,
		alert,
		dest,
		func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("updated_at >= ?", time.Now().Add(-maxRecentAlertDuration))
		},
	); err != nil {
		return nil, err
	}
	return dest, nil
}

func selectMatchingAlert(
	ctx context.Context,
	app *bunapp.App,
	alert *org.BaseAlert,
	dest any,
	customQuery func(*bun.SelectQuery) *bun.SelectQuery,
) error {
	return app.PG.NewSelect().
		Model(dest).
		Where("type = ?", alert.Type).
		Where("attrs_hash = ?", alert.AttrsHash).
		OrderExpr("updated_at DESC").
		Limit(1).
		Apply(func(q *bun.SelectQuery) *bun.SelectQuery {
			if alert.MonitorID != 0 {
				q = q.Where("monitor_id = ?", alert.MonitorID)
			}
			if alert.TrackableModel != "" && alert.TrackableID != 0 {
				q = q.Where("trackable_model = ?", alert.TrackableModel).
					Where("trackable_id = ?", alert.TrackableID)
			}
			return q
		}).
		Apply(customQuery).
		Scan(ctx)
}

func createAlert(ctx context.Context, app *bunapp.App, alert org.Alert) error {
	baseAlert := alert.Base()

	inserted, err := org.InsertAlert(ctx, app, baseAlert)
	if err != nil {
		return err
	}
	if !inserted {
		return nil
	}

	return createAlertEvent(ctx, app, alert, &org.AlertEvent{
		ProjectID: baseAlert.ProjectID,
		AlertID:   baseAlert.ID,
		Name:      org.AlertEventCreated,
		Params:    baseAlert.Params,
		CreatedAt: baseAlert.CreatedAt,
	})
}

func createAlertEvent(
	ctx context.Context, app *bunapp.App, alert org.Alert, alertEvent *org.AlertEvent,
) error {
	if err := org.InsertAlertEvent(ctx, app, alertEvent); err != nil {
		return err
	}

	switch alert := alert.(type) {
	case *ErrorAlert:
		if err := scheduleNotifyOnErrorAlert(ctx, app, alert, alertEvent); err != nil {
			app.Zap(ctx).Error("scheduleNotifyOnErrorAlert failed", zap.Error(err))
		}
		return nil
	case *MetricAlert:
		if err := scheduleNotifyOnMetricAlert(ctx, app, alert, alertEvent); err != nil {
			app.Zap(ctx).Error("scheduleNotifyOnMetricAlert failed", zap.Error(err))
		}
		return nil
	default:
		return fmt.Errorf("unknown alert type: %T", alert)
	}
}
