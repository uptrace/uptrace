package tracing

import (
	"strings"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type SpanIndex struct {
	ch.CHModel `ch:"table:spans_index,alias:s"`

	BaseIndex
}

func initSpanIndex(index *SpanIndex, span *Span) {
	index.InitFromSpan(TableSpans, span)
}

func mapKeys(m AttrMap) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func attrKeysAndValues(table *Table, m AttrMap, sortedKeys []string) ([]string, []string) {
	keys := make([]string, 0, len(m))
	values := make([]string, 0, len(m))
	for _, key := range sortedKeys {
		if table.IsIndexedAttr(key) {
			continue
		}
		//if strings.HasPrefix(key, "_") {
		//	continue
		//}
		//if IsIndexedAttr(key) {
		//	continue
		//}
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

func IsIndexedAttr(table *Table, attrKey string) bool {
	if strings.HasPrefix(attrKey, "_") {
		return true
	}
	return table.IsIndexedAttr(attrKey)
}

type Table struct {
	Name           string          // spans_index
	IndexedColumns map[string]bool //  _kind, _duration, log_severity
}

func (t *Table) IsIndexedAttr(attrKey string) bool {
	return t.IndexedColumns[attrKey]
}

var (
	TableSpans = &Table{
		Name:           TableSpansIndexName,
		IndexedColumns: map[string]bool{},
	}
	TableLogs = &Table{
		Name:           TableLogsIndexName,
		IndexedColumns: map[string]bool{},
	}
	TableEvents = &Table{
		Name:           TableEventsIndexName,
		IndexedColumns: map[string]bool{},
	}
)
