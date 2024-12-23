package tracing

import (
	"maps"
	"strings"

	"github.com/uptrace/pkg/clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type SpanIndex struct {
	ch.CHModel `ch:"table:spans_index,alias:s"`

	BaseIndex

	ClientAddress       string `ch:",lc"`
	ClientSocketAddress string `ch:",lc"`
	ClientSocketPort    int32

	DBSystem    string   `ch:",lc"`
	DBName      string   `ch:",lc"`
	DBSqlTables []string `ch:"type:Array(LowCardinality(String))"`
	DBStatement string
	DBOperation string `ch:",lc"`

	ProcessPID                int32
	ProcessCommand            string `ch:",lc"`
	ProcessRuntimeName        string `ch:",lc"`
	ProcessRuntimeVersion     string `ch:",lc"`
	ProcessRuntimeDescription string `ch:",lc"`
}

func initSpanIndex(index *SpanIndex, span *Span) {
	index.InitFromSpan(TableSpansIndex, span)

	index.ClientAddress = span.Attrs.Text(attrkey.ClientAddress)
	index.ClientSocketAddress = span.Attrs.Text(attrkey.ClientSocketAddress)
	index.ClientSocketPort = int32(span.Attrs.Int64(attrkey.ClientSocketPort))

	index.DBSystem = span.Attrs.Text(attrkey.DBSystem)
	index.DBName = span.Attrs.Text(attrkey.DBName)
	index.DBStatement = span.Attrs.Text(attrkey.DBStatement)
	index.DBOperation = span.Attrs.Text(attrkey.DBOperation)

	// Populate index.DBSqlTables
	if val, ok := span.Attrs.Get(attrkey.DBSqlTables); ok {
		if ss, ok := val.([]string); ok {
			for _, s := range ss {
				index.DBSqlTables = append(index.DBSqlTables, utf8util.TruncLC(s))
			}
		}
	}

	index.ProcessPID = int32(span.Attrs.Int64(attrkey.ProcessPID))
	index.ProcessCommand = span.Attrs.Text(attrkey.ProcessCommand)
	index.ProcessRuntimeName = span.Attrs.Text(attrkey.ProcessRuntimeName)
	index.ProcessRuntimeVersion = span.Attrs.Text(attrkey.ProcessRuntimeVersion)
	index.ProcessRuntimeDescription = span.Attrs.Text(attrkey.ProcessRuntimeDescription)
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
	commonAttrs := listToSet([]string{
		attrkey.SpanID,
		attrkey.SpanTraceID,
		attrkey.SpanParentID,
		attrkey.SpanType,
		attrkey.SpanSystem,
		attrkey.SpanGroupID,
		attrkey.SpanKind,
		attrkey.SpanName,
		attrkey.SpanEventName,
		attrkey.SpanTime,
		attrkey.SpanCount,

		attrkey.DisplayName,

		attrkey.DeploymentEnvironment,
		attrkey.ServiceName,
		attrkey.ServiceVersion,
		attrkey.ServiceNamespace,
		attrkey.HostName,

		attrkey.TelemetrySDKName,
		attrkey.TelemetrySDKLanguage,
		attrkey.TelemetrySDKVersion,
		attrkey.TelemetryAutoVersion,

		attrkey.OtelLibraryName,
		attrkey.OtelLibraryVersion,
	})

	TableSpansIndex.IndexedColumns = maps.Clone(commonAttrs)
	maps.Copy(TableSpansIndex.IndexedColumns, listToSet([]string{
		attrkey.SpanDuration,
		attrkey.SpanStatusCode,
		attrkey.SpanStatusMessage,

		attrkey.ClientAddress,
		attrkey.ClientSocketAddress,
		attrkey.ClientSocketPort,

		attrkey.DBSystem,
		attrkey.DBName,
		attrkey.DBSqlTables,
		attrkey.DBStatement,
		attrkey.DBOperation,

		attrkey.ProcessPID,
		attrkey.ProcessCommand,
		attrkey.ProcessRuntimeName,
		attrkey.ProcessRuntimeVersion,
		attrkey.ProcessRuntimeDescription,
	}))

	TableLogsIndex.IndexedColumns = maps.Clone(commonAttrs)
	maps.Copy(TableLogsIndex.IndexedColumns, listToSet([]string{
		attrkey.LogSeverity,
		attrkey.LogFilePath,
		attrkey.LogFileName,
		attrkey.LogIOStream,
		attrkey.LogSource,

		attrkey.ExceptionType,
		attrkey.ExceptionStacktrace,
	}))

	TableEventsIndex.IndexedColumns = maps.Clone(commonAttrs)
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
