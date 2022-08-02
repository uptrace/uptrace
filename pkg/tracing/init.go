package tracing

import (
	"context"
	"sync"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"go.opentelemetry.io/otel/metric/instrument"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

const (
	protobufContentType = "application/protobuf"
	jsonContentType     = "application/json"
)

var spanCounter, _ = bunapp.Meter.SyncInt64().UpDownCounter(
	"uptrace.projects.spans",
	instrument.WithDescription("Number of processed spans"),
)

func Init(ctx context.Context, app *bunapp.App) {
	initGRPC(ctx, app)
	initRoutes(ctx, app)
}

func initGRPC(ctx context.Context, app *bunapp.App) {
	traceService := NewTraceServiceServer(app, GlobalSpanProcessor(app))
	collectortracepb.RegisterTraceServiceServer(app.GRPCServer(), traceService)

	router := app.Router()
	router.POST("/v1/traces", traceService.httpTraces)
}

func initRoutes(ctx context.Context, app *bunapp.App) {
	router := app.Router()
	sysHandler := NewSystemHandler(app)
	serviceHandler := NewServiceHandler(app)
	hostHandler := NewHostHandler(app)
	spanHandler := NewSpanHandler(app)
	traceHandler := NewTraceHandler(app)
	suggestionHandler := NewSuggestionHandler(app)
	authMiddleware := org.NewAuthMiddleware(app)

	api := app.APIGroup()

	// https://zipkin.io/zipkin-api/#/default/post_spans
	router.WithGroup("/api/v2", func(g *bunrouter.Group) {
		zipkinHandler := NewZipkinHandler(app, GlobalSpanProcessor(app))

		g.POST("/spans", zipkinHandler.PostSpans)
	})

	router.WithGroup("/api/v1", func(g *bunrouter.Group) {
		vectorHandler := NewVectorHandler(app, GlobalSpanProcessor(app))

		g.POST("/vector-logs", vectorHandler.Create)
		g.POST("/vector/logs", vectorHandler.Create)
	})

	api.GET("/traces/search", traceHandler.FindTrace)

	g := api.
		Use(authMiddleware).
		NewGroup("/tracing/:project_id")

	g.GET("/systems", sysHandler.List)
	g.GET("/systems-stats", sysHandler.Stats)
	g.GET("/services", serviceHandler.List)
	g.GET("/hosts", hostHandler.List)
	g.GET("/groups", spanHandler.ListGroups)
	g.GET("/spans", spanHandler.ListSpans)
	g.GET("/percentiles", spanHandler.Percentiles)
	g.GET("/stats", spanHandler.Stats)

	g.GET("/traces/:trace_id", traceHandler.ShowTrace)
	g.GET("/traces/:trace_id/:span_id", traceHandler.ShowSpan)

	g.WithGroup("/suggestions", func(g *bunrouter.Group) {
		g.GET("/attributes", suggestionHandler.Attributes)
		g.GET("/values", suggestionHandler.Values)
	})
}

var (
	spanProcessorOnce sync.Once
	spanProcessor     *SpanProcessor
)

func GlobalSpanProcessor(app *bunapp.App) *SpanProcessor {
	spanProcessorOnce.Do(func() {
		spanProcessor = NewSpanProcessor(app)
	})
	return spanProcessor
}
