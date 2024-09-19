package tracing

import (
	"context"
	"runtime"
	"time"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

type SpanProcessor struct {
	*ProcessorThread[Span, SpanIndex]
	sgp *ServiceGraphProcessor
}

func NewSpanProcessor(app *bunapp.App) *SpanProcessor {
	conf := app.Config()

	processor := NewProcessor[Span](
		app,
		conf.Spans.BatchSize,
		conf.Spans.BufferSize,
	)
	thread := NewProcessorThread[Span, SpanIndex](processor)

	p := &SpanProcessor{
		ProcessorThread: thread,
		sgp:             NewServiceGraphProcessor(app),
	}

	if !conf.ServiceGraph.Disabled {
		p.sgp = NewServiceGraphProcessor(app)
	}

	p.logger.Info("starting processing spans...",
		zap.Int("threads", runtime.GOMAXPROCS(0)),
		zap.Int("batch_size", conf.Spans.BatchSize),
		zap.Int("buffer_size", conf.Spans.BufferSize))

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()
		p.processLoop(app.Context())
	}()

	return p
}
func (p *SpanProcessor) AddSpan(ctx context.Context, span *Span) {
	select {
	case p.queue <- span:
		p.logger.Debug("span added")
	default:
		p.processItems(ctx, []*Span{span})
		p.AddItem(ctx, span)
		p.logger.Error("Span buffer is full (consider increasing spans.buffer_size)", zap.Int("len", len(p.queue)))
		spanCounter.Add(
			ctx,
			1,
			metric.WithAttributes(
				bunotel.ProjectIDAttr(span.ProjectID),
				attribute.String("type", "dropped"),
			),
		)
	}
}

func (p *SpanProcessor) processItems(ctx context.Context, spans []*Span) {
	p.logger.Info("processItems called")
	p.logger.Info("Starting processing spans", zap.Int("batch_size", len(spans)))

	ctx, span := bunotel.Tracer.Start(ctx, "process-spans")

	p.App.WaitGroup().Add(1)
	p.gate.Start()

	go func() {
		defer span.End()
		defer p.gate.Done()
		defer p.App.WaitGroup().Done()

		thread := newSpanProcessorThread(p)
		thread._processSpans(ctx, spans)
	}()
}

func (p *spanProcessorThread) _processSpans(ctx context.Context, spans []*Span) {
	indexedSpans := make([]SpanIndex, 0, len(spans))
	dataSpans := make([]SpanData, 0, len(spans))

	seenErrors := make(map[uint64]bool) // basic deduplication

	for _, span := range spans {
		p.initSpanOrEvent(ctx, span)
		spanCounter.Add(
			ctx,
			1,
			metric.WithAttributes(
				bunotel.ProjectIDAttr(span.ProjectID),
				attribute.String("type", "inserted"),
			),
		)

		indexedSpans = append(indexedSpans, SpanIndex{})
		index := &indexedSpans[len(indexedSpans)-1]
		initSpanIndex(index, span)

		if p.sgp != nil {
			if err := p.sgp.ProcessSpan(ctx, index); err != nil {
				p.App.Logger.Error("service graph failed", zap.Error(err))
			}
		}

		if span.EventName != "" {
			dataSpans = append(dataSpans, SpanData{})
			initSpanData(&dataSpans[len(dataSpans)-1], span)
			continue
		}

		var errorCount int
		var logCount int

		for _, event := range span.Events {
			eventSpan := &Span{
				Attrs: NewAttrMap(),
			}
			initEventFromHostSpan(eventSpan, event, span)
			p.initEvent(ctx, eventSpan)

			spanCounter.Add(
				ctx,
				1,
				metric.WithAttributes(
					bunotel.ProjectIDAttr(span.ProjectID),
					attribute.String("type", "inserted"),
				),
			)

			indexedSpans = append(indexedSpans, SpanIndex{})
			initSpanIndex(&indexedSpans[len(indexedSpans)-1], eventSpan)

			dataSpans = append(dataSpans, SpanData{})
			initSpanData(&dataSpans[len(dataSpans)-1], eventSpan)

			if isErrorSystem(eventSpan.System) {
				errorCount++
				if !seenErrors[eventSpan.GroupID] {
					seenErrors[eventSpan.GroupID] = true
					scheduleCreateErrorAlert(ctx, p.App, eventSpan)
				}
			}
			if isLogSystem(eventSpan.System) {
				logCount++
			}
		}

		index.LinkCount = uint8(len(span.Links))
		index.EventCount = uint8(len(span.Events))
		index.EventErrorCount = uint8(errorCount)
		span.Events = nil

		dataSpans = append(dataSpans, SpanData{})
		initSpanData(&dataSpans[len(dataSpans)-1], span)
	}

	if _, err := p.App.CH.NewInsert().
		Model(&dataSpans).
		Exec(ctx); err != nil {
		p.App.Logger.Error("ch.Insert failed",
			zap.Error(err),
			zap.String("table", "spans_data"))
	}

	if _, err := p.App.CH.NewInsert().
		Model(&indexedSpans).
		Exec(ctx); err != nil {
		p.App.Logger.Error("ch.Insert failed",
			zap.Error(err),
			zap.String("table", "spans_index"))
	}
}

func scheduleCreateErrorAlert(ctx context.Context, app *bunapp.App, span *Span) {
	job := org.CreateErrorAlertTask.NewJob(
		span.ProjectID,
		span.GroupID,
		span.TraceID,
		span.ID,
	)
	job.OnceInPeriod(15*time.Minute, span.GroupID)
	job.SetDelay(time.Minute)

	if err := app.MainQueue.AddJob(ctx, job); err != nil {
		app.Zap(ctx).Error("MainQueue.Add failed", zap.Error(err))
	}
}

//------------------------------------------------------------------------------

type spanProcessorThread struct {
	*ProcessorThread[Span, SpanProcessor]
	sgp *ServiceGraphProcessor
}

func newSpanProcessorThread(p *SpanProcessor) *spanProcessorThread {
	return &spanProcessorThread{
		ProcessorThread: NewProcessorThread[Span, SpanProcessor](p.Processor),
		sgp:             p.sgp,
	}
}

func (p *spanProcessorThread) forceSpanName(ctx context.Context, span *Span) bool {
	return p.forceName(ctx, span, func(s *Span) map[string]interface{} {
		return s.Attrs
	}, func(s *Span) uint32 {
		return s.ProjectID
	}, func(s *Span) string {
		return s.EventName
	})
}
