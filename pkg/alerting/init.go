package alerting

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/vmihailenco/taskq/v4"
	"go.uber.org/fx"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/org"
)

var Module = fx.Module("alerting",
	fx.Provide(
		fx.Private,
		NewMiddleware,
		NewAlertHandler,
		NewMonitorHandler,
		NewNotifChannelHandler,
	),
	fx.Invoke(
		fx.Annotate(initRouter, fx.ParamTags(`name:"router_internal_apiv1"`)),
		initTasks,
	),
)

func initRouter(
	api *bunrouter.Group,
	middleware Middleware,
	alertHandler *AlertHandler,
	monitorHandler *MonitorHandler,
	notifChannelHandler *NotifChannelHandler,
) {
	api.NewGroup("/projects/:project_id",
		bunrouter.WithMiddleware(middleware.UserAndProject),
		bunrouter.WithGroup(func(g *bunrouter.Group) {
			g.GET("/alerts", alertHandler.List)
			g.GET("/alerts/:alert_id", alertHandler.Show)
			g.PUT("/alerts/closed", alertHandler.Close)
			g.PUT("/alerts/open", alertHandler.Open)
			g.DELETE("/alerts", alertHandler.Delete)
		}))

	api.
		Use(middleware.UserAndProject).
		WithGroup("/projects/:project_id/monitors", func(g *bunrouter.Group) {
			g.GET("", monitorHandler.List)

			g.POST("/yaml", monitorHandler.CreateMonitorFromYAML)
			g.POST("/metric", monitorHandler.CreateMetricMonitor)
			g.POST("/error", monitorHandler.CreateErrorMonitor)

			g = g.NewGroup("/:monitor_id").
				Use(middleware.Monitor)

			g.GET("", monitorHandler.Show)
			g.GET("/yaml", monitorHandler.ShowYAML)
			g.DELETE("", monitorHandler.Delete)

			g.PUT("/metric", monitorHandler.UpdateMetricMonitor)
			g.PUT("/error", monitorHandler.UpdateErrorMonitor)

			g.PUT("/active", monitorHandler.Activate)
			g.PUT("/paused", monitorHandler.Pause)
		})

	api.
		Use(middleware.UserAndProject).
		WithGroup("/projects/:project_id/notification-channels", func(g *bunrouter.Group) {
			g.GET("", notifChannelHandler.List)

			g.POST("/slack", notifChannelHandler.SlackCreate)
			g.POST("/webhook", notifChannelHandler.WebhookCreate)
			g.POST("/telegram", notifChannelHandler.TelegramCreate)

			g.GET("/email", notifChannelHandler.EmailShow)
			g.PUT("/email", notifChannelHandler.EmailUpdate)

			g = g.Use(middleware.NotifChannel)

			g.DELETE("/:channel_id", notifChannelHandler.Delete)
			g.PUT("/:channel_id/paused", notifChannelHandler.Pause)
			g.PUT("/:channel_id/unpaused", notifChannelHandler.Unpause)

			g.GET("/slack/:channel_id", notifChannelHandler.SlackShow)
			g.PUT("/slack/:channel_id", notifChannelHandler.SlackUpdate)

			g.GET("/webhook/:channel_id", notifChannelHandler.WebhookShow)
			g.PUT("/webhook/:channel_id", notifChannelHandler.WebhookUpdate)

			g.GET("/telegram/:channel_id", notifChannelHandler.TelegramShow)
			g.PUT("/telegram/:channel_id", notifChannelHandler.TelegramUpdate)
		})
}

//------------------------------------------------------------------------------

type Middleware struct {
	App *bunapp.App
	*org.Middleware
}

func NewMiddleware(app *bunapp.App) Middleware {
	return Middleware{
		App:        app,
		Middleware: org.NewMiddleware(app),
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

		monitor, err := org.SelectMonitor(ctx, m.App, monitorID)
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

		channel, err := SelectNotifChannel(ctx, m.App.PG, channelID)
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

func initTasks(logger *otelzap.Logger, conf *bunconf.Config, handler *NotifChannelHandler) {
	registerTaskHandler(org.CreateErrorAlertTask.Name(), createErrorAlertHandler)
	registerTaskHandler(NotifyByEmailTask.Name(), NewEmailNotifier(logger, conf).NotifyHandler)
	registerTaskHandler(NotifyByTelegramTask.Name(), handler.notifyByTelegramHandler)
	registerTaskHandler(NotifyBySlackTask.Name(), handler.notifyBySlackHandler)
	registerTaskHandler(NotifyByWebhookTask.Name(), handler.notifyByWebhookHandler)
}

func registerTaskHandler(name string, handler any) {
	_ = taskq.RegisterTask(name, &taskq.TaskConfig{
		RetryLimit: 16,
		Handler:    handler,
	})
}
