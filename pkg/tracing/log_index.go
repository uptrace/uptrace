package tracing

import (
	"slices"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type LogIndex struct {
	ch.CHModel `ch:"table:logs_index,insert:logs_index_buffer,alias:s"`

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

	LogSeverity   string `ch:",lc"`
	ExceptionType string `ch:",lc"`
}

func initLogIndex(index *LogIndex, log *Span) {
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

	index.LogSeverity = log.Attrs.Text(attrkey.LogSeverity)
	index.ExceptionType = log.Attrs.Text(attrkey.ExceptionType)

	index.AllKeys = mapKeys(log.Attrs)
	slices.Sort(index.AllKeys)

	index.StringKeys, index.StringValues = attrKeysAndValues(log.Attrs, index.AllKeys)
}
