package tracing

import (
	"maps"
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
	index.InitFromSpan(TableSpansIndex, span)
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
		keys = append(keys, key)
		values = append(values, utf8util.TruncSmall(asString(m[key])))
	}
	return keys, values
}

var (
	commonAttrs = []string{
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
	}
)

func IsIndexedAttr(table *Table, attrKey string) bool {
	if strings.HasPrefix(attrKey, "_") {
		return true
	}
	return table.IsIndexedAttr(attrKey)
}

type Table struct {
	Name           string              // spans_index
	IndexedColumns map[string]struct{} //  _kind, _duration, log_severity
}

func (t *Table) IsIndexedAttr(attrKey string) bool {
	_, ok := t.IndexedColumns[attrKey]
	return ok
}

var (
	TableSpansIndex  = &Table{Name: "spans_index"}
	TableLogsIndex   = &Table{Name: "logs_index"}
	TableEventsIndex = &Table{Name: "events_index"}
)

func init() {
	commonAttrsSet := listToSet(commonAttrs)

	TableSpansIndex.IndexedColumns = maps.Clone(commonAttrsSet)
	maps.Copy(TableSpansIndex.IndexedColumns, listToSet([]string{
		attrkey.ClientAddress,
		attrkey.ClientSocketAddress,
		attrkey.ClientSocketPort,

		attrkey.DBSystem,
		attrkey.DBName,
		attrkey.DBSqlTable,
		attrkey.DBStatement,
		attrkey.DBOperation,

		attrkey.ProcessPID,
		attrkey.ProcessCommand,
		attrkey.ProcessRuntimeName,
		attrkey.ProcessRuntimeVersion,
		attrkey.ProcessRuntimeDescription,
	}))

	TableLogsIndex.IndexedColumns = maps.Clone(commonAttrsSet)
	maps.Copy(TableLogsIndex.IndexedColumns, listToSet([]string{
		attrkey.LogSeverity,
		attrkey.LogFilePath,
		attrkey.LogFileName,
		attrkey.LogIOStream,
		attrkey.LogSource,

		attrkey.ExceptionType,
		attrkey.ExceptionStacktrace,
	}))

	TableEventsIndex.IndexedColumns = maps.Clone(commonAttrsSet)
	maps.Copy(TableEventsIndex.IndexedColumns, listToSet([]string{
		attrkey.ProcessPID,
		attrkey.ProcessCommand,
		attrkey.ProcessRuntimeName,
		attrkey.ProcessRuntimeVersion,
		attrkey.ProcessRuntimeDescription,

		attrkey.MessagingMessageID,
		attrkey.MessagingMessageType,
		attrkey.MessagingMessagePayloadSizeBytes,
	}))
}
