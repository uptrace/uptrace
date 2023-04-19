package tracing

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/otlpconv"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	collectorlogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type LogsServiceServer struct {
	collectorlogspb.UnimplementedLogsServiceServer

	*bunapp.App

	sp *SpanProcessor
}

var _ collectorlogspb.LogsServiceServer = (*LogsServiceServer)(nil)

func NewLogsServiceServer(app *bunapp.App, sp *SpanProcessor) *LogsServiceServer {
	return &LogsServiceServer{
		App: app,
		sp:  sp,
	}
}

func (s *LogsServiceServer) ExportHTTP(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn := req.Header.Get("uptrace-dsn")
	if dsn == "" {
		return errors.New("uptrace-dsn header is empty or missing")
	}

	project, err := org.SelectProjectByDSN(ctx, s.App, dsn)
	if err != nil {
		return err
	}

	switch contentType := req.Header.Get("content-type"); contentType {
	case jsonContentType:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}

		logsReq := new(collectorlogspb.ExportLogsServiceRequest)
		if err := protojson.Unmarshal(body, logsReq); err != nil {
			return err
		}

		resp, err := s.export(ctx, logsReq.ResourceLogs, project)
		if err != nil {
			return err
		}

		b, err := protojson.Marshal(resp)
		if err != nil {
			return err
		}

		if _, err := w.Write(b); err != nil {
			return err
		}

		return nil
	case xprotobufContentType, protobufContentType:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}

		logsReq := new(collectorlogspb.ExportLogsServiceRequest)
		if err := proto.Unmarshal(body, logsReq); err != nil {
			return err
		}

		resp, err := s.export(ctx, logsReq.ResourceLogs, project)
		if err != nil {
			return err
		}

		b, err := proto.Marshal(resp)
		if err != nil {
			return err
		}

		if _, err := w.Write(b); err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("unsupported content type: %q", contentType)
	}
}

func (s *LogsServiceServer) Export(
	ctx context.Context, req *collectorlogspb.ExportLogsServiceRequest,
) (*collectorlogspb.ExportLogsServiceResponse, error) {
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

	return s.export(ctx, req.ResourceLogs, project)
}

func (s *LogsServiceServer) export(
	ctx context.Context, resourceLogs []*logspb.ResourceLogs, project *org.Project,
) (*collectorlogspb.ExportLogsServiceResponse, error) {
	for _, rl := range resourceLogs {
		resource := AttrMap(otlpconv.Map(rl.Resource.Attributes))

		for _, sl := range rl.ScopeLogs {
			var scope AttrMap
			if sl.Scope != nil {
				scope = maps.Clone(resource)
				if sl.Scope.Name != "" {
					scope[attrkey.OtelLibraryName] = sl.Scope.Name
				}
				if sl.Scope.Version != "" {
					scope[attrkey.OtelLibraryVersion] = sl.Scope.Version
				}
				otlpconv.ForEachKeyValue(sl.Scope.Attributes, func(key string, value any) {
					scope[key] = value
				})

			} else {
				scope = resource
			}

			for _, lr := range sl.LogRecords {
				span := s.convLog(scope, lr)
				span.ProjectID = project.ID
				s.sp.AddSpan(ctx, span)
			}
		}
	}
	return &collectorlogspb.ExportLogsServiceResponse{}, nil
}

func (s *LogsServiceServer) convLog(resource AttrMap, lr *logspb.LogRecord) *Span {
	span := new(Span)

	span.ID = rand.Uint64()
	span.ParentID = otlpSpanID(lr.SpanId)
	if lr.TraceId != nil {
		span.TraceID = otlpTraceID(lr.TraceId)
	}

	span.EventName = otelEventLog
	span.Kind = InternalSpanKind
	span.Time = time.Unix(0, int64(lr.TimeUnixNano))

	span.Attrs = make(AttrMap, len(resource)+len(lr.Attributes)+1)
	span.Attrs.Merge(resource)
	otlpconv.ForEachKeyValue(lr.Attributes, func(key string, value any) {
		span.Attrs[key] = value
	})

	if lr.SeverityText != "" {
		span.Attrs[attrkey.LogSeverity] = lr.SeverityText
	}
	if lr.Body.Value != nil {
		s.processLogRecordBody(span, lr.Body.Value)
	}

	if !span.Attrs.Has(attrkey.LogMessage) {
		if msg := popLogMessageParam(span.Attrs); msg != "" {
			span.Attrs[attrkey.LogMessage] = msg
		}
	}

	return span
}

func (s *LogsServiceServer) processLogRecordBody(span *Span, bodyValue any) {
	switch v := bodyValue.(type) {
	case *commonpb.AnyValue_StringValue:
		span.Attrs[attrkey.LogMessage] = v.StringValue
	case *commonpb.AnyValue_KvlistValue:
		params := otlpconv.Map(v.KvlistValue.Values)
		populateSpanFromParams(span, params)
	}
}
