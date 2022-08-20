package tracing

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/otlpconv"
	"github.com/uptrace/uptrace/pkg/tracing/attrkey"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type TraceServiceServer struct {
	collectortrace.UnimplementedTraceServiceServer

	*bunapp.App

	sp *SpanProcessor
}

var _ collectortrace.TraceServiceServer = (*TraceServiceServer)(nil)

func NewTraceServiceServer(app *bunapp.App, sp *SpanProcessor) *TraceServiceServer {
	s := &TraceServiceServer{
		App: app,
		sp:  sp,
	}
	return s
}

func (s *TraceServiceServer) Export(
	ctx context.Context, req *collectortrace.ExportTraceServiceRequest,
) (*collectortrace.ExportTraceServiceResponse, error) {
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

	s.process(ctx, project, req.ResourceSpans)

	return &collectortrace.ExportTraceServiceResponse{}, nil
}

func (s *TraceServiceServer) process(
	ctx context.Context, project *bunconf.Project, resourceSpans []*tracepb.ResourceSpans,
) {
	for _, rs := range resourceSpans {
		if len(rs.ScopeSpans) == 0 {
			for _, ils := range rs.InstrumentationLibrarySpans {
				scopeSpans := tracepb.ScopeSpans{
					Scope: &commonpb.InstrumentationScope{
						Name:    ils.InstrumentationLibrary.Name,
						Version: ils.InstrumentationLibrary.Version,
					},
					Spans:     ils.Spans,
					SchemaUrl: ils.SchemaUrl,
				}
				rs.ScopeSpans = append(rs.ScopeSpans, &scopeSpans)
			}
		}
		rs.InstrumentationLibrarySpans = nil
	}

	for _, rss := range resourceSpans {
		resource := AttrMap(otlpconv.Map(rss.Resource.Attributes))

		for _, ss := range rss.ScopeSpans {
			lib := ss.Scope
			if lib != nil {
				resource[attrkey.OtelLibraryName] = lib.Name
				if lib.Version != "" {
					resource[attrkey.OtelLibraryVersion] = lib.Version
				}
			}

			for _, otlpSpan := range ss.Spans {
				// TODO(vmihailenco): allocate spans in batches
				span := new(Span)
				initSpanFromOTLP(span, resource, otlpSpan)
				span.ProjectID = project.ID
				s.sp.AddSpan(ctx, span)
			}
		}
	}
}

func (s *TraceServiceServer) httpTraces(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn := req.Header.Get("uptrace-dsn")
	if dsn == "" {
		return errors.New("uptrace-dsn header is required")
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attribute.String("dsn", dsn))
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

		td := new(tracepb.TracesData)
		if err := protojson.Unmarshal(body, td); err != nil {
			return err
		}

		s.process(ctx, project, td.ResourceSpans)

		resp := new(collectortrace.ExportTraceServiceResponse)
		b, err := protojson.Marshal(resp)
		if err != nil {
			return err
		}

		if _, err := w.Write(b); err != nil {
			return err
		}

		return nil
	case protobufContentType:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}

		td := new(collectortrace.ExportTraceServiceRequest)
		if err := proto.Unmarshal(body, td); err != nil {
			return err
		}

		s.process(ctx, project, td.ResourceSpans)

		resp := new(collectortrace.ExportTraceServiceResponse)
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
