package tracing

import (
	"context"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	collectormetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

func init() {
	bunapp.OnStart("tracing.init", func(ctx context.Context, app *bunapp.App) error {
		sp := NewSpanProcessor(app)
		initGRPC(ctx, app, sp)
		initRoutes(ctx, app, sp)
		return nil
	})
}

func initGRPC(ctx context.Context, app *bunapp.App, sp *SpanProcessor) {
	traceService := NewTraceServiceServer(app, sp)
	collectortracepb.RegisterTraceServiceServer(app.GRPCServer(), traceService)

	metricsService := NewMetricsServiceServer(app)
	collectormetricspb.RegisterMetricsServiceServer(app.GRPCServer(), metricsService)

	router := app.Router()
	router.POST("/v1/traces", traceService.httpTraces)
}

func initRoutes(ctx context.Context, app *bunapp.App, sp *SpanProcessor) {
	router := app.Router()
	sysHandler := NewSystemHandler(app)
	serviceHandler := NewServiceHandler(app)
	hostHandler := NewHostHandler(app)
	spanHandler := NewSpanHandler(app)
	traceHandler := NewTraceHandler(app)
	suggestionHandler := NewSuggestionHandler(app)
	authMiddleware := org.NewAuthMiddleware(app)

	api := app.APIGroup()

	// https://grafana.com/docs/tempo/latest/api_docs/
	router.WithGroup("", func(g *bunrouter.Group) {
		tempoHandler := NewTempoHandler(app)

		g = g.Use(tempoHandler.checkProjectAccess)

		g.GET("/ready", tempoHandler.Ready)
		g.GET("/api/echo", tempoHandler.Echo)

		g.GET("/api/traces/:trace_id", tempoHandler.QueryTrace)
		g.GET("/api/traces/:trace_id/json", tempoHandler.QueryTraceJSON)

		g.GET("/api/search/tags", tempoHandler.Tags)
		g.GET("/api/search/tag/:tag/values", tempoHandler.TagValues)
		g.GET("/api/search", tempoHandler.Search)
	})

	// https://zipkin.io/zipkin-api/#/default/post_spans
	router.WithGroup("/api/v2", func(g *bunrouter.Group) {
		zipkinHandler := NewZipkinHandler(app, sp)

		g.POST("/spans", zipkinHandler.PostSpans)
	})

	router.WithGroup("/api/v1", func(g *bunrouter.Group) {
		vectorHandler := NewVectorHandler(app, sp)

		g.POST("/vector-logs", vectorHandler.Create)
	})

	{
		lokiProxyHandler := NewLokiProxyHandler(app, sp)

		router.
			Use(lokiProxyHandler.checkProjectAccess).
			WithGroup("/loki/api", func(g *bunrouter.Group) {
				registerLokiProxy(g, lokiProxyHandler)
			})

		router.
			Use(lokiProxyHandler.checkProjectAccess).
			Use(lokiProxyHandler.trimProjectID).
			WithGroup("/:project_id/loki/api", func(g *bunrouter.Group) {
				registerLokiProxy(g, lokiProxyHandler)
			})
	}

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

	g.GET("/conn-info", func(w http.ResponseWriter, req bunrouter.Request) error {
		projectID, err := req.Params().Uint32("project_id")
		if err != nil {
			return err
		}

		project, err := org.SelectProjectByID(ctx, app, projectID)
		if err != nil {
			return err
		}

		return httputil.JSON(w, bunrouter.H{
			"grpc": map[string]any{
				"endpoint": app.Config().GRPCEndpoint(project),
				"dsn":      app.Config().GRPCDsn(project),
			},
			"http": map[string]any{
				"endpoint": app.Config().HTTPEndpoint(project),
				"dsn":      app.Config().HTTPDsn(project),
			},
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
