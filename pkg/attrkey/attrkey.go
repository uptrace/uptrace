package attrkey

const (
	DisplayName = "display.name"

	SpanSystem  = ".system"
	SpanGroupID = ".group_id"

	SpanID       = ".id"
	SpanParentID = ".parent_id"
	SpanTraceID  = ".trace_id"

	SpanName      = ".name"
	SpanEventName = ".event_name"
	SpanIsEvent   = ".is_event"

	SpanKind     = ".kind"
	SpanTime     = ".time"
	SpanDuration = ".duration"

	SpanStatusCode    = ".status_code"
	SpanStatusMessage = ".status_message"

	SpanCount       = ".count"
	SpanCountPerMin = ".count_per_min"
	SpanErrorCount  = ".error_count"
	SpanErrorPct    = ".error_pct"
	SpanErrorRate   = ".error_rate"

	SpanLinkCount       = ".link_count"
	SpanEventCount      = ".event_count"
	SpanEventErrorCount = ".event_error_count"
	SpanEventLogCount   = ".event_log_count"

	ServiceName    = "service.name"
	ServiceVersion = "service.version"
	PeerService    = "peer.service"
	HostName       = "host.name"

	OtelLibraryName      = "otel.library.name"
	OtelLibraryVersion   = "otel.library.version"
	TelemetrySDKName     = "telemetry.sdk.name"
	TelemetrySDKVersion  = "telemetry.sdk.version"
	TelemetrySDKLanguage = "telemetry.sdk.language"

	HTTPUrl         = "http.url"    // GET
	HTTPMethod      = "http.method" // GET
	HTTPRoute       = "http.route"
	HTTPTarget      = "http.target"
	HTTPStatusCode  = "http.status_code"  // 200
	HTTPStatusClass = "http.status_class" // 2xx

	HTTPUserAgent          = "http.user_agent"
	HTTPUserAgentName      = "http.user_agent.name"
	HTTPUserAgentVersion   = "http.user_agent.version"
	HTTPUserAgentOS        = "http.user_agent.os"
	HTTPUserAgentOSVersion = "http.user_agent.os_version"
	HTTPUserAgentDevice    = "http.user_agent.device"
	HTTPUserAgentBot       = "http.user_agent.bot"

	DBSystem    = "db.system"
	DBName      = "db.name"
	DBStatement = "db.statement"
	DBOperation = "db.operation"
	DBSqlTable  = "db.sql.table"

	RPCSystem  = "rpc.system"
	RPCService = "rpc.service"
	RPCMethod  = "rpc.method"

	EnduserID    = "enduser.id"
	EnduserRole  = "enduser.role"
	EnduserScope = "enduser.scope"

	MessagingSystem                            = "messaging.system"
	MessagingOperation                         = "messaging.operation"
	MessagingDestinationName                   = "messaging.destination.name"
	MessagingDestinationKind                   = "messaging.destination.kind"
	MessagingDestinationTemporary              = "messaging.destination.temporary"
	MessagingMessageID                         = "messaging.message.id"
	MessagingMessagePayloadSizeBytes           = "messaging.message.payload_size_bytes"
	MessagingMessagePayloadCompressedSizeBytes = "messaging.message.payload_compressed_size_bytes"

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
	LogSource         = "log.source"
	LogFilePath       = "log.file.path"
	LogFileName       = "log.file.name"

	ExceptionType       = "exception.type"
	ExceptionMessage    = "exception.message"
	ExceptionStacktrace = "exception.stacktrace"

	CodeFunction          = "code.function"
	CodeFilepath          = "code.filepath"
	DeploymentEnvironment = "deployment.environment"

	MessageType             = "message.type"
	MessageID               = "message.id"
	MessageCompressedSize   = "message.compressed_size"
	MessageUncompressedSize = "message.uncompressed_size"

	CloudProvider  = "cloud.provider"
	CloudAccountID = "cloud.account.id"
	CloudRegion    = "cloud.region"
	CloudZone      = "cloud.zone"
)
