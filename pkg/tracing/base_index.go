package tracing

import (
	"slices"

	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type BaseIndex struct {
	*Span

	DisplayName string

	Count float32

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
}

func (index *BaseIndex) InitFromSpan(table *Table, span *Span) {
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

	index.AllKeys = mapKeys(span.Attrs)
	slices.Sort(index.AllKeys)

	index.StringKeys, index.StringValues = attrKeysAndValues(table, span.Attrs, index.AllKeys)
}
