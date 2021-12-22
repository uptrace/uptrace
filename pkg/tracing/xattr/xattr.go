package xattr

const (
	ItemID   = "item.id"
	ItemName = "item.name"

	SpanSystem  = "span.system"
	SpanGroupID = "span.group_id"

	SpanID       = "span.id"
	SpanParentID = "span.parent_id"
	SpanTraceID  = "span.trace_id"

	SpanName         = "span.name"
	SpanKind         = "span.kind"
	SpanTime         = "span.time"
	SpanDuration     = "span.duration"
	SpanDurationSelf = "span.duration_self"

	SpanStatusCode    = "span.status_code"
	SpanStatusMessage = "span.status_message"

	SpanCount       = "span.count"
	SpanCountPerMin = "span.count_per_min"
	SpanErrorCount  = "span.error_count"
	SpanErrorPct    = "span.error_pct"

	ServiceName = "service.name"
	HostName    = "host.name"
	RPCSystem   = "rpc.system"

	MessagingSystem    = "messaging.system"
	MessagingOperation = "messaging.operation"

	DBSystem    = "db.system"
	DBStatement = "db.statement"
	DBSqlTable  = "db.sql.table"

	HTTPRoute  = "http.route"
	HTTPTarget = "http.target"

	LogMessage  = "log.message"
	LogSeverity = "log.severity"

	ExceptionType       = "exception.type"
	ExceptionMessage    = "exception.message"
	ExceptionStacktrace = "exception.stacktrace"

	OtelLibraryName    = "otel.library.name"
	OtelLibraryVersion = "otel.library.version"
)
