package tracing

import (
	"strings"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type SpanIndex struct {
	ch.CHModel `ch:"table:spans_index_buffer,alias:s"`

	*Span

	Count           float32 `ch:"count"`
	LinkCount       uint8   `ch:"link_count"`
	EventCount      uint8   `ch:"event_count"`
	EventErrorCount uint8   `ch:"event_error_count"`
	EventLogCount   uint8   `ch:"event_log_count"`

	AllKeys    []string `ch:"all_keys,lc"`
	AttrKeys   []string `ch:"attr_keys,lc"`
	AttrValues []string `ch:"attr_values,lc"`

	DeploymentEnvironment string `ch:"deployment_environment,lc"`

	ServiceName string `ch:"service_name,lc"`
	HostName    string `ch:"host_name,lc"`

	DBSystem    string `ch:"db_system,lc"`
	DBStatement string `ch:"db_statement"`
	DBOperation string `ch:"db_operation,lc"`
	DBSqlTable  string `ch:"db_sql_table,lc"`

	LogSeverity string `ch:"log_severity,lc"`
	LogMessage  string `ch:"log_message"`

	ExceptionType    string `ch:"exception_type,lc"`
	ExceptionMessage string `ch:"exception_message"`
}

func initSpanIndex(index *SpanIndex, span *Span) {
	index.Span = span
	index.Count = 1

	index.DeploymentEnvironment, _ = span.Attrs[attrkey.DeploymentEnvironment].(string)

	index.ServiceName = span.Attrs.ServiceName()
	index.HostName = span.Attrs.HostName()

	index.DBSystem, _ = span.Attrs[attrkey.DBSystem].(string)
	index.DBStatement, _ = span.Attrs[attrkey.DBStatement].(string)
	index.DBOperation, _ = span.Attrs[attrkey.DBOperation].(string)
	index.DBSqlTable, _ = span.Attrs[attrkey.DBSqlTable].(string)

	index.LogSeverity, _ = span.Attrs[attrkey.LogSeverity].(string)
	index.LogMessage, _ = span.Attrs[attrkey.LogMessage].(string)

	index.ExceptionType, _ = span.Attrs[attrkey.ExceptionType].(string)
	index.ExceptionMessage, _ = span.Attrs[attrkey.ExceptionMessage].(string)

	index.AllKeys = mapKeys(span.Attrs)
	index.AttrKeys, index.AttrValues = attrKeysAndValues(span.Attrs)
}

func mapKeys(m AttrMap) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

var (
	indexedAttrs = []string{
		attrkey.DisplayName,

		attrkey.DeploymentEnvironment,
		attrkey.ServiceName,
		attrkey.HostName,

		attrkey.DBSystem,
		attrkey.DBStatement,
		attrkey.DBOperation,
		attrkey.DBSqlTable,

		attrkey.LogSeverity,
		attrkey.LogMessage,

		attrkey.ExceptionType,
		attrkey.ExceptionMessage,
	}
	indexedAttrSet = listToSet(indexedAttrs)
)

func attrKeysAndValues(m AttrMap) ([]string, []string) {
	keys := make([]string, 0, len(m))
	values := make([]string, 0, len(m))
	for k, v := range m {
		if strings.HasPrefix(k, "_") {
			continue
		}
		if _, ok := indexedAttrSet[k]; ok {
			continue
		}
		keys = append(keys, k)
		values = append(values, utf8util.TruncMedium(asString(v)))
	}
	return keys, values
}
