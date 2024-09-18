package tracing

import (
	"slices"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type LogIndex struct {
	ch.CHModel `ch:"table:logs_index,insert:logs_index_buffer,alias:s"`

	*Span

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

func initLogIndex(index *LogIndex, log *Span) {
	index.Span = log

	index.DisplayName = utf8util.TruncLarge(log.DisplayName)
	index.Count = 1

	index.TelemetrySDKName = log.Attrs.Text(attrkey.TelemetrySDKName)
	index.TelemetrySDKLanguage = log.Attrs.Text(attrkey.TelemetrySDKLanguage)
	index.TelemetrySDKVersion = log.Attrs.Text(attrkey.TelemetrySDKVersion)
	index.TelemetryAutoVersion = log.Attrs.Text(attrkey.TelemetryAutoVersion)

	index.OtelLibraryName = log.Attrs.Text(attrkey.OtelLibraryName)
	index.OtelLibraryVersion = log.Attrs.Text(attrkey.OtelLibraryVersion)

	index.DeploymentEnvironment, _ = log.Attrs[attrkey.DeploymentEnvironment].(string)

	index.ServiceName = log.Attrs.ServiceName()
	index.ServiceVersion = log.Attrs.Text(attrkey.ServiceVersion)
	index.ServiceNamespace = log.Attrs.Text(attrkey.ServiceNamespace)
	index.HostName = log.Attrs.HostName()

	index.ClientAddress = log.Attrs.Text(attrkey.ClientAddress)
	index.ClientSocketAddress = log.Attrs.Text(attrkey.ClientSocketAddress)
	index.ClientSocketPort = int32(log.Attrs.Int64(attrkey.ClientSocketPort))

	index.URLScheme = log.Attrs.Text(attrkey.URLScheme)
	index.URLFull = log.Attrs.Text(attrkey.URLFull)
	index.URLPath = log.Attrs.Text(attrkey.URLPath)

	index.HTTPRequestMethod = log.Attrs.Text(attrkey.HTTPRequestMethod)
	index.HTTPResponseStatusCode = uint16(log.Attrs.Uint64(attrkey.HTTPResponseStatusCode))
	index.HTTPRoute = log.Attrs.Text(attrkey.HTTPRoute)

	index.RPCMethod = log.Attrs.Text(attrkey.RPCMethod)
	index.RPCService = log.Attrs.Text(attrkey.RPCService)

	index.DBSystem = log.Attrs.Text(attrkey.DBSystem)
	index.DBName = log.Attrs.Text(attrkey.DBName)
	index.DBStatement = log.Attrs.Text(attrkey.DBStatement)
	index.DBOperation = log.Attrs.Text(attrkey.DBOperation)
	index.DBSqlTable = log.Attrs.Text(attrkey.DBSqlTable)

	index.LogSeverity = log.Attrs.Text(attrkey.LogSeverity)
	index.ExceptionType = log.Attrs.Text(attrkey.ExceptionType)

	index.AllKeys = mapKeys(log.Attrs)
	slices.Sort(index.AllKeys)

	index.StringKeys, index.StringValues = attrKeysAndValues(log.Attrs, index.AllKeys)
}
