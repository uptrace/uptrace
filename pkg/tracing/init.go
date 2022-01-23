package tracing

import (
	"context"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

func init() {
	bunapp.OnStart("tracing.initGRPC", initGRPC)
	bunapp.OnStart("tracing.registerRoutes", registerRoutes)
}

func initGRPC(ctx context.Context, app *bunapp.App) error {
	traceService := NewTraceServiceServer(app)
	collectortrace.RegisterTraceServiceServer(app.GRPCServer(), traceService)

	router := app.Router()
	router.POST("/v1/traces", traceService.httpTraces)

	return nil
}

func registerRoutes(ctx context.Context, app *bunapp.App) error {
	sysHandler := NewSystemHandler(app)
	serviceHandler := NewServiceHandler(app)
	hostHandler := NewHostHandler(app)
	spanHandler := NewSpanHandler(app)
	traceHandler := NewTraceHandler(app)
	suggestionHandler := NewSuggestionHandler(app)

	g := app.APIGroup().
		Use(org.NewAuthMiddleware(app)).
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

	return nil
}
