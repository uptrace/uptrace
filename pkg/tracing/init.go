package tracing

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	collectorlogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/uptrace/pkg/clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/run"
	"github.com/uptrace/uptrace/pkg/tracing/norm"
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
		NewSpanConsumer,
		NewLogConsumer,
		NewEventConsumer,
		NewTraceServiceServer,
		NewLogsServiceServer,

		NewVectorHandler,
		NewSystemHandler,
		NewAttrHandler,
		NewSavedViewHandler,
		NewSpanHandler,
		NewGroupHandler,
		NewPublicHandler,
		NewTraceHandler,
	),
	fx.Invoke(
		registerVectorHandler,
		registerSystemHandler,
		registerAttrHandler,
		registerSavedViewHandler,
		registerSpanHandler,
		registerGroupHandler,
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

func init() {
	ch.RegisterEnum("log_severity", logSeverityEnum)
}

var logSeverityEnum = []string{
	"",
	norm.SeverityTrace,
	norm.SeverityTrace2,
	norm.SeverityTrace3,
	norm.SeverityTrace4,
	norm.SeverityDebug,
	norm.SeverityDebug2,
	norm.SeverityDebug3,
	norm.SeverityDebug4,
	norm.SeverityInfo,
	norm.SeverityInfo2,
	norm.SeverityInfo3,
	norm.SeverityInfo4,
	norm.SeverityWarn,
	norm.SeverityWarn2,
	norm.SeverityWarn3,
	norm.SeverityWarn4,
	norm.SeverityError,
	norm.SeverityError2,
	norm.SeverityError3,
	norm.SeverityError4,
	norm.SeverityFatal,
	norm.SeverityFatal2,
	norm.SeverityFatal3,
	norm.SeverityFatal4,
}
