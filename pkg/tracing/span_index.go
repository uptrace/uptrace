package tracing

import (
	"slices"
	"strings"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type SpanIndex struct {
	ch.CHModel `ch:"table:spans_index,insert:spans_index_buffer,alias:s"`

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

	// LogSeverity   string `ch:",lc"`
	// ExceptionType string `ch:",lc"`
}

func initSpanIndex(index *SpanIndex, span *Span) {
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

	// index.LogSeverity = span.Attrs.Text(attrkey.LogSeverity)
	// index.ExceptionType = span.Attrs.Text(attrkey.ExceptionType)

	index.AllKeys = mapKeys(span.Attrs)
	slices.Sort(index.AllKeys)

	index.StringKeys, index.StringValues = attrKeysAndValues(span.Attrs, index.AllKeys)
}

func mapKeys(m AttrMap) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func attrKeysAndValues(m AttrMap, sortedKeys []string) ([]string, []string) {
	keys := make([]string, 0, len(m))
	values := make([]string, 0, len(m))
	for _, key := range sortedKeys {
		if strings.HasPrefix(key, "_") {
			continue
		}
		if IsIndexedAttr(key) {
			continue
		}
		keys = append(keys, key)
		values = append(values, utf8util.TruncSmall(asString(m[key])))
	}
	return keys, values
}

var (
	indexedAttrs = []string{
		attrkey.DisplayName,

		attrkey.TelemetrySDKName,
		attrkey.TelemetrySDKLanguage,
		attrkey.TelemetrySDKVersion,
		attrkey.TelemetryAutoVersion,

		attrkey.OtelLibraryName,
		attrkey.OtelLibraryVersion,

		attrkey.DeploymentEnvironment,

		attrkey.ServiceName,
		attrkey.ServiceVersion,
		attrkey.ServiceNamespace,
		attrkey.HostName,

		attrkey.ClientAddress,
		attrkey.ClientSocketAddress,
		attrkey.ClientSocketPort,

		attrkey.URLScheme,
		attrkey.URLFull,
		attrkey.URLPath,

		attrkey.HTTPRequestMethod,
		attrkey.HTTPResponseStatusCode,
		attrkey.HTTPRoute,

		attrkey.RPCMethod,
		attrkey.RPCService,

		attrkey.DBSystem,
		attrkey.DBName,
		attrkey.DBStatement,
		attrkey.DBOperation,
		attrkey.DBSqlTable,

		attrkey.LogSeverity,
		attrkey.ExceptionType,
	}
	indexedAttrSet = listToSet(indexedAttrs)
)

func IsIndexedAttr(attrKey string) bool {
	if strings.HasPrefix(attrKey, "_") {
		return true
	}
	_, ok := indexedAttrSet[attrKey]
	return ok
}
