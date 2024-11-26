package tracing

import (
	"context"
	"slices"
	"time"

	"github.com/cespare/xxhash/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
)

type IndexRecord interface {
	SpanIndex | LogIndex
}

type DataRecord interface {
	SpanData | LogData
}

type transformer[IT IndexRecord, DT DataRecord] interface {
	initIndexFromSpan(*IT, *Span)
	initDataFromSpan(*DT, *Span)
	postprocessIndex(context.Context, *IT)
}

type Consumer[IT IndexRecord, DT DataRecord] struct {
	*bunapp.App
	logger *otelzap.Logger

	batchSize   int
	queue       chan *Span
	transformer transformer[IT, DT]
	workerPool  chan *consumerWorker[IT, DT]
	workerCount int
}

const maxWorkers = 10

func NewConsumer[IT IndexRecord, DT DataRecord](
	app *bunapp.App,
	batchSize, bufferSize, maxWorkers int,
	transformer transformer[IT, DT],
) *Consumer[IT, DT] {
	return &Consumer[IT, DT]{
		App:         app,
		batchSize:   batchSize,
		queue:       make(chan *Span, bufferSize),
		transformer: transformer,
		workerPool:  make(chan *consumerWorker[IT, DT], maxWorkers),
		workerCount: 0,
	}
}

func (p *Consumer[IT, DT]) AddSpan(ctx context.Context, span *Span) {
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

func (p *Consumer[IT, DT]) processLoop(ctx context.Context) {
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

func (p *Consumer[IT, DT]) processSpans(ctx context.Context, src []*Span) {
	ctx, span := bunotel.Tracer.Start(ctx, "process-spans")

	p.WaitGroup().Add(1)

	var worker *consumerWorker[IT, DT]

	select {
	case worker = <-p.workerPool:
	default:
		if p.workerCount < maxWorkers {
			worker = newConsumerWorker(p.App, p.transformer, cap(p.queue))
			p.workerCount++
		}
	}

	spans := make([]*Span, len(src))
	copy(spans, src)

	go func(worker *consumerWorker[IT, DT]) {
		defer span.End()
		defer p.WaitGroup().Done()

		if worker == nil {
			worker = <-p.workerPool
		}
		worker._processSpans(ctx, spans)
		p.workerPool <- worker
	}(worker)
}

type consumerWorker[IT IndexRecord, DT DataRecord] struct {
	*bunapp.App

	transformer transformer[IT, DT]
	projects    map[uint32]*org.Project
	digest      *xxhash.Digest

	indexedSpans []IT
	dataSpans    []DT
}

func newConsumerWorker[IT IndexRecord, DT DataRecord](
	app *bunapp.App,
	transformer transformer[IT, DT],
	bufSize int,
) *consumerWorker[IT, DT] {
	return &consumerWorker[IT, DT]{
		App:          app,
		transformer:  transformer,
		projects:     make(map[uint32]*org.Project),
		digest:       xxhash.New(),
		indexedSpans: make([]IT, 0, bufSize),
		dataSpans:    make([]DT, 0, bufSize),
	}
}

func (p *consumerWorker[IT, DT]) _processSpans(ctx context.Context, spans []*Span) {
	seenErrors := make(map[uint64]bool) // basic deduplication

	defer func() {
		clear(p.indexedSpans)
		p.indexedSpans = p.indexedSpans[:0]
		clear(p.dataSpans)
		p.dataSpans = p.dataSpans[:0]
	}()

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

		var index IT
		p.indexedSpans = append(p.indexedSpans, index)
		p.transformer.initIndexFromSpan(&index, span)

		if span.IsEvent() || span.IsLog() {
			var data DT
			p.dataSpans = append(p.dataSpans, data)
			p.transformer.initDataFromSpan(&data, span)
			continue
		}

		p.transformer.postprocessIndex(ctx, &index)

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

			var index IT
			p.indexedSpans = append(p.indexedSpans, index)
			p.transformer.initIndexFromSpan(&index, eventSpan)

			var data DT
			p.dataSpans = append(p.dataSpans, data)
			p.transformer.initDataFromSpan(&data, eventSpan)

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

		span.Events = nil

		var data DT
		p.dataSpans = append(p.dataSpans, data)
		p.transformer.initDataFromSpan(&data, span)
	}

	query := p.CH.NewInsert().Model(&p.dataSpans)
	if _, err := query.Exec(ctx); err != nil {
		p.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err),
			zap.String("table", query.GetTableName()))
	}

	query = p.CH.NewInsert().Model(&p.indexedSpans)
	if _, err := query.Exec(ctx); err != nil {
		p.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err),
			zap.String("table", query.GetTableName()))
	}
}

func (p *consumerWorker[IT, DT]) project(ctx context.Context, projectID uint32) (*org.Project, bool) {
	if project, ok := p.projects[projectID]; ok {
		return project, true
	}

	project, err := org.SelectProject(ctx, p.App, projectID)
	if err != nil {
		p.Zap(ctx).Error("SelectProject failed", zap.Error(err))
		return nil, false
	}

	p.projects[projectID] = project
	return project, true
}

func (p *consumerWorker[IT, DT]) forceSpanName(ctx context.Context, span *Span) bool {
	if span.EventName != "" {
		return false
	}

	project, ok := p.project(ctx, span.ProjectID)
	if !ok {
		return false
	}

	if libName, _ := span.Attrs[attrkey.OtelLibraryName].(string); libName != "" {
		return slices.Contains(project.ForceSpanName, libName)
	}
	return false
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
