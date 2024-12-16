package tracing

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/vmihailenco/taskq/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
)

const maxWorkers = 10

type IndexRecord interface {
	SpanIndex | LogIndex | EventIndex
}

type DataRecord interface {
	SpanData | LogData | EventData
}

type transformer[IT IndexRecord, DT DataRecord] interface {
	initIndexFromSpan(*IT, *Span)
	initDataFromSpan(*DT, *Span)
	postprocessIndex(context.Context, *IT)
}

type BaseConsumer[IT IndexRecord, DT DataRecord] struct {
	logger      *otelzap.Logger
	pg          *bun.DB
	ch          *ch.DB
	ps          *org.ProjectGateway
	mainQueue   taskq.Queue
	batchSize   int
	transformer transformer[IT, DT]

	cancel      context.CancelFunc
	wg          sync.WaitGroup
	queue       chan *Span
	workerPool  chan *consumerWorker[IT, DT]
	workerCount int
}

type BaseConsumerParams struct {
	fx.In

	Logger    *otelzap.Logger
	Conf      *bunconf.Config
	PG        *bun.DB
	CH        *ch.DB
	PS        *org.ProjectGateway
	MainQueue taskq.Queue
}

func NewBaseConsumer[IT IndexRecord, DT DataRecord](
	logger *otelzap.Logger,
	pg *bun.DB,
	ch *ch.DB,
	ps *org.ProjectGateway,
	mainQueue taskq.Queue,
	signalName string,
	batchSize, bufferSize, maxWorkers int,
	transformer transformer[IT, DT],
) *BaseConsumer[IT, DT] {
	c := &BaseConsumer[IT, DT]{
		logger:      logger,
		pg:          pg,
		ch:          ch,
		ps:          ps,
		mainQueue:   mainQueue,
		batchSize:   batchSize,
		queue:       make(chan *Span, bufferSize),
		transformer: transformer,
		workerPool:  make(chan *consumerWorker[IT, DT], maxWorkers),
		workerCount: 0,
	}

	queueLen, _ := bunotel.Meter.Int64ObservableGauge(signalName, metric.WithUnit("{spans}"))
	if _, err := bunotel.Meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			o.ObserveInt64(queueLen, int64(len(c.queue)))
			return nil
		},
		queueLen,
	); err != nil {
		panic(err)
	}

	return c
}

func (p *BaseConsumer[IT, DT]) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	p.processLoop(ctx)
}

func (p *BaseConsumer[IT, DT]) Stop() {
	if p.cancel == nil {
		p.logger.Error("no cancel function registered for BaseConsumer")
		return
	}

	p.cancel()
	p.wg.Wait()
}

func (p *BaseConsumer[IT, DT]) AddSpan(ctx context.Context, span *Span) {
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

func (p *BaseConsumer[IT, DT]) processLoop(ctx context.Context) {
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
		case <-ctx.Done():
			break loop
		}
	}

	if len(spans) > 0 {
		p.processSpans(ctx, spans)
	}
}

func (p *BaseConsumer[IT, DT]) processSpans(ctx context.Context, src []*Span) {
	ctx, span := bunotel.Tracer.Start(ctx, "process-spans")

	var worker *consumerWorker[IT, DT]

	select {
	case worker = <-p.workerPool:
	default:
		if p.workerCount < maxWorkers {
			p.workerCount++
			worker = newConsumerWorker(
				p.logger,
				p.pg, p.ch, p.ps,
				p.transformer,
				cap(p.queue),
				p.spanErrorHandler,
			)
		}
	}

	spans := make([]*Span, len(src))
	copy(spans, src)

	p.wg.Add(1)
	go func(worker *consumerWorker[IT, DT]) {
		defer span.End()
		defer p.wg.Done()

		if worker == nil {
			worker = <-p.workerPool
		}
		worker._processSpans(ctx, spans)
		p.workerPool <- worker
	}(worker)
}

func (p *BaseConsumer[IT, DT]) spanErrorHandler(ctx context.Context, span *Span) {
	job := org.CreateErrorAlertTask.NewJob(
		span.ProjectID,
		span.GroupID,
		span.TraceID,
		span.ID,
	)
	job.OnceInPeriod(15*time.Minute, span.GroupID)
	job.SetDelay(time.Minute)

	if err := p.mainQueue.AddJob(ctx, job); err != nil {
		p.logger.Error("MainQueue.Add failed", zap.Error(err))
	}
}

type consumerWorker[IT IndexRecord, DT DataRecord] struct {
	logger           *otelzap.Logger
	pg               *bun.DB
	ch               *ch.DB
	ps               *org.ProjectGateway
	transformer      transformer[IT, DT]
	spanErrorHandler func(context.Context, *Span)

	projects     map[uint32]*org.Project
	digest       *xxhash.Digest
	indexedSpans []IT
	dataSpans    []DT
}

func newConsumerWorker[IT IndexRecord, DT DataRecord](
	logger *otelzap.Logger,
	pg *bun.DB,
	ch *ch.DB,
	ps *org.ProjectGateway,
	transformer transformer[IT, DT],
	bufSize int,
	spanErrorHandler func(context.Context, *Span),
) *consumerWorker[IT, DT] {
	return &consumerWorker[IT, DT]{
		logger:           logger,
		pg:               pg,
		ch:               ch,
		ps:               ps,
		transformer:      transformer,
		spanErrorHandler: spanErrorHandler,
		projects:         make(map[uint32]*org.Project),
		digest:           xxhash.New(),
		indexedSpans:     make([]IT, 0, bufSize),
		dataSpans:        make([]DT, 0, bufSize),
	}
}

func (p *consumerWorker[IT, DT]) appendIndexed(span *Span) *IT {
	var item IT
	p.indexedSpans = append(p.indexedSpans, item)
	index := &p.indexedSpans[len(p.indexedSpans)-1]
	p.transformer.initIndexFromSpan(index, span)
	return index
}

func (p *consumerWorker[IT, DT]) appendData(span *Span) *DT {
	var item DT
	p.dataSpans = append(p.dataSpans, item)
	data := &p.dataSpans[len(p.dataSpans)-1]
	p.transformer.initDataFromSpan(data, span)
	return data
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

		index := p.appendIndexed(span)
		p.transformer.postprocessIndex(ctx, index)

		if span.IsEvent() || span.IsLog() {
			_ = p.appendData(span)
			continue
		}

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

			_ = p.appendIndexed(eventSpan)
			_ = p.appendData(eventSpan)

			if isErrorSystem(eventSpan.System) && !seenErrors[eventSpan.GroupID] {
				seenErrors[eventSpan.GroupID] = true
				p.spanErrorHandler(ctx, eventSpan)
			}
		}

		span.Events = nil

		_ = p.appendData(span)
	}

	query := p.ch.NewInsert().Model(&p.dataSpans)
	if _, err := query.Exec(ctx); err != nil {
		p.logger.Error("ch.Insert failed",
			zap.Error(err),
			zap.String("table", query.GetTableName()))
	}

	query = p.ch.NewInsert().Model(&p.indexedSpans)
	if _, err := query.Exec(ctx); err != nil {
		p.logger.Error("ch.Insert failed",
			zap.Error(err),
			zap.String("table", query.GetTableName()))
	}
}

func (p *consumerWorker[IT, DT]) project(ctx context.Context, projectID uint32) (*org.Project, bool) {
	if project, ok := p.projects[projectID]; ok {
		return project, true
	}

	project, err := p.ps.SelectByID(ctx, projectID)
	if err != nil {
		p.logger.Error("SelectProject failed", zap.Error(err))
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
