package tracing

import (
	"strings"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/tracing/xattr"
	"github.com/uptrace/uptrace/pkg/tracing/xotel"
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

	index.DBSystem, _ = span.Attrs[xattr.DBSystem].(string)
	index.DBStatement, _ = span.Attrs[xattr.DBStatement].(string)
	index.DBOperation, _ = span.Attrs[xattr.DBOperation].(string)
	index.DBSqlTable, _ = span.Attrs[xattr.DBSqlTable].(string)

	index.LogSeverity, _ = span.Attrs[xattr.LogSeverity].(string)
	index.LogMessage, _ = span.Attrs[xattr.LogMessage].(string)

	index.ExceptionType, _ = span.Attrs[xattr.ExceptionType].(string)
	index.ExceptionMessage, _ = span.Attrs[xattr.ExceptionMessage].(string)

	index.AllKeys = mapKeys(span.Attrs)
	index.AttrKeys, index.AttrValues = attrKeysAndValues(span.Attrs)
}

func mapKeys(m xotel.AttrMap) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

var (
	indexedAttrs = []string{
		xattr.ServiceName,
		xattr.HostName,

		xattr.DBSystem,
		xattr.DBStatement,
		xattr.DBOperation,
		xattr.DBSqlTable,

		xattr.LogSeverity,
		xattr.LogMessage,

		xattr.ExceptionType,
		xattr.ExceptionMessage,
	}
	indexedAttrSet = listToSet(indexedAttrs)
)

func attrKeysAndValues(m xotel.AttrMap) ([]string, []string) {
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
