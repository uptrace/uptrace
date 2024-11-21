package tracing

import (
	"context"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"go.uber.org/zap"
	"go4.org/syncutil"
)

type SpanConsumer struct {
	consumer *Consumer[SpanIndex, SpanData]
	logger   *otelzap.Logger
}

func NewSpanConsumer(app *bunapp.App, gate *syncutil.Gate) *SpanConsumer {
	conf := app.Config()
	batchSize := conf.Spans.BatchSize
	bufferSize := conf.Spans.BufferSize

	p := &SpanConsumer{logger: app.Logger}

	var sgp *ServiceGraphProcessor
	if !conf.ServiceGraph.Disabled {
		sgp = NewServiceGraphProcessor(app)
	}

	c := &spanConsumer{sgp: sgp, logger: app.Logger}
	p.consumer = NewConsumer[SpanIndex, SpanData](app, batchSize, bufferSize, gate, c)

	p.logger.Info("starting processing spans...",
		zap.Int("batch_size", batchSize),
		zap.Int("buffer_size", bufferSize))

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()

		p.consumer.processLoop(app.Context())
	}()

	/*
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
	*/

	return p
}

func (p *SpanConsumer) AddSpan(ctx context.Context, span *Span) {
	p.consumer.AddSpan(ctx, span)
}

type spanConsumer struct {
	sgp    *ServiceGraphProcessor
	logger *otelzap.Logger
}

func (p *spanConsumer) indexFromSpan(index *SpanIndex, span *Span) {
	initSpanIndex(index, span)
}

func (p *spanConsumer) dataFromSpan(data *SpanData, span *Span) {
	initSpanData(data, span)
}

func (p *spanConsumer) updateIndexStats(
	index *SpanIndex,
	linkCount, eventCount, eventErrorCount, eventLogCount uint8,
) {
	index.LinkCount = linkCount
	index.EventCount = eventCount
	index.EventErrorCount = eventErrorCount
	index.EventLogCount = eventLogCount
}

func (p *spanConsumer) postprocessIndex(ctx context.Context, index *SpanIndex) {
	if p.sgp != nil {
		if err := p.sgp.ProcessSpan(ctx, index); err != nil {
			p.logger.Error("service graph failed", zap.Error(err))
		}
	}
}
