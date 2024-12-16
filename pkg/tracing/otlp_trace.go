package tracing

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/fx"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/otlpconv"
)

type TraceServiceServerParams struct {
	fx.In

	Logger   *otelzap.Logger
	PG       *bun.DB
	PS       *org.ProjectGateway
	Consumer *SpanConsumer
}

type TraceServiceServer struct {
	collectortrace.UnimplementedTraceServiceServer
	*TraceServiceServerParams
}

var _ collectortrace.TraceServiceServer = (*TraceServiceServer)(nil)

func NewTraceServiceServer(p TraceServiceServerParams) *TraceServiceServer {
	return &TraceServiceServer{
		TraceServiceServerParams: &p,
	}
}

func (s *TraceServiceServer) ExportHTTP(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn, err := org.DSNFromRequest(req)
	if err != nil {
		return err
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attribute.String("dsn", dsn))
	}

	project, err := s.PS.SelectByDSN(ctx, dsn)
	if err != nil {
		return err
	}

	switch contentType := req.Header.Get("content-type"); contentType {
	case jsonContentType:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}

		traceReq := new(collectortrace.ExportTraceServiceRequest)
		if err := protojson.Unmarshal(body, traceReq); err != nil {
			return err
		}

		resp, err := s.process(ctx, project, traceReq.ResourceSpans)
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

		traceReq := new(collectortrace.ExportTraceServiceRequest)
		if err := proto.Unmarshal(body, traceReq); err != nil {
			return err
		}

		resp, err := s.process(ctx, project, traceReq.ResourceSpans)
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

func (s *TraceServiceServer) Export(
	ctx context.Context, req *collectortrace.ExportTraceServiceRequest,
) (*collectortrace.ExportTraceServiceResponse, error) {
	if ctx.Err() == context.Canceled {
		return nil, status.Error(codes.Canceled, "Client cancelled, abandoning.")
	}

	dsn, err := org.DSNFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	project, err := s.PS.SelectByDSN(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return s.process(ctx, project, req.ResourceSpans)
}

func (s *TraceServiceServer) process(
	ctx context.Context, project *org.Project, resourceSpans []*tracepb.ResourceSpans,
) (*collectortrace.ExportTraceServiceResponse, error) {
	for _, rss := range resourceSpans {
		var resource AttrMap
		if rss.Resource != nil {
			resource = AttrMap(otlpconv.Map(rss.Resource.Attributes))
		}

		for _, ss := range rss.ScopeSpans {
			var scope AttrMap
			if ss.Scope != nil {
				scope = maps.Clone(resource)
				if ss.Scope.Name != "" {
					scope[attrkey.OtelLibraryName] = ss.Scope.Name
				}
				if ss.Scope.Version != "" {
					scope[attrkey.OtelLibraryVersion] = ss.Scope.Version
				}
				otlpconv.ForEachKeyValue(ss.Scope.Attributes, func(key string, value any) {
					scope[key] = value
				})
			} else {
				scope = resource
			}

			mem := make([]Span, len(ss.Spans))
			for i, otlpSpan := range ss.Spans {
				span := &mem[i]
				initSpanFromOTLP(span, scope, otlpSpan)
				span.ProjectID = project.ID
				s.Consumer.AddSpan(ctx, span)
			}
		}
	}

	org.CreateAchievementOnce(ctx, s.Logger, s.PG, &org.Achievement{
		ProjectID: project.ID,
		Name:      org.AchievConfigureTracing,
	})

	return &collectortrace.ExportTraceServiceResponse{}, nil
}
