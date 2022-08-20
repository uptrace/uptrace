package tracing

import (
	"strings"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/tracing/attrkey"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type SpanIndex struct {
	ch.CHModel `ch:"table:spans_index_buffer,alias:s"`

	*Span

	Count float32 `ch:"span.count"` // sampling adjusted count

	LinkCount       uint8 `ch:"span.link_count"`
	EventCount      uint8 `ch:"span.event_count"`
	EventErrorCount uint8 `ch:"span.event_error_count"`
	EventLogCount   uint8 `ch:"span.event_log_count"`

	AllKeys    []string `ch:",lc"`
	AttrKeys   []string `ch:",lc"`
	AttrValues []string `ch:",lc"`

	ServiceName string `ch:"service.name,lc"`
	HostName    string `ch:"host.name,lc"`

	DBSystem    string `ch:"db.system,lc"`
	DBStatement string `ch:"db.statement"`
	DBOperation string `ch:"db.operation,lc"`
	DBSqlTable  string `ch:"db.sql.table,lc"`

	LogSeverity string `ch:"log.severity,lc"`
	LogMessage  string `ch:"log.message"`

	ExceptionType    string `ch:"exception.type,lc"`
	ExceptionMessage string `ch:"exception.message"`
}

func initSpanIndex(index *SpanIndex, span *Span) {
	index.Span = span
	index.Count = 1

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
