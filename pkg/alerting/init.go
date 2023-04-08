package alerting

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/vmihailenco/taskq/v4"
)

func Init(ctx context.Context, app *bunapp.App) {
	initRouter(ctx, app)
	initTasks(ctx, app)
}

func initRouter(ctx context.Context, app *bunapp.App) {
	middleware := NewMiddleware(app)
	api := app.APIGroup()

	api.NewGroup("/projects/:project_id",
		bunrouter.WithMiddleware(middleware.UserAndProject),
		bunrouter.WithGroup(func(g *bunrouter.Group) {
			alertHandler := NewAlertHandler(app)

			g.GET("/alerts", alertHandler.List)
			g.GET("/alerts/:alert_id", alertHandler.Show)
			g.PUT("/alerts/closed", alertHandler.Close)
			g.PUT("/alerts/open", alertHandler.Open)
			g.PUT("/alerts/close-all", alertHandler.CloseAll)
		}))

	api.
		Use(middleware.UserAndProject).
		WithGroup("/projects/:project_id/monitors", func(g *bunrouter.Group) {
			monitorHandler := NewMonitorHandler(app)

			g.GET("", monitorHandler.List)

			g.POST("/metric", monitorHandler.CreateMetricMonitor)
			g.POST("/error", monitorHandler.CreateErrorMonitor)

			g = g.NewGroup("/:monitor_id").
				Use(middleware.Monitor)

			g.GET("", monitorHandler.Show)
			g.DELETE("", monitorHandler.Delete)

			g.PUT("/metric", monitorHandler.UpdateMetricMonitor)
			g.PUT("/error", monitorHandler.UpdateErrorMonitor)

			g.PUT("/active", monitorHandler.Activate)
			g.PUT("/paused", monitorHandler.Pause)
		})

	api.
		Use(middleware.UserAndProject).
		WithGroup("/projects/:project_id/notification-channels", func(g *bunrouter.Group) {
			handler := NewNotifChannelHandler(app)

			g.GET("", handler.List)

			g.POST("/slack", handler.SlackCreate)
			g.POST("/webhook", handler.WebhookCreate)

			g.GET("/email", handler.EmailShow)
			g.PUT("/email", handler.EmailUpdate)

			g = g.Use(middleware.NotifChannel)

			g.DELETE("/:channel_id", handler.Delete)
			g.PUT("/:channel_id/paused", handler.Pause)
			g.PUT("/:channel_id/unpaused", handler.Unpause)

			g.GET("/slack/:channel_id", handler.SlackShow)
			g.PUT("/slack/:channel_id", handler.SlackUpdate)

			g.GET("/webhook/:channel_id", handler.WebhookShow)
			g.PUT("/webhook/:channel_id", handler.WebhookUpdate)
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

		monitor, err := SelectMonitor(ctx, m.App, monitorID)
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

func monitorFromContext(ctx context.Context) Monitor {
	return ctx.Value(monitorCtxKey{}).(Monitor)
}

func metricMonitorFromContext(ctx context.Context) (*MetricMonitor, error) {
	monitor, ok := ctx.Value(monitorCtxKey{}).(*MetricMonitor)
	if !ok {
		return nil, sql.ErrNoRows
	}
	return monitor, nil
}

func errorMonitorFromContext(ctx context.Context) (*ErrorMonitor, error) {
	monitor, ok := ctx.Value(monitorCtxKey{}).(*ErrorMonitor)
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

		channel, err := SelectNotifChannel(ctx, m.App, channelID)
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
	NotifyByEmailTask   = taskq.NewTask("notify-by-email")
	NotifyBySlackTask   = taskq.NewTask("notify-by-slack")
	NotifyByWebhookTask = taskq.NewTask("notify-by-webhook")
)

func initTasks(ctx context.Context, app *bunapp.App) {
	_ = app.RegisterTask(org.CreateErrorAlertTask.Name(), &taskq.TaskConfig{
		Handler: createErrorAlertHandler,
	})
	_ = app.RegisterTask(NotifyByEmailTask.Name(), &taskq.TaskConfig{
		Handler: NewEmailNotifier(app).NotifyHandler,
	})
	_ = app.RegisterTask(NotifyBySlackTask.Name(), &taskq.TaskConfig{
		Handler: notifyBySlackHandler,
	})
	_ = app.RegisterTask(NotifyByWebhookTask.Name(), &taskq.TaskConfig{
		Handler: notifyByWebhookHandler,
	})
}
