package tracing

import (
	"context"
	"runtime"
	"slices"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"go4.org/syncutil"
)

type SpanProcessor struct {
	*bunapp.App

	batchSize int
	queue     chan *Span
	gate      *syncutil.Gate

	sgp *ServiceGraphProcessor

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

	if !conf.ServiceGraph.Disabled {
		p.sgp = NewServiceGraphProcessor(app)
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
		metric.WithUnit("{spans}"),
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

		thread := newSpanProcessorThread(p)
		thread._processSpans(ctx, spans)
	}()
}

func (p *spanProcessorThread) _processSpans(ctx context.Context, spans []*Span) {
	indexedSpans := make([]BaseIndex, 0, len(spans))
	dataSpans := make([]BaseData, 0, len(spans))

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

		indexedSpans = append(indexedSpans, BaseIndex{})
		index := &indexedSpans[len(indexedSpans)-1]
		consumeIndex(index, span)

		if p.sgp != nil {
			if err := p.sgp.ProcessSpan(ctx, index); err != nil {
				p.Zap(ctx).Error("service graph failed", zap.Error(err))
			}
		}

		if span.EventName != "" {
			dataSpans = append(dataSpans, BaseData{})
			consumeData(&dataSpans[len(dataSpans)-1], span)
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

			indexedSpans = append(indexedSpans, BaseIndex{})
			consumeIndex(&indexedSpans[len(indexedSpans)-1], eventSpan)

			dataSpans = append(dataSpans, BaseData{})
			consumeData(&dataSpans[len(dataSpans)-1], eventSpan)

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
		index.EventLogCount = uint8(logCount)
		span.Events = nil

		dataSpans = append(dataSpans, BaseData{})
		consumeData(&dataSpans[len(dataSpans)-1], span)
	}

	p.insertData(ctx, dataSpans)
	p.insertIndex(ctx, indexedSpans)
}

func consumeIndex(index *BaseIndex, span *Span) {
	switch {
	// TODO: case event
	case span.Type == "log":
		b := NewBaseIndexer(NewLogIndex(index))
		b.initIndex(span)
	default:
		b := NewBaseIndexer(NewSpanIndex(index))
		b.initIndex(span)
	}
}

func consumeData(data *BaseData, span *Span) {
	switch {
	// TODO: case event
	case span.Type == "log":
		d := NewBaseDater(NewLogData(data))
		d.initData(span)
	default:
		d := NewBaseDater(NewSpanData(data))
		d.initData(span)
	}
}

func (p *spanProcessorThread) insertCH(ctx context.Context, val any, table string) {
	if _, err := p.CH.NewInsert().
		Model(val).
		Exec(ctx); err != nil {
		p.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err),
			zap.String("table", table))
	}
}

func (p *spanProcessorThread) insertData(ctx context.Context, datas []BaseData) {
	dataSpans := make([]SpanData, 0, len(datas))
	dataLogs := make([]LogData, 0, len(datas))

	for _, d := range datas {
		switch {
		// TODO: case event
		case d.Type == "log":
			dataLogs = append(dataLogs, *NewLogData(&d))
		default:
			dataSpans = append(dataSpans, *NewSpanData(&d))
		}
	}
	p.insertCH(ctx, &dataSpans, new(SpanData).TableName())
	p.insertCH(ctx, &dataLogs, new(LogData).TableName())
}

func (p *spanProcessorThread) insertIndex(ctx context.Context, indexes []BaseIndex) {
	indexedSpans := make([]SpanIndex, 0, len(indexes))
	indexedLogs := make([]LogIndex, 0, len(indexes))

	for _, i := range indexes {
		switch {
		//TODO: case event
		case i.Type == "log":
			indexedLogs = append(indexedLogs, *NewLogIndex(&i))
		default:
			indexedSpans = append(indexedSpans, *NewSpanIndex(&i))
		}
	}
	p.insertCH(ctx, &indexedSpans, new(SpanIndex).TableName())
	p.insertCH(ctx, &indexedLogs, new(LogIndex).TableName())

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
	*SpanProcessor

	projects map[uint32]*org.Project
	digest   *xxhash.Digest
}

func newSpanProcessorThread(p *SpanProcessor) *spanProcessorThread {
	return &spanProcessorThread{
		SpanProcessor: p,

		projects: make(map[uint32]*org.Project),
		digest:   xxhash.New(),
	}
}

func (p *spanProcessorThread) project(ctx context.Context, projectID uint32) (*org.Project, bool) {
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

func (p *spanProcessorThread) forceSpanName(ctx context.Context, span *Span) bool {
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
