package tracing

import (
	"context"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type SpanConsumer struct {
	*BaseConsumer[SpanIndex, SpanData]
}

type SpanConsumerParams struct {
	fx.In
	BaseConsumerParams

	SGP *ServiceGraphProcessor `optional:"true"`
}

func NewSpanConsumer(p SpanConsumerParams) *SpanConsumer {
	batchSize := p.Conf.Spans.BatchSize
	bufferSize := p.Conf.Spans.BufferSize
	maxWorkers := p.Conf.Spans.MaxWorkers

	transformer := &spanTransformer{sgp: p.SGP, logger: p.Logger}

	c := &SpanConsumer{
		BaseConsumer: NewBaseConsumer[SpanIndex, SpanData](
			p.Logger,
			p.PG,
			p.CH,
			p.MainQueue,
			"uptrace.tracing.queue_length",
			batchSize, bufferSize, maxWorkers,
			transformer,
		),
	}

	p.Logger.Info("starting processing spans...",
		zap.Int("batch_size", batchSize),
		zap.Int("buffer_size", bufferSize),
		zap.Int("max_workers", maxWorkers),
	)

	return c
}

type spanTransformer struct {
	sgp    *ServiceGraphProcessor
	logger *otelzap.Logger
}

func (c *spanTransformer) initIndexFromSpan(index *SpanIndex, span *Span) {
	initSpanIndex(index, span)
}

func (c *spanTransformer) initDataFromSpan(data *SpanData, span *Span) {
	initSpanData(data, span)
}

func (c *spanTransformer) postprocessIndex(ctx context.Context, index *SpanIndex) {
	if c.sgp != nil {
		if err := c.sgp.ProcessSpan(ctx, index); err != nil {
			c.logger.Error("service graph failed", zap.Error(err))
		}
	}
}
