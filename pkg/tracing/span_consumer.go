package tracing

import (
	"context"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"go.uber.org/zap"
)

type SpanConsumer struct {
	*BaseConsumer[SpanIndex, SpanData]
	logger *otelzap.Logger
}

func NewSpanConsumer(app *bunapp.App) *SpanConsumer {
	conf := app.Config()
	batchSize := conf.Spans.BatchSize
	bufferSize := conf.Spans.BufferSize
	maxWorkers := conf.Spans.MaxWorkers

	var sgp *ServiceGraphProcessor
	if !conf.ServiceGraph.Disabled {
		sgp = NewServiceGraphProcessor(app)
	}
	transformer := &spanTransformer{sgp: sgp, logger: app.Logger}

	p := &SpanConsumer{
		logger: app.Logger,
		BaseConsumer: NewBaseConsumer[SpanIndex, SpanData](
			app,
			"uptrace.tracing.queue_length",
			batchSize, bufferSize, maxWorkers,
			transformer,
		),
	}

	p.logger.Info("starting processing spans...",
		zap.Int("batch_size", batchSize),
		zap.Int("buffer_size", bufferSize),
		zap.Int("max_workers", maxWorkers),
	)
	p.Run()

	return p
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
