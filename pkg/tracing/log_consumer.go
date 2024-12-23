package tracing

import (
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
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

type LogConsumerParams struct {
	fx.In

	BaseConsumerParams
}

type LogConsumer struct {
	*BaseConsumer[LogIndex, LogData]
}

func NewLogConsumer(p LogConsumerParams) *LogConsumer {
	batchSize := p.Conf.Logs.BatchSize
	bufferSize := p.Conf.Logs.BufferSize
	maxWorkers := p.Conf.Logs.MaxWorkers
	transformer := &logTransformer{logger: p.Logger}

	c := &LogConsumer{
		BaseConsumer: NewBaseConsumer[LogIndex, LogData](
			p.Logger,
			p.PG,
			p.CH,
			p.Projects,
			p.MainQueue,
			"uptrace.tracing.logs_queue_length",
			batchSize, bufferSize, maxWorkers,
			transformer,
		),
	}

	p.Logger.Info("starting processing logs...",
		zap.Int("batch_size", batchSize),
		zap.Int("buffer_size", bufferSize),
		zap.Int("max_workers", maxWorkers),
	)

	return c
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

func initLogIndex(index *LogIndex, span *Span) {
	index.InitFromSpan(TableLogsIndex, span)

	index.LogSeverity = span.Attrs.Text(attrkey.LogSeverity)
	index.ExceptionType = span.Attrs.Text(attrkey.ExceptionType)
}

func initLogData(data *LogData, span *Span) {
	data.InitFromSpan(span)
}
