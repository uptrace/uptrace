package alerting

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing"
	"go.uber.org/zap"
)

func scheduleNotifyOnErrorAlert(
	ctx context.Context, app *bunapp.App, alert *ErrorAlert,
) error {
	span, err := tracing.SelectSpan(
		ctx, app, alert.ProjectID, alert.Event.Params.TraceID, alert.Event.Params.SpanID,
	)
	if err != nil {
		return err
	}

	monitors, err := selectErrorMonitors(ctx, app, alert, span)
	if err != nil {
		return err
	}

	if len(monitors) == 0 {
		return nil
	}

	if err := scheduleNotifyByEmailOnErrorAlert(ctx, app, alert, monitors); err != nil {
		app.Zap(ctx).Error("scheduleNotifyByEmailOnErrorAlert failed", zap.Error(err))
	}

	if err := scheduleNotifyByChannelsOnErrorAlert(ctx, app, alert, monitors); err != nil {
		app.Zap(ctx).Error("scheduleNotifyByChannelsOnErrorAlert failed", zap.Error(err))
	}

	return nil
}

func selectErrorMonitors(
	ctx context.Context, app *bunapp.App, alert *ErrorAlert, span *tracing.Span,
) ([]*org.ErrorMonitor, error) {
	var monitors []*org.ErrorMonitor

	q := app.PG.NewSelect().
		Model(&monitors).
		Where("project_id = ?", alert.ProjectID).
		Where("type = ?", org.MonitorError).
		Where("state = ?", org.MonitorActive).
		Limit(100)

	if alert.Event.Params.SpanCount > 0 {
		q = q.Where("(params->>'notifyOnRecurringErrors')::boolean")
	} else {
		q = q.Where("(params->>'notifyOnNewErrors')::boolean")
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	for i := len(monitors) - 1; i >= 0; i-- {
		monitor := monitors[i]
		if !monitorMatches(monitor, span) {
			monitors = append(monitors[:i], monitors[i+1:]...)
		}
	}

	return monitors, nil
}

func monitorMatches(m *org.ErrorMonitor, span *tracing.Span) bool {
	for i := range m.Params.Matchers {
		if !m.Params.Matchers[i].Matches(span.Attrs) {
			return false
		}
	}
	return true
}

func scheduleNotifyByEmailOnErrorAlert(
	ctx context.Context,
	app *bunapp.App,
	alert *ErrorAlert,
	monitors []*org.ErrorMonitor,
) error {
	var recipients []string

	seenEmails := make(map[string]bool)
	for _, monitor := range monitors {
		emails, err := selectEmailRecipientsForMonitor(
			ctx,
			app,
			monitor.Base(),
			func(q *bun.SelectQuery) *bun.SelectQuery {
				if alert.Event.Params.SpanCount > 0 {
					return q.Where("up IS NULL OR up.notify_on_recurring_errors")
				}
				return q.Where("up IS NULL OR up.notify_on_new_errors")
			},
		)
		if err != nil {
			return err
		}

		for _, email := range emails {
			if seenEmails[email] {
				continue
			}
			seenEmails[email] = true
			recipients = append(recipients, email)
		}
	}

	if len(recipients) > 0 {
		job := NotifyByEmailTask.NewJob(alert.EventID, recipients)
		if err := app.MainQueue.AddJob(ctx, job); err != nil {
			return err
		}
	}

	return nil
}

func scheduleNotifyByChannelsOnErrorAlert(
	ctx context.Context,
	app *bunapp.App,
	alert *ErrorAlert,
	monitors []*org.ErrorMonitor,
) error {
	monitorIDs := make([]uint64, len(monitors))
	for i, monitor := range monitors {
		monitorIDs[i] = monitor.ID
	}

	var channels []*BaseNotifChannel

	if err := app.PG.NewSelect().
		Model(&channels).
		Where("project_id = ?", alert.ProjectID).
		Where("state = ?", NotifChannelDelivering).
		Limit(100).
		Apply(func(q *bun.SelectQuery) *bun.SelectQuery {
			subq := app.PG.NewSelect().
				Model((*org.MonitorChannel)(nil)).
				ColumnExpr("channel_id").
				Where("monitor_id IN (?)", bun.In(monitorIDs))

			return q.Where("id IN (?)", subq)
		}).
		Scan(ctx); err != nil {
		return err
	}

	var firstErr error

	for _, channel := range channels {
		switch channel.Type {
		case NotifChannelSlack:
			job := NotifyBySlackTask.NewJob(alert.EventID, channel.ID)
			if err := app.MainQueue.AddJob(ctx, job); err != nil && firstErr == nil {
				firstErr = err
			}
		case NotifChannelTelegram:
			job := NotifyByTelegramTask.NewJob(alert.EventID, channel.ID)
			if err := app.MainQueue.AddJob(ctx, job); err != nil && firstErr == nil {
				firstErr = err
			}
		case NotifChannelDiscord:
			job := NotifyByDiscordTask.NewJob(alert.EventID, 0)
			if err := app.MainQueue.AddJob(ctx, job); err != nil && firstErr == nil {
				firstErr = err
			}
		case NotifChannelWebhook, NotifChannelAlertmanager:
			job := NotifyByWebhookTask.NewJob(alert.EventID, channel.ID)
			if err := app.MainQueue.AddJob(ctx, job); err != nil && firstErr == nil {
				firstErr = err
			}
		default:
			return fmt.Errorf("unknown notification channel type: %s", channel.Type)
		}
	}

	job := NotifyByDiscordTask.NewJob(alert.EventID, 0)
	if err := app.MainQueue.AddJob(ctx, job); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}

//------------------------------------------------------------------------------

func scheduleNotifyOnMetricAlert(
	ctx context.Context, app *bunapp.App, alert *MetricAlert,
) error {
	monitor, err := org.SelectBaseMonitor(ctx, app, alert.MonitorID)
	if err != nil {
		return err
	}

	recipients, err := selectEmailRecipientsForMonitor(
		ctx,
		app,
		monitor,
		func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("up IS NULL OR up.notify_on_metrics")
		},
	)
	if err != nil {
		return err
	}

	if len(recipients) > 0 {
		job := NotifyByEmailTask.NewJob(alert.EventID, recipients)
		if err := app.MainQueue.AddJob(ctx, job); err != nil {
			return err
		}
	}

	var channels []*BaseNotifChannel

	if err := app.PG.NewSelect().
		Model(&channels).
		Where("project_id = ?", alert.ProjectID).
		Where("state = ?", NotifChannelDelivering).
		Limit(100).
		Apply(func(q *bun.SelectQuery) *bun.SelectQuery {
			subq := app.PG.NewSelect().
				Model((*org.MonitorChannel)(nil)).
				ColumnExpr("channel_id").
				Where("monitor_id = ?", alert.MonitorID)

			return q.Where("id IN (?)", subq)
		}).
		Scan(ctx); err != nil {
		return err
	}

	for _, channel := range channels {
		switch channel.Type {
		case NotifChannelSlack:
			job := NotifyBySlackTask.NewJob(alert.EventID, channel.ID)
			if err := app.MainQueue.AddJob(ctx, job); err != nil {
				return err
			}
		case NotifChannelTelegram:
			job := NotifyByTelegramTask.NewJob(alert.EventID, channel.ID)
			if err := app.MainQueue.AddJob(ctx, job); err != nil {
				return err
			}
		case NotifChannelWebhook, NotifChannelAlertmanager:
			job := NotifyByWebhookTask.NewJob(alert.EventID, channel.ID)
			if err := app.MainQueue.AddJob(ctx, job); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown notification channel type: %s", channel.Type)
		}
	}

	return nil
}

func selectEmailRecipientsForMonitor(
	ctx context.Context,
	app *bunapp.App,
	monitor *org.BaseMonitor,
	cb func(q *bun.SelectQuery) *bun.SelectQuery,
) ([]string, error) {
	if monitor.NotifyEveryoneByEmail {
		return selectAllEmailRecipients(
			ctx,
			app,
			monitor.ProjectID,
			cb,
		)
	}
	return nil, nil
}

func selectAllEmailRecipients(
	ctx context.Context,
	app *bunapp.App,
	projectID uint32,
	cb func(q *bun.SelectQuery) *bun.SelectQuery,
) ([]string, error) {
	var recipients []string

	if err := app.PG.NewSelect().
		ColumnExpr("u.email").
		Model((*org.User)(nil)).
		Join(
			"LEFT JOIN user_project_data AS up ON up.user_id = u.id AND up.project_id = ?",
			projectID,
		).
		Where("u.notify_by_email").
		Apply(cb).
		Scan(ctx, &recipients); err != nil {
		return nil, err
	}

	return recipients, nil
}
