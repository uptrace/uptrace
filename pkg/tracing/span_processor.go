package tracing

import (
	"context"
	"runtime"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/zap"
	"go4.org/syncutil"
)

type SpanProcessor struct {
	*bunapp.App

	batchSize int
	ch        chan OTLPSpan
	gate      *syncutil.Gate

	logger *otelzap.Logger
}

type OTLPSpan struct {
	project *bunapp.Project
	*tracepb.Span
	resource AttrMap
}

func NewSpanProcessor(app *bunapp.App) *SpanProcessor {
	batchSize := scaleWithCPU(1000, 32000)
	p := &SpanProcessor{
		App: app,

		batchSize: batchSize,
		ch:        make(chan OTLPSpan, runtime.GOMAXPROCS(0)*batchSize),
		gate:      syncutil.NewGate(runtime.GOMAXPROCS(0)),

		logger: app.ZapLogger(),
	}

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()

		p.processLoop(app.Context())
	}()

	return p
}

func (s *SpanProcessor) AddSpan(span OTLPSpan) {
	select {
	case s.ch <- span:
	default:
		s.logger.Error("span buffer is full (span is dropped)")
	}
}

func (s *SpanProcessor) processLoop(ctx context.Context) {
	const timeout = time.Second

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	spans := make([]OTLPSpan, 0, s.batchSize)
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
				spans = make([]OTLPSpan, 0, len(spans))
			}
			timer.Reset(timeout)
		case <-s.Done():
			break loop
		}

		if numSpan == s.batchSize {
			s.flushSpans(ctx, spans, numSpan)
			spans = make([]OTLPSpan, 0, len(spans))
		}
	}

	if len(spans) > 0 {
		s.flushSpans(ctx, spans, numSpan)
	}
}

func (s *SpanProcessor) flushSpans(ctx context.Context, OTLPSpans []OTLPSpan, numSpan int) {
	ctx, span := bunapp.Tracer.Start(ctx, "flush-spans")

	s.WaitGroup().Add(1)
	s.gate.Start()

	go func() {
		defer span.End()
		defer s.gate.Done()
		defer s.WaitGroup().Done()

		s._flushSpans(ctx, OTLPSpans, numSpan)
	}()
}

func (s *SpanProcessor) _flushSpans(ctx context.Context, OTLPSpans []OTLPSpan, numSpan int) {
	spans := make([]Span, 0, numSpan)
	indexedSpans := make([]SpanIndex, 0, numSpan)
	dataSpans := make([]SpanData, 0, numSpan)

	spanCtx := newSpanContext(ctx)
	for i := range OTLPSpans {
		OTLPSpan := &OTLPSpans[i]

		spans = append(spans, Span{})
		span := &spans[len(spans)-1]

		span.ProjectID = OTLPSpan.project.ID
		newSpan(spanCtx, span, OTLPSpan)

		indexedSpans = append(indexedSpans, SpanIndex{})
		index := &indexedSpans[len(indexedSpans)-1]
		newSpanIndex(index, span)

		dataSpans = append(dataSpans, SpanData{})
		newSpanData(&dataSpans[len(dataSpans)-1], span)

		var errorCount int
		var logCount int

		for _, otlpEvent := range OTLPSpan.Events {
			spans = append(spans, Span{})
			eventSpan := &spans[len(spans)-1]
			newSpanFromEvent(spanCtx, eventSpan, span, otlpEvent)

			indexedSpans = append(indexedSpans, SpanIndex{})
			newSpanIndex(&indexedSpans[len(indexedSpans)-1], eventSpan)

			dataSpans = append(dataSpans, SpanData{})
			newSpanData(&dataSpans[len(dataSpans)-1], eventSpan)

			if isErrorSystem(eventSpan.System) {
				errorCount++
			}
			if isLogSystem(eventSpan.System) {
				logCount++
			}
		}

		index.LinkCount = uint8(len(OTLPSpan.Links))
		index.EventCount = uint8(len(OTLPSpan.Events))
		index.EventErrorCount = uint8(errorCount)
		index.EventLogCount = uint8(logCount)
	}

	if _, err := s.CH().NewInsert().Model(&dataSpans).Exec(ctx); err != nil {
		s.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err), zap.String("table", "spans_data"))
	}

	if _, err := s.CH().NewInsert().Model(&indexedSpans).Exec(ctx); err != nil {
		s.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err), zap.String("table", "spans_index"))
	}
}
