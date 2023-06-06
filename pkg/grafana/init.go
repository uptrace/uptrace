package grafana

import (
	"context"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

const (
	protobufContentType = "application/protobuf"
	jsonContentType     = "application/json"
)

func Init(ctx context.Context, app *bunapp.App) {
	initRoutes(ctx, app)
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
}
