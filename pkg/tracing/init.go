package tracing

import (
	"context"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"go.opentelemetry.io/otel/metric/instrument"
	collectorlogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

const (
	protobufContentType  = "application/protobuf"
	xprotobufContentType = "application/x-protobuf"
	jsonContentType      = "application/json"
)

var spanCounter, _ = bunotel.Meter.SyncInt64().Counter(
	"uptrace.projects.spans",
	instrument.WithDescription("Number of processed spans"),
)

func Init(ctx context.Context, app *bunapp.App) {
	sp := NewSpanProcessor(app)

	initOTLP(ctx, app, sp)
	initRoutes(ctx, app, sp)
}

func initOTLP(ctx context.Context, app *bunapp.App, sp *SpanProcessor) {
	traceService := NewTraceServiceServer(app, sp)
	collectortracepb.RegisterTraceServiceServer(app.GRPCServer(), traceService)

	logsService := NewLogsServiceServer(app, sp)
	collectorlogspb.RegisterLogsServiceServer(app.GRPCServer(), logsService)

	router := app.Router()
	router.POST("/v1/traces", traceService.ExportHTTP)
	router.POST("/v1/logs", logsService.ExportHTTP)
}

func initRoutes(ctx context.Context, app *bunapp.App, sp *SpanProcessor) {
	router := app.Router()
	spanHandler := NewSpanHandler(app)
	traceHandler := NewTraceHandler(app)
	middleware := org.NewMiddleware(app)

	api := app.APIGroup()

	// https://zipkin.io/zipkin-api/#/default/post_spans
	router.WithGroup("/api/v2", func(g *bunrouter.Group) {
		zipkinHandler := NewZipkinHandler(app, sp)

		g.POST("/spans", zipkinHandler.PostSpans)
	})

	router.WithGroup("/api/v1", func(g *bunrouter.Group) {
		vectorHandler := NewVectorHandler(app, sp)

		g.POST("/vector-logs", vectorHandler.Create)
		g.POST("/vector/logs", vectorHandler.Create)
	})

	api.GET("/traces/search", traceHandler.FindTrace)

	g := api.
		Use(middleware.UserAndProject).
		NewGroup("/tracing/:project_id")

	api.
		Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			sysHandler := NewSystemHandler(app)

			g.GET("/envs", sysHandler.ListEnvs)
			g.GET("/services", sysHandler.ListServices)

			g.GET("/systems", sysHandler.ListSystems)
			g.GET("/systems-stats", sysHandler.ListSystemStats)
			g.GET("/overview", sysHandler.Overview)
		})

	api.
		Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			attrHandler := NewAttrHandler(app)

			g.GET("/attr-keys", attrHandler.AttrKeys)
			g.GET("/attr-values", attrHandler.AttrValues)
		})

	api.Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id/saved-views", func(g *bunrouter.Group) {
			viewHandler := NewSavedViewHandler(app)

			g.GET("", viewHandler.List)

			g.POST("", viewHandler.Create)
			g.DELETE("/:view_id", viewHandler.Delete)

			g.PUT("/:view_id/pinned", viewHandler.Pin)
			g.PUT("/:view_id/unpinned", viewHandler.Unpin)
		})

	g.GET("/groups", spanHandler.ListGroups)
	g.GET("/spans", spanHandler.ListSpans)
	g.GET("/percentiles", spanHandler.Percentiles)
	g.GET("/stats", spanHandler.Stats)

	g.GET("/traces/:trace_id", traceHandler.ShowTrace)
	g.GET("/traces/:trace_id/:span_id", traceHandler.ShowSpan)
}
