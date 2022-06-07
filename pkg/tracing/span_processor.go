package tracing

import (
	"context"
	"runtime"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"go.uber.org/zap"
	"go4.org/syncutil"
)

type SpanProcessor struct {
	*bunapp.App

	batchSize int
	ch        chan *Span
	gate      *syncutil.Gate

	logger *otelzap.Logger
}

func NewSpanProcessor(app *bunapp.App) *SpanProcessor {
	cfg := app.Config()
	p := &SpanProcessor{
		App: app,

		batchSize: cfg.Spans.BatchSize,
		ch:        make(chan *Span, cfg.Spans.BufferSize),
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

func (s *SpanProcessor) AddSpan(span *Span) {
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

	spans := make([]*Span, 0, s.batchSize)
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
				spans = make([]*Span, 0, len(spans))
			}
			timer.Reset(timeout)
		case <-s.Done():
			break loop
		}

		if numSpan == s.batchSize {
			s.flushSpans(ctx, spans, numSpan)
			spans = make([]*Span, 0, len(spans))
		}
	}

	if len(spans) > 0 {
		s.flushSpans(ctx, spans, numSpan)
	}
}

func (s *SpanProcessor) flushSpans(ctx context.Context, spans []*Span, numSpan int) {
	ctx, span := bunapp.Tracer.Start(ctx, "flush-spans")

	s.WaitGroup().Add(1)
	s.gate.Start()

	go func() {
		defer span.End()
		defer s.gate.Done()
		defer s.WaitGroup().Done()

		s._flushSpans(ctx, spans, numSpan)
	}()
}

func (s *SpanProcessor) _flushSpans(ctx context.Context, spans []*Span, numSpan int) {
	indexedSpans := make([]SpanIndex, 0, numSpan)
	dataSpans := make([]SpanData, 0, numSpan)

	spanCtx := newSpanContext(ctx)
	for _, span := range spans {
		initSpan(spanCtx, span)

		indexedSpans = append(indexedSpans, SpanIndex{})
		index := &indexedSpans[len(indexedSpans)-1]
		newSpanIndex(index, span)

		dataSpans = append(dataSpans, SpanData{})
		newSpanData(&dataSpans[len(dataSpans)-1], span)

		var errorCount int
		var logCount int

		for _, eventSpan := range span.Events {
			initSpanEvent(spanCtx, eventSpan, span)

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

		index.LinkCount = uint8(len(span.Links))
		index.EventCount = uint8(len(span.Events))
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
