package tracing

import (
	"context"
	"slices"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/utf8util"
	"go.uber.org/zap"
	"go4.org/syncutil"
)

type LogIndex struct {
	ch.CHModel `ch:"table:logs_index,insert:logs_index_buffer,alias:s"`

	*Span
	Duration      time.Duration `ch:"-"`
	StatusCode    string        `ch:"-"`
	StatusMessage string        `ch:"-"`

	DisplayName string

	Count           float32
	LinkCount       uint8
	EventCount      uint8
	EventErrorCount uint8
	EventLogCount   uint8

	AllKeys      []string `ch:"type:Array(LowCardinality(String))"`
	StringKeys   []string `ch:"type:Array(LowCardinality(String))"`
	StringValues []string

	TelemetrySDKName     string `ch:",lc"`
	TelemetrySDKLanguage string `ch:",lc"`
	TelemetrySDKVersion  string `ch:",lc"`
	TelemetryAutoVersion string `ch:",lc"`

	OtelLibraryName    string `ch:",lc"`
	OtelLibraryVersion string `ch:",lc"`

	DeploymentEnvironment string `ch:",lc"`

	ServiceName      string `ch:",lc"`
	ServiceVersion   string `ch:",lc"`
	ServiceNamespace string `ch:",lc"`
	HostName         string `ch:",lc"`

	ClientAddress       string `ch:",lc"`
	ClientSocketAddress string `ch:",lc"`
	ClientSocketPort    int32

	URLScheme string `attr:"url.scheme" ch:",lc"`
	URLFull   string `attr:"url.full"`
	URLPath   string `attr:"url.path" ch:",lc"`

	HTTPRequestMethod      string `ch:",lc"`
	HTTPResponseStatusCode uint16
	HTTPRoute              string `ch:",lc"`

	RPCMethod  string `ch:",lc"`
	RPCService string `ch:",lc"`

	DBSystem    string `ch:",lc"`
	DBName      string `ch:",lc"`
	DBStatement string
	DBOperation string `ch:",lc"`
	DBSqlTable  string `ch:",lc"`

	LogSeverity   string `ch:",lc"`
	ExceptionType string `ch:",lc"`
}

type LogData struct {
	ch.CHModel `ch:"table:logs_data,insert:logs_data_buffer,alias:s"`

	Type      string `ch:",lc"`
	ProjectID uint32
	TraceID   idgen.TraceID
	ID        idgen.SpanID
	ParentID  idgen.SpanID
	Time      time.Time `ch:"type:DateTime64(6)"`
	Data      []byte
}

type LogConsumer struct {
	consumer *Consumer[LogIndex, LogData]
	logger   *otelzap.Logger
}

func NewLogConsumer(app *bunapp.App, gate *syncutil.Gate) *LogConsumer {
	conf := app.Config()
	batchSize := conf.Spans.BatchSize
	bufferSize := conf.Spans.BufferSize

	p := &LogConsumer{logger: app.Logger}

	c := &logConsumer{logger: app.Logger}
	p.consumer = NewConsumer[LogIndex, LogData](app, batchSize, bufferSize, gate, c)

	p.logger.Info("starting processing logs...",
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

func (p *LogConsumer) AddSpan(ctx context.Context, span *Span) {
	p.consumer.AddSpan(ctx, span)
}

type logConsumer struct {
	logger *otelzap.Logger
}

func (p *logConsumer) indexFromSpan(index *LogIndex, span *Span) {
	initLogIndex(index, span)
}

func (p *logConsumer) dataFromSpan(data *LogData, span *Span) {
	initLogData(data, span)
}

func (p *logConsumer) updateIndexStats(
	index *LogIndex,
	linkCount, eventCount, eventErrorCount, eventLogCount uint8,
) {
	index.LinkCount = linkCount
	index.EventCount = eventCount
	index.EventErrorCount = eventErrorCount
	index.EventLogCount = eventLogCount
}

func (p *logConsumer) postprocessIndex(ctx context.Context, index *LogIndex) {
}

func initLogIndex(index *LogIndex, span *Span) {
	index.Span = span

	index.DisplayName = utf8util.TruncLarge(span.DisplayName)
	index.Count = 1

	index.TelemetrySDKName = span.Attrs.Text(attrkey.TelemetrySDKName)
	index.TelemetrySDKLanguage = span.Attrs.Text(attrkey.TelemetrySDKLanguage)
	index.TelemetrySDKVersion = span.Attrs.Text(attrkey.TelemetrySDKVersion)
	index.TelemetryAutoVersion = span.Attrs.Text(attrkey.TelemetryAutoVersion)

	index.OtelLibraryName = span.Attrs.Text(attrkey.OtelLibraryName)
	index.OtelLibraryVersion = span.Attrs.Text(attrkey.OtelLibraryVersion)

	index.DeploymentEnvironment, _ = span.Attrs[attrkey.DeploymentEnvironment].(string)

	index.ServiceName = span.Attrs.ServiceName()
	index.ServiceVersion = span.Attrs.Text(attrkey.ServiceVersion)
	index.ServiceNamespace = span.Attrs.Text(attrkey.ServiceNamespace)
	index.HostName = span.Attrs.HostName()

	index.ClientAddress = span.Attrs.Text(attrkey.ClientAddress)
	index.ClientSocketAddress = span.Attrs.Text(attrkey.ClientSocketAddress)
	index.ClientSocketPort = int32(span.Attrs.Int64(attrkey.ClientSocketPort))

	index.URLScheme = span.Attrs.Text(attrkey.URLScheme)
	index.URLFull = span.Attrs.Text(attrkey.URLFull)
	index.URLPath = span.Attrs.Text(attrkey.URLPath)

	index.HTTPRequestMethod = span.Attrs.Text(attrkey.HTTPRequestMethod)
	index.HTTPResponseStatusCode = uint16(span.Attrs.Uint64(attrkey.HTTPResponseStatusCode))
	index.HTTPRoute = span.Attrs.Text(attrkey.HTTPRoute)

	index.RPCMethod = span.Attrs.Text(attrkey.RPCMethod)
	index.RPCService = span.Attrs.Text(attrkey.RPCService)

	index.DBSystem = span.Attrs.Text(attrkey.DBSystem)
	index.DBName = span.Attrs.Text(attrkey.DBName)
	index.DBStatement = span.Attrs.Text(attrkey.DBStatement)
	index.DBOperation = span.Attrs.Text(attrkey.DBOperation)
	index.DBSqlTable = span.Attrs.Text(attrkey.DBSqlTable)

	index.LogSeverity = span.Attrs.Text(attrkey.LogSeverity)
	index.ExceptionType = span.Attrs.Text(attrkey.ExceptionType)

	index.AllKeys = mapKeys(span.Attrs)
	slices.Sort(index.AllKeys)

	index.StringKeys, index.StringValues = attrKeysAndValues(span.Attrs, index.AllKeys)
}

func initLogData(data *LogData, span *Span) {
	data.Type = span.Type
	data.ProjectID = span.ProjectID
	data.TraceID = span.TraceID
	data.ID = span.ID
	data.ParentID = span.ParentID
	data.Time = span.Time
	data.Data = marshalSpanData(span)
}
