package tracing

import (
	"context"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"go.uber.org/zap"
)

type EventIndex struct {
	ch.CHModel `ch:"table:events_index,alias:s"`

	BaseIndex
	// Hide some fields from Span
	Duration      time.Duration `ch:"-"`
	StatusCode    string        `ch:"-"`
	StatusMessage string        `ch:"-"`

	LogSeverity   string `ch:",lc"`
	ExceptionType string `ch:",lc"`
}

type EventData struct {
	ch.CHModel `ch:"table:events_data,alias:s"`

	BaseData
}

type EventConsumer struct {
	*BaseConsumer[EventIndex, EventData]
}

func NewEventConsumer(app *bunapp.App) *EventConsumer {
	conf := app.Config()
	batchSize := conf.Events.BatchSize
	bufferSize := conf.Events.BufferSize
	maxWorkers := conf.Events.MaxWorkers
	transformer := &eventTransformer{logger: app.Logger}

	p := &EventConsumer{
		BaseConsumer: NewBaseConsumer[EventIndex, EventData](
			app,
			app.Logger,
			"uptrace.tracing.events_queue_length",
			batchSize, bufferSize, maxWorkers,
			transformer,
		),
	}

	p.logger.Info("starting processing events...",
		zap.Int("batch_size", batchSize),
		zap.Int("buffer_size", bufferSize),
		zap.Int("max_workers", maxWorkers),
	)
	p.Run()

	return p
}

type eventTransformer struct {
	logger *otelzap.Logger
}

func (c *eventTransformer) initIndexFromSpan(index *EventIndex, span *Span) {
	initEventIndex(index, span)
}

func (c *eventTransformer) initDataFromSpan(data *EventData, span *Span) {
	initEventData(data, span)
}

func (c *eventTransformer) postprocessIndex(ctx context.Context, index *EventIndex) {}

func initEventIndex(index *EventIndex, span *Span) {
	index.InitFromSpan(span)

	index.LogSeverity = span.Attrs.Text(attrkey.LogSeverity)
	index.ExceptionType = span.Attrs.Text(attrkey.ExceptionType)
}

func initEventData(data *EventData, span *Span) {
	data.InitFromSpan(span)
}
