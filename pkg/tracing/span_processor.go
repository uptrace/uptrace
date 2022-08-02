package tracing

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/prometheus/model/labels"
	promlabels "github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/notifier"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/tracing/xattr"
	"go.opentelemetry.io/otel/attribute"
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

		logger: app.Logger(),
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
		s.logger.Error("span buffer is full (consider increasing spans.buffer_size)")
	}
}

func (s *SpanProcessor) processLoop(ctx context.Context) {
	const timeout = time.Second

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	spans := make([]*Span, 0, s.batchSize)

loop:
	for {
		select {
		case span := <-s.ch:
			spans = append(spans, span)
		case <-timer.C:
			if len(spans) > 0 {
				s.flushSpans(ctx, spans)
				spans = make([]*Span, 0, len(spans))
			}
			timer.Reset(timeout)
		case <-s.Done():
			break loop
		}

		if len(spans) == s.batchSize {
			s.flushSpans(ctx, spans)
			spans = make([]*Span, 0, len(spans))
		}
	}

	if len(spans) > 0 {
		s.flushSpans(ctx, spans)
	}
}

func (s *SpanProcessor) flushSpans(ctx context.Context, spans []*Span) {
	ctx, span := bunapp.Tracer.Start(ctx, "flush-spans")

	s.WaitGroup().Add(1)
	s.gate.Start()

	go func() {
		defer span.End()
		defer s.gate.Done()
		defer s.WaitGroup().Done()

		s._flushSpans(ctx, spans)
	}()
}

func (s *SpanProcessor) _flushSpans(ctx context.Context, spans []*Span) {
	indexedSpans := make([]SpanIndex, 0, len(spans))
	dataSpans := make([]SpanData, 0, len(spans))

	seenErrors := make(map[uint64]bool) // basic deduplication
	var errors []*Span

	spanCtx := newSpanContext(ctx)
	for _, span := range spans {
		initSpan(spanCtx, span)
		spanCounter.Add(
			ctx,
			int64(len(spans)),
			attribute.Int64("project_id", int64(span.ProjectID)),
		)

		indexedSpans = append(indexedSpans, SpanIndex{})
		index := &indexedSpans[len(indexedSpans)-1]
		initSpanIndex(index, span)

		dataSpans = append(dataSpans, SpanData{})
		initSpanData(&dataSpans[len(dataSpans)-1], span)

		var errorCount int
		var logCount int

		for _, eventSpan := range span.Events {
			initSpanEvent(spanCtx, eventSpan, span)
			spanCounter.Add(ctx, int64(len(spans)),
				attribute.Int64("project_id", int64(span.ProjectID)))

			indexedSpans = append(indexedSpans, SpanIndex{})
			initSpanIndex(&indexedSpans[len(indexedSpans)-1], eventSpan)

			dataSpans = append(dataSpans, SpanData{})
			initSpanData(&dataSpans[len(dataSpans)-1], eventSpan)

			if isErrorSystem(eventSpan.System) {
				errorCount++
				if !seenErrors[eventSpan.GroupID] {
					seenErrors[eventSpan.GroupID] = true
					errors = append(errors, eventSpan)
				}
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

	if _, err := s.CH.NewInsert().Model(&dataSpans).Exec(ctx); err != nil {
		s.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err), zap.String("table", "spans_data"))
	}

	if _, err := s.CH.NewInsert().Model(&indexedSpans).Exec(ctx); err != nil {
		s.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err), zap.String("table", "spans_index"))
	}

	if len(errors) > 0 {
		s.notifyOnErrors(errors)
	}
}

func (s *SpanProcessor) notifyOnErrors(errors []*Span) {
	alerts := make([]*notifier.Alert, len(errors))

	for i, error := range errors {
		labels := []string{
			labels.AlertName, error.EventName,
			"alert_kind", "error",
			"project_id", strconv.FormatUint(uint64(error.ProjectID), 10),
			"system", error.System,
			"group_id", strconv.FormatUint(error.GroupID, 10),
		}
		if service := error.Attrs.ServiceName(); service != "" {
			labels = append(labels, "service", service)
		}
		if sev, _ := error.Attrs[xattr.LogSeverity].(string); sev != "" {
			labels = append(labels, "severity", sev)
		}

		alerts[i] = &notifier.Alert{
			Labels: promlabels.FromStrings(labels...),
			Annotations: promlabels.FromStrings(
				"span_name", error.Name,
				"trace_id", error.TraceID.String(),
			),
			GeneratorURL: s.Config().SitePath(fmt.Sprintf("/traces/%s", error.TraceID.String())),
		}
	}

	s.NotifierManager.Send(alerts...)
}
