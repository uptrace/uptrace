package tracing

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/xattr"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/zap"
	"go4.org/syncutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TraceServiceServer struct {
	collectortrace.UnimplementedTraceServiceServer

	*bunapp.App

	batchSize int
	ch        chan otlpSpan
	gate      *syncutil.Gate
}

type otlpSpan struct {
	project *bunapp.Project
	*tracepb.Span
	resource AttrMap
}

var _ collectortrace.TraceServiceServer = (*TraceServiceServer)(nil)

func NewTraceServiceServer(app *bunapp.App) *TraceServiceServer {
	batchSize := scaleWithCPU(2000, 32000)
	s := &TraceServiceServer{
		App: app,

		batchSize: batchSize,
		ch:        make(chan otlpSpan, batchSize),
		gate:      syncutil.NewGate(runtime.GOMAXPROCS(0)),
	}

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()

		s.processLoop(app.Context())
	}()

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

	project, err := s.findProjectByDSN(dsn[0])
	if err != nil {
		return nil, err
	}

	s.process(project, req.ResourceSpans)

	return &collectortrace.ExportTraceServiceResponse{}, nil
}

func (s *TraceServiceServer) findProjectByDSN(dsnStr string) (*bunapp.Project, error) {
	dsn, err := org.ParseDSN(dsnStr)
	if err != nil {
		return nil, err
	}

	if dsn.Token == "" {
		return nil, fmt.Errorf("dsn %q does not contain a token", dsnStr)
	}

	projects := s.Config().Projects
	for i := range projects {
		project := &projects[i]
		if project.Token == dsn.Token {
			return project, nil
		}
	}
	return nil, fmt.Errorf("project with token %q not found", dsn.Token)
}

func (s *TraceServiceServer) process(
	project *bunapp.Project, resourceSpans []*tracepb.ResourceSpans,
) {
	for _, rss := range resourceSpans {
		resource := otlpAttrs(rss.Resource.Attributes)

		for _, ils := range rss.InstrumentationLibrarySpans {
			lib := ils.InstrumentationLibrary
			if lib != nil {
				resource[xattr.OtelLibraryName] = lib.Name
				if lib.Version != "" {
					resource[xattr.OtelLibraryVersion] = lib.Version
				}
			}

			for _, span := range ils.Spans {
				s.ch <- otlpSpan{
					project:  project,
					Span:     span,
					resource: resource,
				}
			}
		}
	}
}

func (s *TraceServiceServer) processLoop(ctx context.Context) {
	const timeout = time.Second

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	spans := make([]otlpSpan, 0, s.batchSize)
	var numSpan int

loop:
	for {
		select {
		case span := <-s.ch:
			spans = append(spans, span)
			numSpan += 1 + len(span.Events)
		case <-timer.C:
			if len(spans) > 0 {
				s.flushSpans(ctx, spans, numSpan)
				spans = make([]otlpSpan, 0, len(spans))
			}
			timer.Reset(timeout)
		case <-s.Done():
			break loop
		}

		if numSpan == s.batchSize {
			s.flushSpans(ctx, spans, numSpan)
			spans = make([]otlpSpan, 0, len(spans))
		}
	}

	if len(spans) > 0 {
		s.flushSpans(ctx, spans, numSpan)
	}
}

func (s *TraceServiceServer) flushSpans(ctx context.Context, otlpSpans []otlpSpan, numSpan int) {
	ctx, span := bunapp.Tracer.Start(ctx, "flush-spans")

	s.WaitGroup().Add(1)
	s.gate.Start()

	go func() {
		defer span.End()
		defer s.gate.Done()
		defer s.WaitGroup().Done()

		spans := make([]Span, 0, numSpan)
		index := make([]SpanIndex, 0, numSpan)
		data := make([]SpanData, 0, numSpan)

		ctx := newSpanContext(ctx)
		for i := range otlpSpans {
			otlpSpan := &otlpSpans[i]

			spans = append(spans, Span{})
			span := &spans[len(spans)-1]

			span.ProjectID = otlpSpan.project.ID
			newSpan(ctx, span, otlpSpan)

			index = append(index, SpanIndex{})
			newSpanIndex(&index[len(index)-1], span)

			data = append(data, SpanData{})
			newSpanData(&data[len(data)-1], span)

			for _, otlpEvent := range otlpSpan.Events {
				spans = append(spans, Span{})
				eventSpan := &spans[len(spans)-1]
				newSpanFromEvent(ctx, eventSpan, span, otlpEvent)

				index = append(index, SpanIndex{})
				newSpanIndex(&index[len(index)-1], eventSpan)

				data = append(data, SpanData{})
				newSpanData(&data[len(data)-1], eventSpan)
			}
		}

		if _, err := s.CH().NewInsert().Model(&data).Exec(ctx); err != nil {
			s.Zap(ctx).Error("ch.Insert failed",
				zap.Error(err), zap.String("table", "spans_data"))
		}

		if _, err := s.CH().NewInsert().Model(&index).Exec(ctx); err != nil {
			s.Zap(ctx).Error("ch.Insert failed",
				zap.Error(err), zap.String("table", "spans_index"))
		}
	}()
}
