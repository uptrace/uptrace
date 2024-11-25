package tracing

import (
	"context"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"go.opentelemetry.io/otel/metric"
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

	sp := &spanTransformer{sgp: sgp, logger: app.Logger}
	p.consumer = NewConsumer[SpanIndex, SpanData](app, batchSize, bufferSize, gate, sp)

	p.logger.Info("starting processing spans...",
		zap.Int("batch_size", batchSize),
		zap.Int("buffer_size", bufferSize))

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()

		p.consumer.processLoop(app.Context())
	}()

	queueLen, _ := bunotel.Meter.Int64ObservableGauge("uptrace.tracing.queue_length",
		metric.WithUnit("{spans}"),
	)

	if _, err := bunotel.Meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			o.ObserveInt64(queueLen, int64(len(p.consumer.queue)))
			return nil
		},
		queueLen,
	); err != nil {
		panic(err)
	}

	return p
}

func (p *SpanConsumer) AddSpan(ctx context.Context, span *Span) {
	p.consumer.AddSpan(ctx, span)
}

type spanTransformer struct {
	sgp    *ServiceGraphProcessor
	logger *otelzap.Logger
}

func (c *spanTransformer) indexFromSpan(span *Span) SpanIndex {
	index := SpanIndex{}
	initSpanIndex(&index, span)
	return index
}

func (c *spanTransformer) dataFromSpan(span *Span) SpanData {
	data := SpanData{}
	initSpanData(&data, span)
	return data
}

func (c *spanTransformer) postprocessIndex(ctx context.Context, index *SpanIndex) {
	if c.sgp != nil {
		if err := c.sgp.ProcessSpan(ctx, index); err != nil {
			c.logger.Error("service graph failed", zap.Error(err))
		}
	}
}
