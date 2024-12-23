package tracing

import (
	"slices"

	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type BaseIndex struct {
	*Span

	DisplayName string
	Count       float32

	AllKeys      []string `ch:"type:Array(LowCardinality(String))"`
	StringKeys   []string `ch:"type:Array(LowCardinality(String))"`
	StringValues []string

	DeploymentEnvironment string `ch:",lc"`
	ServiceName           string `ch:",lc"`
	ServiceVersion        string `ch:",lc"`
	ServiceNamespace      string `ch:",lc"`
	HostName              string `ch:",lc"`

	OtelLibraryName    string `ch:",lc"`
	OtelLibraryVersion string `ch:",lc"`

	TelemetrySDKName     string `ch:",lc"`
	TelemetrySDKLanguage string `ch:",lc"`
	TelemetrySDKVersion  string `ch:",lc"`
	TelemetryAutoVersion string `ch:",lc"`
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

	index.AllKeys = mapKeys(span.Attrs)
	slices.Sort(index.AllKeys)

	index.StringKeys, index.StringValues = attrKeysAndValues(table, span.Attrs, index.AllKeys)
}
