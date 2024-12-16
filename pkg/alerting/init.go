package alerting

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/vmihailenco/taskq/v4"
	"go.uber.org/fx"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/org"
)

var Module = fx.Module("alerting",
	fx.Provide(
		fx.Private,
		NewMiddleware,
		NewEmailNotifier,
		NewAlertNotifier,
		NewManager,

		NewAlertHandler,
		NewMonitorHandler,
		NewNotifChannelHandler,
	),
	fx.Invoke(
		registerAlertHandler,
		registerMonitorHandler,
		registerNotifChannelHandler,

		initTasks,
		runManager,
	),
)

type Middleware struct {
	*org.Middleware
}

func NewMiddleware(p org.MiddlewareParams) *Middleware {
	return &Middleware{
		Middleware: org.NewMiddleware(p),
	}
}

type monitorCtxKey struct{}

func (m *Middleware) Monitor(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()

		monitorID, err := req.Params().Uint64("monitor_id")
		if err != nil {
			return err
		}

		monitor, err := org.SelectMonitor(ctx, m.PG, monitorID)
		if err != nil {
			return err
		}

		project := org.ProjectFromContext(ctx)
		if monitor.Base().ProjectID != project.ID {
			return org.ErrAccessDenied
		}

		ctx = context.WithValue(ctx, monitorCtxKey{}, monitor)
		return next(w, req.WithContext(ctx))
	}
}

func monitorFromContext(ctx context.Context) org.Monitor {
	return ctx.Value(monitorCtxKey{}).(org.Monitor)
}

func metricMonitorFromContext(ctx context.Context) (*org.MetricMonitor, error) {
	monitor, ok := ctx.Value(monitorCtxKey{}).(*org.MetricMonitor)
	if !ok {
		return nil, sql.ErrNoRows
	}
	return monitor, nil
}

func errorMonitorFromContext(ctx context.Context) (*org.ErrorMonitor, error) {
	monitor, ok := ctx.Value(monitorCtxKey{}).(*org.ErrorMonitor)
	if !ok {
		return nil, sql.ErrNoRows
	}
	return monitor, nil
}

//------------------------------------------------------------------------------

type notifChannelCtxKey struct{}

func NotifChannelFromContext(ctx context.Context) NotifChannel {
	return ctx.Value(notifChannelCtxKey{}).(NotifChannel)
}

func SlackNotifChannelFromContext(ctx context.Context) (*SlackNotifChannel, error) {
	channelAny := NotifChannelFromContext(ctx)
	channel, ok := channelAny.(*SlackNotifChannel)
	if !ok {
		return nil, fmt.Errorf("unexpected notification channel: %T", channelAny)
	}
	return channel, nil
}

func TelegramNotifChannelFromContext(ctx context.Context) (*TelegramNotifChannel, error) {
	channelAny := NotifChannelFromContext(ctx)

	channel, ok := channelAny.(*TelegramNotifChannel)
	if !ok {
		return nil, fmt.Errorf("unexpected notification channel: %T", channelAny)
	}
	return channel, nil
}

func WebhookNotifChannelFromContext(ctx context.Context) (*WebhookNotifChannel, error) {
	channelAny := NotifChannelFromContext(ctx)
	channel, ok := channelAny.(*WebhookNotifChannel)
	if !ok {
		return nil, fmt.Errorf("unexpected notification channel: %T", channelAny)
	}
	return channel, nil
}

func (m *Middleware) NotifChannel(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()
		project := org.ProjectFromContext(ctx)

		channelID, err := req.Params().Uint64("channel_id")
		if err != nil {
			return err
		}

		channel, err := SelectNotifChannel(ctx, m.PG, channelID)
		if err != nil {
			return err
		}

		if channel.Base().ProjectID != project.ID {
			msg := "you don't have enough permissions to access this notification channel"
			return httperror.Forbidden(msg)
		}

		ctx = context.WithValue(ctx, notifChannelCtxKey{}, channel)
		return next(w, req.WithContext(ctx))
	}
}

//------------------------------------------------------------------------------

var (
	NotifyByEmailTask    = taskq.NewTask("notify-by-email")
	NotifyBySlackTask    = taskq.NewTask("notify-by-slack")
	NotifyByWebhookTask  = taskq.NewTask("notify-by-webhook")
	NotifyByTelegramTask = taskq.NewTask("notify-by-telegram")
)

func initTasks(
	alertNotifier *AlertNotifier,
	emailNotifier *EmailNotifier,
	handler *NotifChannelHandler,
) {
	bunapp.RegisterTaskHandler(org.CreateErrorAlertTask.Name(), alertNotifier.ErrorHandler)
	bunapp.RegisterTaskHandler(NotifyByEmailTask.Name(), emailNotifier.NotifyHandler)
	bunapp.RegisterTaskHandler(NotifyByTelegramTask.Name(), handler.notifyByTelegramHandler)
	bunapp.RegisterTaskHandler(NotifyBySlackTask.Name(), handler.notifyBySlackHandler)
	bunapp.RegisterTaskHandler(NotifyByWebhookTask.Name(), handler.notifyByWebhookHandler)
}
