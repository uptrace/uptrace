package tracing

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/otlpconv"
	"github.com/uptrace/uptrace/pkg/tracing/attrkey"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	collectorlogs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type LogsServiceServer struct {
	collectorlogs.UnimplementedLogsServiceServer

	*bunapp.App

	sp *SpanProcessor
}

var _ collectorlogs.LogsServiceServer = (*LogsServiceServer)(nil)

func NewLogsServiceServer(app *bunapp.App, sp *SpanProcessor) *LogsServiceServer {
	return &LogsServiceServer{
		App: app,
		sp:  sp,
	}
}

func (s *LogsServiceServer) Export(
	ctx context.Context, req *collectorlogs.ExportLogsServiceRequest,
) (*collectorlogs.ExportLogsServiceResponse, error) {
	fmt.Println("UP")

	if ctx.Err() == context.Canceled {
		return nil, status.Error(codes.Canceled, "Client cancelled, abandoning.")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("metadata is empty")
	}

	dsn := md.Get("uptrace-dsn")
	if len(dsn) == 0 {
		return nil, errors.New("uptrace-dsn header is required")
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attribute.String("dsn", dsn[0]))
	}

	project, err := org.SelectProjectByDSN(ctx, s.App, dsn[0])
	if err != nil {
		return nil, err
	}

	s.process(ctx, project, req.ResourceLogs)

	return &collectorlogs.ExportLogsServiceResponse{}, nil
}

func (s *LogsServiceServer) process(
	ctx context.Context, project *bunconf.Project, resourceLogs []*logspb.ResourceLogs,
) {
	for _, rl := range resourceLogs {
		resource := AttrMap(otlpconv.Map(rl.Resource.Attributes))

		for _, sl := range rl.ScopeLogs {
			scope := sl.Scope
			resource[attrkey.OtelLibraryName] = scope.Name
			if scope.Version != "" {
				resource[attrkey.OtelLibraryVersion] = scope.Version
			}

			for _, lr := range sl.LogRecords {
				span := s.convLog(resource, lr)
				span.ProjectID = project.ID
				s.sp.AddSpan(ctx, span)
			}
		}
	}
}

func (s *LogsServiceServer) convLog(resource AttrMap, lr *logspb.LogRecord) *Span {
	span := new(Span)

	span.ID = rand.Uint64()
	span.ParentID = otlpSpanID(lr.SpanId)
	span.TraceID = otlpTraceID(lr.TraceId)

	span.EventName = LogEventType
	span.Kind = InternalSpanKind
	span.Time = time.Unix(0, int64(lr.TimeUnixNano))

	span.Attrs = make(AttrMap, len(resource)+len(lr.Attributes)+1)
	span.Attrs.Merge(resource)
	otlpconv.ForEachKeyValue(lr.Attributes, func(key string, value any) {
		span.Attrs[key] = value
	})

	if str, _ := span.Attrs[attrkey.LogMessage].(string); str == "" {
		span.Attrs[attrkey.LogMessage] = s.logMessage(span, lr)
	}
	if lr.SeverityText != "" {
		span.Attrs[attrkey.LogSeverity] = lr.SeverityText
	}

	return span
}

func (s *LogsServiceServer) logMessage(span *Span, lr *logspb.LogRecord) string {
	switch v := lr.Body.Value.(type) {
	case nil:
		// skip
	case *commonpb.AnyValue_StringValue:
		return v.StringValue
	default:
		bodyType := reflect.TypeOf(lr.Body.Value).String()
		s.Logger.Info("unsupported body type", zap.String("type", bodyType))
	}

	str, _ := span.Attrs["{OriginalFormat}"].(string)
	return str
}
