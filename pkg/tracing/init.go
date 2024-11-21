package tracing

import (
	"context"
	"runtime"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"go.opentelemetry.io/otel/metric"
	collectorlogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"go4.org/syncutil"
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
	maxprocs := runtime.GOMAXPROCS(0)
	gate := syncutil.NewGate(maxprocs)
	sp := NewSpanConsumer(app, gate)
	lc := NewLogConsumer(app, gate)

	initOTLP(ctx, app, sp)
	initRoutes(ctx, app, sp, lc)
}

func initOTLP(ctx context.Context, app *bunapp.App, sp *SpanConsumer) {
	traceService := NewTraceServiceServer(app, sp)
	collectortracepb.RegisterTraceServiceServer(app.GRPCServer(), traceService)

	logsService := NewLogsServiceServer(app, sp)
	collectorlogspb.RegisterLogsServiceServer(app.GRPCServer(), logsService)

	router := app.Router()
	router.POST("/v1/traces", traceService.ExportHTTP)
	router.POST("/v1/logs", logsService.ExportHTTP)
}

func initRoutes(ctx context.Context, app *bunapp.App, sp *SpanConsumer, lc *LogConsumer) {
	router := app.Router()
	middleware := org.NewMiddleware(app)
	internalV1 := app.InternalAPIV1()
	publicV1 := app.PublicAPIV1()

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
		vectorHandler := NewVectorHandler(app, lc)

		g.POST("/vector-logs", vectorHandler.Create)
		g.POST("/vector/logs", vectorHandler.Create)
	})

	router.WithGroup("/api/v1/cloudwatch", func(g *bunrouter.Group) {
		handler := NewKinesisHandler(app, sp)

		g.POST("/logs", handler.Logs)
	})

	internalV1.
		Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			sysHandler := NewSystemHandler(app)

			g.GET("/systems", sysHandler.ListSystems)
		})

	internalV1.
		Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			attrHandler := NewAttrHandler(app)

			g.GET("/attributes", attrHandler.AttrKeys)
			g.GET("/attributes/:attr", attrHandler.AttrValues)
		})

	internalV1.Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id/saved-views", func(g *bunrouter.Group) {
			viewHandler := NewSavedViewHandler(app)

			g.GET("", viewHandler.List)

			g.POST("", viewHandler.Create)
			g.DELETE("/:view_id", viewHandler.Delete)

			g.PUT("/:view_id/pinned", viewHandler.Pin)
			g.PUT("/:view_id/unpinned", viewHandler.Unpin)
		})

	internalV1.Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			spanHandler := NewSpanHandler(app)

			g.GET("/groups", spanHandler.ListGroups)
			g.GET("/spans", spanHandler.ListSpans)
			g.GET("/percentiles", spanHandler.Percentiles)
			g.GET("/group-stats", spanHandler.GroupStats)
			g.GET("/timeseries", spanHandler.Timeseries)
		})

	internalV1.Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id/groups/:group_id", func(g *bunrouter.Group) {
			groupHandler := NewGroupHandler(app)

			g.GET("", groupHandler.ShowSummary)
		})

	internalV1.Use(middleware.User).
		WithGroup("", func(g *bunrouter.Group) {
			traceHandler := NewTraceHandler(app)

			g.GET("/traces/search", traceHandler.FindTrace)

			g = g.Use(middleware.UserAndProject).NewGroup("/tracing/:project_id")

			g.GET("/traces/:trace_id", traceHandler.ShowTrace)
			g.GET("/traces/:trace_id/:span_id", traceHandler.ShowSpan)
		})

	internalV1.WithGroup("/tracing/:project_id/service-graph", func(g *bunrouter.Group) {
		g = g.Use(middleware.UserAndProject)

		handler := NewServiceGraphHandler(app)

		g.GET("", handler.List)
	})

	publicV1.
		Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			handler := NewPublicHandler(app)

			g.GET("/spans", handler.Spans)
			g.GET("/groups", handler.Groups)
		})
}
