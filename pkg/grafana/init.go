package grafana

import (
	"context"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/tracing"
)

const (
	protobufContentType = "application/protobuf"
	jsonContentType     = "application/json"
)

func init() {
	bunapp.OnStart("grafana.init", func(ctx context.Context, app *bunapp.App) error {
		initRoutes(ctx, app)
		return nil
	})
}

func initRoutes(ctx context.Context, app *bunapp.App) {
	router := app.Router()

	// https://grafana.com/docs/tempo/latest/api_docs/
	router.WithGroup("/api/tempo", func(g *bunrouter.Group) {
		tempoHandler := NewTempoHandler(app)

		g = g.Use(tempoHandler.CheckProjectAccess)

		g.GET("/ready", tempoHandler.Ready)
		g.GET("/api/echo", tempoHandler.Echo)

		g.GET("/api/traces/:trace_id", tempoHandler.QueryTrace)
		g.GET("/api/traces/:trace_id/json", tempoHandler.QueryTraceJSON)

		g.GET("/api/search/tags", tempoHandler.Tags)
		g.GET("/api/search/tag/:tag/values", tempoHandler.TagValues)
		g.GET("/api/search", tempoHandler.Search)
	})

	router.WithGroup("", func(g *bunrouter.Group) {
		lokiProxyHandler := NewLokiProxyHandler(app, tracing.GlobalSpanProcessor(app))

		g.GET("/ready", lokiProxyHandler.Ready)

		g.Use(lokiProxyHandler.CheckProjectAccess).
			WithGroup("/loki/api", func(g *bunrouter.Group) {
				registerLokiProxy(g, lokiProxyHandler)
			})

		g.Use(lokiProxyHandler.CheckProjectAccess).
			Use(lokiProxyHandler.trimProjectID).
			WithGroup("/:project_id/loki/api", func(g *bunrouter.Group) {
				registerLokiProxy(g, lokiProxyHandler)
			})
	})
}

func registerLokiProxy(g *bunrouter.Group, lokiProxyHandler *LokiProxyHandler) {
	g.GET("/v1/tail", lokiProxyHandler.ProxyWS)
	g.POST("/v1/push", lokiProxyHandler.Push)

	g.GET("/*path", lokiProxyHandler.Proxy)
	g.POST("/*path", lokiProxyHandler.Proxy)
	g.PUT("/*path", lokiProxyHandler.Proxy)
	g.DELETE("/*path", lokiProxyHandler.Proxy)
}
