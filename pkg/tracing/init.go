package tracing

import (
	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/vmihailenco/taskq/v4"
	"go.opentelemetry.io/otel/metric"
	collectorlogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
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

type ModuleParams struct {
	fx.In
	bunapp.RouterParams

	Conf      *bunconf.Config
	Logger    *otelzap.Logger
	GRPC      *grpc.Server
	PG        *bun.DB
	CH        *ch.DB
	MainQueue taskq.Queue
}

func Init(p *ModuleParams) {
	sc := NewSpanConsumer(p)
	lc := NewLogConsumer(p)

	initOTLP(p, sc)
	initRoutes(p, sc, lc)
}

func initOTLP(p *ModuleParams, sp *SpanConsumer) {
	fakeApp := &bunapp.App{
		PG: p.PG,
	}
	traceService := NewTraceServiceServer(fakeApp, sp)
	collectortracepb.RegisterTraceServiceServer(p.GRPC, traceService)

	logsService := NewLogsServiceServer(fakeApp, sp)
	collectorlogspb.RegisterLogsServiceServer(p.GRPC, logsService)

	p.Router.POST("/v1/traces", traceService.ExportHTTP)
	p.Router.POST("/v1/logs", logsService.ExportHTTP)
}

func initRoutes(p *ModuleParams, spanConsumer *SpanConsumer, logConsumer *LogConsumer) {
	fakeApp := &bunapp.App{
		Conf:   p.Conf,
		Logger: p.Logger,
		PG:     p.PG,
		CH:     p.CH,
	}
	middleware := org.NewMiddleware(fakeApp)
	internalV1 := p.RouterInternalV1
	publicV1 := p.RouterPublicV1

	// https://zipkin.io/zipkin-api/#/default/post_spans
	p.Router.WithGroup("/api/v2", func(g *bunrouter.Group) {
		zipkinHandler := NewZipkinHandler(p.Logger, p.PG, spanConsumer)

		g.POST("/spans", zipkinHandler.PostSpans)
	})

	p.Router.WithGroup("/api", func(g *bunrouter.Group) {
		sentryHandler := NewSentryHandler(p.Logger, p.PG, spanConsumer)

		g.POST("/:project_id/store/", sentryHandler.Store)
		g.POST("/:project_id/envelope/", sentryHandler.Envelope)
	})

	p.Router.WithGroup("/api/v1", func(g *bunrouter.Group) {
		vectorHandler := NewVectorHandler(p.Logger, p.PG, logConsumer)

		g.POST("/vector-logs", vectorHandler.Create)
		g.POST("/vector/logs", vectorHandler.Create)
	})

	p.Router.WithGroup("/api/v1/cloudwatch", func(g *bunrouter.Group) {
		handler := NewKinesisHandler(p.Logger, p.PG, spanConsumer)

		g.POST("/logs", handler.Logs)
	})

	internalV1.
		Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			sysHandler := NewSystemHandler(p.Logger, p.CH)

			g.GET("/systems", sysHandler.ListSystems)
		})

	internalV1.
		Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			attrHandler := NewAttrHandler(p.Logger, p.PG, p.CH)

			g.GET("/attributes", attrHandler.AttrKeys)
			g.GET("/attributes/:attr", attrHandler.AttrValues)
		})

	internalV1.Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id/saved-views", func(g *bunrouter.Group) {
			viewHandler := NewSavedViewHandler(p.Logger, p.PG)

			g.GET("", viewHandler.List)

			g.POST("", viewHandler.Create)
			g.DELETE("/:view_id", viewHandler.Delete)

			g.PUT("/:view_id/pinned", viewHandler.Pin)
			g.PUT("/:view_id/unpinned", viewHandler.Unpin)
		})

	internalV1.Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			spanHandler := NewSpanHandler(p.Logger, p.CH)

			g.GET("/groups", spanHandler.ListGroups)
			g.GET("/spans", spanHandler.ListSpans)
			g.GET("/percentiles", spanHandler.Percentiles)
			g.GET("/group-stats", spanHandler.GroupStats)
			g.GET("/timeseries", spanHandler.Timeseries)
		})

	internalV1.Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id/groups/:group_id", func(g *bunrouter.Group) {
			groupHandler := NewGroupHandler(p.Logger, p.CH)

			g.GET("", groupHandler.ShowSummary)
		})

	internalV1.Use(middleware.User).
		WithGroup("", func(g *bunrouter.Group) {
			traceHandler := NewTraceHandler(p.Logger, p.CH)

			g.GET("/traces/search", traceHandler.FindTrace)

			g = g.Use(middleware.UserAndProject).NewGroup("/tracing/:project_id")

			g.GET("/traces/:trace_id", traceHandler.ShowTrace)
			g.GET("/traces/:trace_id/:span_id", traceHandler.ShowSpan)
		})

	internalV1.WithGroup("/tracing/:project_id/service-graph", func(g *bunrouter.Group) {
		g = g.Use(middleware.UserAndProject)

		handler := NewServiceGraphHandler(p.Logger, p.CH)

		g.GET("", handler.List)
	})

	publicV1.
		Use(middleware.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			handler := NewPublicHandler(p.Logger, p.CH)

			g.GET("/spans", handler.Spans)
			g.GET("/groups", handler.Groups)
		})
}
