package tracing

import (
	"github.com/uptrace/go-clickhouse/ch"
)

type SpanIndex struct {
	ch.CHModel `ch:"table:spans_index_buffer,alias:s"`

	*Span

	Count float32 `ch:"span.count"` // sampling adjusted count

	EventCount      uint8 `ch:"span.event_count"`
	EventErrorCount uint8 `ch:"span.event_error_count"`
	EventLogCount   uint8 `ch:"span.event_log_count"`

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
