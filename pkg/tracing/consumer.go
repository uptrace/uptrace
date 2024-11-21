package tracing

import (
	"context"
	"slices"
	"time"

	"github.com/cespare/xxhash/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"go4.org/syncutil"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
)

type IndexType interface {
	SpanIndex | LogIndex
}

type DataType interface {
	SpanData | LogData
}

type consumer[IT IndexType, DT DataType] interface {
	indexFromSpan(*IT, *Span)
	dataFromSpan(*DT, *Span)
	updateIndexStats(*IT, uint8, uint8, uint8, uint8)
	postprocessIndex(context.Context, *IT)
}

type Consumer[IT IndexType, DT DataType] struct {
	*bunapp.App

	batchSize int
	queue     chan *Span
	gate      *syncutil.Gate

	c consumer[IT, DT]

	logger *otelzap.Logger
}

func NewConsumer[IT IndexType, DT DataType](
	app *bunapp.App,
	batchSize, bufferSize int,
	gate *syncutil.Gate,
	c consumer[IT, DT],
) *Consumer[IT, DT] {
	return &Consumer[IT, DT]{
		App:       app,
		batchSize: batchSize,
		queue:     make(chan *Span, bufferSize),
		gate:      gate,
		c:         c,
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
	p.gate.Start()

	spans := make([]*Span, len(src))
	copy(spans, src)

	go func() {
		defer span.End()
		defer p.gate.Done()
		defer p.WaitGroup().Done()

		thread := newConsumerThread[IT, DT](p.c, p.App)
		thread._processSpans(ctx, spans)
	}()
}

type consumerThread[IT IndexType, DT DataType] struct {
	consumer[IT, DT]

	*bunapp.App
	projects map[uint32]*org.Project
	digest   *xxhash.Digest
}

func newConsumerThread[IT IndexType, DT DataType](c consumer[IT, DT], app *bunapp.App) *consumerThread[IT, DT] {
	return &consumerThread[IT, DT]{
		App:      app,
		consumer: c,
		projects: make(map[uint32]*org.Project),
		digest:   xxhash.New(),
	}
}

func (p *consumerThread[IT, DT]) appendIndexed(indexed []IT, span *Span) []IT {
	var iv IT
	indexed = append(indexed, iv)
	item := &indexed[len(indexed)-1]
	p.indexFromSpan(item, span)
	return indexed
}

func (p *consumerThread[IT, DT]) appendData(data []DT, span *Span) []DT {
	var dv DT
	data = append(data, dv)
	item := &data[len(data)-1]
	p.dataFromSpan(item, span)
	return data
}

func (p *consumerThread[IT, DT]) _processSpans(ctx context.Context, spans []*Span) {
	indexedSpans := make([]IT, 0, len(spans))
	dataSpans := make([]DT, 0, len(spans))

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

		indexedSpans = p.appendIndexed(indexedSpans, span)
		index := &indexedSpans[len(indexedSpans)-1]

		if span.IsEvent() || span.IsLog() {
			dataSpans = p.appendData(dataSpans, span)
			continue
		}

		p.postprocessIndex(ctx, index)

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

			indexedSpans = p.appendIndexed(indexedSpans, eventSpan)
			dataSpans = p.appendData(dataSpans, eventSpan)

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

		p.updateIndexStats(
			index,
			uint8(len(span.Links)),
			uint8(len(span.Events)),
			uint8(errorCount),
			uint8(logCount),
		)
		span.Events = nil

		dataSpans = p.appendData(dataSpans, span)
	}

	query := p.CH.NewInsert().Model(&dataSpans)
	if _, err := query.Exec(ctx); err != nil {
		p.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err),
			zap.String("table", query.GetTableName()))
	}

	query = p.CH.NewInsert().Model(&indexedSpans)
	if _, err := query.Exec(ctx); err != nil {
		p.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err),
			zap.String("table", query.GetTableName()))
	}
}

func (p *consumerThread[IT, DT]) project(ctx context.Context, projectID uint32) (*org.Project, bool) {
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

func (p *consumerThread[IT, DT]) forceSpanName(ctx context.Context, span *Span) bool {
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
