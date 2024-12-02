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

type LogIndex struct {
	ch.CHModel `ch:"table:logs_index,alias:s"`

	BaseIndex
	// Hide some fields from Span
	Duration      time.Duration `ch:"-"`
	StatusCode    string        `ch:"-"`
	StatusMessage string        `ch:"-"`

	LogSeverity   string `ch:",lc"`
	ExceptionType string `ch:",lc"`
}

type LogData struct {
	ch.CHModel `ch:"table:logs_data,alias:s"`

	BaseData
}

type LogConsumer struct {
	*BaseConsumer[LogIndex, LogData]
	logger *otelzap.Logger
}

func NewLogConsumer(app *bunapp.App) *LogConsumer {
	conf := app.Config()
	batchSize := conf.Logs.BatchSize
	bufferSize := conf.Logs.BufferSize
	maxWorkers := conf.Logs.MaxWorkers
	transformer := &logTransformer{logger: app.Logger}

	p := &LogConsumer{
		logger: app.Logger,
		BaseConsumer: NewBaseConsumer[LogIndex, LogData](
			app,
			app.Logger,
			"uptrace.tracing.logs_queue_length",
			batchSize, bufferSize, maxWorkers,
			transformer,
		),
	}

	p.logger.Info("starting processing logs...",
		zap.Int("batch_size", batchSize),
		zap.Int("buffer_size", bufferSize),
		zap.Int("max_workers", maxWorkers),
	)
	p.Run()

	return p
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
