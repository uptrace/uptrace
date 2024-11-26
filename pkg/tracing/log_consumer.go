package tracing

import (
	"context"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

type LogIndex struct {
	ch.CHModel `ch:"table:logs_index,insert:logs_index_buffer,alias:s"`

	BaseIndex
	// Hide some fields from Span
	Duration      time.Duration `ch:"-"`
	StatusCode    string        `ch:"-"`
	StatusMessage string        `ch:"-"`

	LogSeverity   string `ch:",lc"`
	ExceptionType string `ch:",lc"`
}

type LogData struct {
	ch.CHModel `ch:"table:logs_data,insert:logs_data_buffer,alias:s"`

	BaseData
}

type LogConsumer struct {
	consumer *Consumer[LogIndex, LogData]
	logger   *otelzap.Logger
}

func NewLogConsumer(app *bunapp.App) *LogConsumer {
	conf := app.Config()
	batchSize := conf.Logs.BatchSize
	bufferSize := conf.Logs.BufferSize
	maxWorkers := conf.Logs.MaxWorkers

	transformer := &logTransformer{logger: app.Logger}

	p := &LogConsumer{logger: app.Logger}
	p.consumer = NewConsumer[LogIndex, LogData](app, batchSize, bufferSize, maxWorkers, transformer)

	p.logger.Info("starting processing logs...",
		zap.Int("batch_size", batchSize),
		zap.Int("buffer_size", bufferSize))

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()

		p.consumer.processLoop(app.Context())
	}()

	queueLen, _ := bunotel.Meter.Int64ObservableGauge("uptrace.tracing.log_queue_length",
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

func (p *LogConsumer) AddSpan(ctx context.Context, span *Span) {
	p.consumer.AddSpan(ctx, span)
}

type logTransformer struct {
	logger *otelzap.Logger
}

func (c *logTransformer) initIndexFromSpan(index *LogIndex, span *Span) {
	initLogIndex(index, span)
}

func (c *logTransformer) initDataFromSpan(data *LogData, span *Span) {
	initLogData(data, span)
}

func (c *logTransformer) postprocessIndex(ctx context.Context, index *LogIndex) {}

func initLogIndex(index *LogIndex, span *Span) {
	index.InitFromSpan(span)

	index.LogSeverity = span.Attrs.Text(attrkey.LogSeverity)
	index.ExceptionType = span.Attrs.Text(attrkey.ExceptionType)
}

func initLogData(data *LogData, span *Span) {
	data.InitFromSpan(span)
}
