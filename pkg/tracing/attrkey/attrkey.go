package attrkey

const (
	ItemID    = "item.id"
	ItemName  = "item.name"
	ItemQuery = "item.query"

	SpanSystem  = "span.system"
	SpanGroupID = "span.group_id"

	SpanID       = "span.id"
	SpanParentID = "span.parent_id"
	SpanTraceID  = "span.trace_id"

	SpanName      = "span.name"
	SpanEventName = "span.event_name"
	SpanIsEvent   = "span.is_event"

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

	Service        = "service"
	ServiceName    = "service.name"
	ServiceVersion = "service.version"
	PeerService    = "peer.service"
	HostName       = "host.name"

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

	NetTransport         = "net.transport"
	NetPeerIP            = "net.peer.ip"
	NetPeerIPCountryCode = "net.peer.ip.country_code"
	NetPeerIPCountryName = "net.peer.ip.country_name"
	NetPeerIPCityName    = "net.peer.ip.city_name"
	NetPeerPort          = "net.peer.port"
	NetPeerName          = "net.peer.name"
	NetHostIP            = "net.host.ip"
	NetHostIPCountryCode = "net.host.ip.country_code"
	NetHostIPCountryName = "net.host.ip.country_name"
	NetHostIPCityName    = "net.host.ip.city_name"
	NetHostPort          = "net.host.port"
	NetHostName          = "net.host.name"

	LogMessage        = "log.message"
	LogSeverity       = "log.severity"
	LogSeverityNumber = "log.severity_number"
	LogFilepath       = "log.filepath"

	ExceptionType       = "exception.type"
	ExceptionMessage    = "exception.message"
	ExceptionStacktrace = "exception.stacktrace"

	OtelLibraryName    = "otel.library.name"
	OtelLibraryVersion = "otel.library.version"

	TelemetrySDKName     = "telemetry.sdk.name"
	TelemetrySDKVersion  = "telemetry.sdk.version"
	TelemetrySDKLanguage = "telemetry.sdk.language"

	CodeFunction          = "code.function"
	DeploymentEnvironment = "deployment.environment"
)
