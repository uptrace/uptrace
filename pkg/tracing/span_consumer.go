package tracing

import (
	"context"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"go.uber.org/zap"
)

type SpanConsumer struct {
	*BaseConsumer[SpanIndex, SpanData]
}

func NewSpanConsumer(p *ModuleParams) *SpanConsumer {
	batchSize := p.Conf.Spans.BatchSize
	bufferSize := p.Conf.Spans.BufferSize
	maxWorkers := p.Conf.Spans.MaxWorkers

	fakeApp := &bunapp.App{
		Conf:   p.Conf,
		Logger: p.Logger,
		CH:     p.CH,
	}
	var sgp *ServiceGraphProcessor
	if !p.Conf.ServiceGraph.Disabled {
		sgp = NewServiceGraphProcessor(fakeApp)
	}
	transformer := &spanTransformer{sgp: sgp, logger: p.Logger}

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
