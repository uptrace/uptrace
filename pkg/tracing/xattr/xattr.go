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
	SpanEventName    = "span.event_name"
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

	SpanLinkCount       = "span.link_count"
	SpanEventCount      = "span.event_count"
	SpanEventErrorCount = "span.event_error_count"
	SpanEventLogCount   = "span.event_log_count"

	ServiceName = "service.name"
	HostName    = "host.name"

	RPCSystem  = "rpc.system"
	RPCService = "rpc.service"
	RPCMethod  = "rpc.method"

	MessagingSystem          = "messaging.system"
	MessagingOperation       = "messaging.operation"
	MessagingDestination     = "messaging.destination"
	MessagingDestinationKind = "messaging.destination_kind"

	DBSystem    = "db.system"
	DBStatement = "db.statement"
	DBOperation = "db.operation"
	DBSqlTable  = "db.sql.table"

	HTTPMethod = "http.method" // GET
	HTTPRoute  = "http.route"
	HTTPTarget = "http.target"

	HTTPUserAgent          = "http.user_agent"
	HTTPUserAgentName      = "http.user_agent.name"
	HTTPUserAgentVersion   = "http.user_agent.version"
	HTTPUserAgentOS        = "http.user_agent.os"
	HTTPUserAgentOSVersion = "http.user_agent.os_version"
	HTTPUserAgentDevice    = "http.user_agent.device"
	HTTPUserAgentBot       = "http.user_agent.bot"

	LogMessage  = "log.message"
	LogSeverity = "log.severity"
	LogSource   = "log.source"
	LogFilepath = "log.filepath"

	ExceptionType       = "exception.type"
	ExceptionMessage    = "exception.message"
	ExceptionStacktrace = "exception.stacktrace"

	OtelLibraryName    = "otel.library.name"
	OtelLibraryVersion = "otel.library.version"

	TelemetrySDKName     = "telemetry.sdk.name"
	TelemetrySDKVersion  = "telemetry.sdk.version"
	TelemetrySDKLanguage = "telemetry.sdk.language"

	DeploymentEnvironment = "deployment.environment"
)
