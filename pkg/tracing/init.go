package tracing

import (
	"context"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

func init() {
	bunapp.OnStart("tracing.initGRPC", initGRPC)
	bunapp.OnStart("tracing.registerRoutes", registerRoutes)
}

func initGRPC(ctx context.Context, app *bunapp.App) error {
	collectortrace.RegisterTraceServiceServer(app.GRPCServer(), NewTraceServiceServer(app))
	return nil
}

func registerRoutes(ctx context.Context, app *bunapp.App) error {
	sysHandler := NewSystemHandler(app)
	spanHandler := NewSpanHandler(app)
	traceHandler := NewTraceHandler(app)
	suggestionHandler := NewSuggestionHandler(app)

	g := app.APIGroup().NewGroup("/tracing")

	g.GET("/systems", sysHandler.List)
	g.GET("/systems-stats", sysHandler.Stats)
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
		return bunrouter.JSON(w, bunrouter.H{
			"dsn":  app.Config().UptraceDSN(),
			"otlp": app.Config().OTLPEndpoint(),
		})
	})

	return nil
}
