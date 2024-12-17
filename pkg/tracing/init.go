package tracing

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	collectorlogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/run"
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

var Module = fx.Module("tracing",
	fx.Provide(
		fx.Private,

		org.NewMiddleware,
		NewServiceGraphProcessor,
		NewSpanConsumer,
		NewLogConsumer,
		NewEventConsumer,
		NewTraceServiceServer,
		NewLogsServiceServer,

		NewZipkinHandler,
		NewSentryHandler,
		NewVectorHandler,
		NewKinesisHandler,
		NewSystemHandler,
		NewAttrHandler,
		NewSavedViewHandler,
		NewSpanHandler,
		NewGroupHandler,
		NewServiceGraphHandler,
		NewPublicHandler,
		NewTraceHandler,
	),
	fx.Invoke(
		registerZipkinHandler,
		registerSentryHandler,
		registerVectorHandler,
		registerKinesisHandler,
		registerSystemHandler,
		registerAttrHandler,
		registerSavedViewHandler,
		registerSpanHandler,
		registerGroupHandler,
		registerServiceGraphHandler,
		registerPublicHandler,
		registerTraceHandler,

		initOTLP,
		runConsumers,
	),
)

type OTLPParams struct {
	fx.In

	GRPC        *grpc.Server
	TraceServer *TraceServiceServer
	LogsServer  *LogsServiceServer
}

func initOTLP(p OTLPParams, router bunapp.RouterParams) {
	collectortracepb.RegisterTraceServiceServer(p.GRPC, p.TraceServer)
	collectorlogspb.RegisterLogsServiceServer(p.GRPC, p.LogsServer)

	router.Router.POST("/v1/traces", p.TraceServer.ExportHTTP)
	router.Router.POST("/v1/logs", p.LogsServer.ExportHTTP)
}

func runConsumers(
	group *run.Group,
	spanConsumer *SpanConsumer,
	logConsumer *LogConsumer,
	eventConsumer *EventConsumer,
) {
	group.Add("tracing.spanConsumer.Run", func() error {
		spanConsumer.Run()
		return nil
	})
	group.OnStop(func(context.Context, error) error {
		spanConsumer.Stop()
		return nil
	})
	group.Add("tracing.logConsumer.Run", func() error {
		logConsumer.Run()
		return nil
	})
	group.OnStop(func(context.Context, error) error {
		logConsumer.Stop()
		return nil
	})
	group.Add("tracing.eventConsumer.Run", func() error {
		eventConsumer.Run()
		return nil
	})
	group.OnStop(func(context.Context, error) error {
		eventConsumer.Stop()
		return nil
	})
}
