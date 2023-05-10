package tracing

import (
	"context"
	"runtime"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.uber.org/zap"
	"go4.org/syncutil"
)

type SpanProcessor struct {
	*bunapp.App

	batchSize int
	queue     chan *Span
	gate      *syncutil.Gate

	logger *otelzap.Logger
}

func NewSpanProcessor(app *bunapp.App) *SpanProcessor {
	conf := app.Config()
	maxprocs := runtime.GOMAXPROCS(0)
	p := &SpanProcessor{
		App: app,

		batchSize: conf.Spans.BatchSize,
		queue:     make(chan *Span, conf.Spans.BufferSize),
		gate:      syncutil.NewGate(maxprocs),

		logger: app.Logger,
	}

	p.logger.Info("starting processing spans...",
		zap.Int("threads", maxprocs),
		zap.Int("batch_size", p.batchSize),
		zap.Int("buffer_size", conf.Spans.BufferSize))

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()

		p.processLoop(app.Context())
	}()

	queueLen, _ := bunotel.Meter.Int64ObservableGauge("uptrace.tracing.queue_length",
		instrument.WithUnit("{spans}"),
	)

	if _, err := bunotel.Meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			o.ObserveInt64(queueLen, int64(len(p.queue)))
			return nil
		},
		queueLen,
	); err != nil {
		panic(err)
	}

	return p
}

func (p *SpanProcessor) AddSpan(ctx context.Context, span *Span) {
	select {
	case p.queue <- span:
	default:
		p.logger.Error("span buffer is full (consider increasing spans.buffer_size)",
			zap.Int("len", len(p.queue)))
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

func (p *SpanProcessor) processLoop(ctx context.Context) {
	const timeout = 5 * time.Second

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	spans := make([]*Span, 0, p.batchSize)

loop:
	for {
		select {
		case span := <-p.queue:
			spans = append(spans, span)

			if len(spans) < p.batchSize {
				break
			}

			p.processSpans(ctx, spans)
			spans = spans[:0]

			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(timeout)
		case <-timer.C:
			if len(spans) > 0 {
				p.processSpans(ctx, spans)
				spans = spans[:0]
			}
			timer.Reset(timeout)
		case <-p.Done():
			break loop
		}
	}

	if len(spans) > 0 {
		p.processSpans(ctx, spans)
	}
}

func (p *SpanProcessor) processSpans(ctx context.Context, src []*Span) {
	ctx, span := bunotel.Tracer.Start(ctx, "process-spans")

	p.WaitGroup().Add(1)
	p.gate.Start()

	spans := make([]*Span, len(src))
	copy(spans, src)

	go func() {
		defer span.End()
		defer p.gate.Done()
		defer p.WaitGroup().Done()

		spanCtx := newSpanContext(ctx)
		p._processSpans(spanCtx, spans)
	}()
}

func (s *SpanProcessor) _processSpans(ctx *spanContext, spans []*Span) {
	indexedSpans := make([]SpanIndex, 0, len(spans))
	dataSpans := make([]SpanData, 0, len(spans))

	seenErrors := make(map[uint64]bool) // basic deduplication

	for _, span := range spans {
		initSpanOrEvent(ctx, s.App, span)
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

		dataSpans = append(dataSpans, SpanData{})
		initSpanData(&dataSpans[len(dataSpans)-1], span)

		if span.EventName != "" {
			continue
		}

		var errorCount int
		var logCount int

		for _, event := range span.Events {
			eventSpan := new(Span)
			initEventFromHostSpan(eventSpan, event, span)
			initEvent(ctx, s.App, eventSpan)

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
					scheduleCreateErrorAlert(ctx, s.App, eventSpan)
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
		span.Events = nil
	}

	if _, err := s.CH.NewInsert().
		Model(&dataSpans).
		ModelTableExpr("?", s.DistTable("spans_data_buffer")).
		Exec(ctx); err != nil {
		s.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err), zap.String("table", "spans_data"))
	}

	if _, err := s.CH.NewInsert().
		Model(&indexedSpans).
		ModelTableExpr("?", s.DistTable("spans_index_buffer")).
		Exec(ctx); err != nil {
		s.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err), zap.String("table", "spans_index"))
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

type spanContext struct {
	context.Context

	projects map[uint32]*org.Project
	digest   *xxhash.Digest
}

func newSpanContext(ctx context.Context) *spanContext {
	return &spanContext{
		Context: ctx,

		projects: make(map[uint32]*org.Project),
		digest:   xxhash.New(),
	}
}

func (c *spanContext) Project(app *bunapp.App, projectID uint32) (*org.Project, bool) {
	if p, ok := c.projects[projectID]; ok {
		return p, true
	}

	project, err := org.SelectProject(c.Context, app, projectID)
	if err != nil {
		app.Zap(c.Context).Error("SelectProject failed", zap.Error(err))
		return nil, false
	}

	c.projects[projectID] = project
	return project, true
}
