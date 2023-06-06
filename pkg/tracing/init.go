package tracing

import (
	"context"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"go.opentelemetry.io/otel/metric"
	collectorlogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

const (
	protobufContentType  = "application/protobuf"
	xprotobufContentType = "application/x-protobuf"
	jsonContentType      = "application/json"
)

var spanCounter, _ = bunotel.Meter.Int64Counter(
	"uptrace.projects.spans",
	metric.WithDescription("Number of processed spans"),
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
	middleware := org.NewMiddleware(app)

	api := app.APIGroup()

	// https://zipkin.io/zipkin-api/#/default/post_spans
	router.WithGroup("/api/v2", func(g *bunrouter.Group) {
		zipkinHandler := NewZipkinHandler(app, sp)

		g.POST("/spans", zipkinHandler.PostSpans)
	})

	router.WithGroup("/api", func(g *bunrouter.Group) {
		sentryHandler := NewSentryHandler(app, sp)

		g.POST("/:project_id/store/", sentryHandler.Store)
		g.POST("/:project_id/envelope/", sentryHandler.Envelope)
	})

	router.WithGroup("/api/v1", func(g *bunrouter.Group) {
		vectorHandler := NewVectorHandler(app, sp)

		g.POST("/vector-logs", vectorHandler.Create)
		g.POST("/vector/logs", vectorHandler.Create)
	})

	api.
		Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			sysHandler := NewSystemHandler(app)

			g.GET("/systems", sysHandler.ListSystems)
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

	api.
		Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			spanHandler := NewSpanHandler(app)

			g.GET("/groups", spanHandler.ListGroups)
			g.GET("/spans", spanHandler.ListSpans)
			g.GET("/percentiles", spanHandler.Percentiles)
			g.GET("/group-stats", spanHandler.GroupStats)
			g.GET("/timeseries", spanHandler.Timeseries)
		})

	api.
		Use(middleware.User).
		WithGroup("", func(g *bunrouter.Group) {
			traceHandler := NewTraceHandler(app)

			g.GET("/traces/search", traceHandler.FindTrace)

			g = g.Use(middleware.UserAndProject).NewGroup("/tracing/:project_id")

			g.GET("/traces/:trace_id", traceHandler.ShowTrace)
			g.GET("/traces/:trace_id/:span_id", traceHandler.ShowSpan)
		})

	api.WithGroup("/cloudwatch", func(g *bunrouter.Group) {
		handler := NewKinesisHandler(app, sp)

		g.POST("/logs", handler.Logs)
	})
}
