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
	SpanErrorRate   = ".error_rate"

	SpanLinkCount       = ".link_count"
	SpanEventCount      = ".event_count"
	SpanEventErrorCount = ".event_error_count"
	SpanEventLogCount   = ".event_log_count"

	TelemetrySDKName     = "telemetry.sdk.name"
	TelemetrySDKVersion  = "telemetry.sdk.version"
	TelemetrySDKLanguage = "telemetry.sdk.language"
	TelemetryAutoVersion = "telemetry.auto.version"

	OtelLibraryName    = "otel.library.name"
	OtelLibraryVersion = "otel.library.version"

	DeploymentEnvironment = "deployment.environment"

	ServiceName       = "service.name"
	ServiceVersion    = "service.version"
	ServiceNamespace  = "service.namespace"
	ServiceInstanceID = "service.instance.id"
	PeerService       = "peer.service"

	HostID           = "host.id"
	HostName         = "host.name"
	HostType         = "host.type"
	HostArch         = "host.arch"
	HostImageName    = "host.image.name"
	HostImageID      = "host.image.id"
	HostImageVersion = "host.image.version"

	ServerAddress       = "server.address"
	ServerPort          = "server.port"
	ServerSocketDomain  = "server.socket.domain"
	ServerSocketAddress = "server.socket.address"
	ServerSocketPort    = "server.socket.port"

	ClientAddress       = "client.address"
	ClientPort          = "client.port"
	ClientSocketAddress = "client.socket.address"
	ClientSocketPort    = "client.socket.port"

	URLScheme   = "url.scheme"
	URLFull     = "url.full"
	URLPath     = "url.path"
	URLQuery    = "url.query"
	URLFragment = "url.fragment"

	HTTPRequestMethod       = "http.request.method"
	HTTPRequestBodySize     = "http.request.body.size"
	HTTPResponseBodySize    = "http.response.body.size"
	HTTPResponseStatusCode  = "http.response.status_code"
	HTTPResponseStatusClass = "http.response.status_class"
	HTTPRoute               = "http.route"

	RPCSystem  = "rpc.system"
	RPCService = "rpc.service"
	RPCMethod  = "rpc.method"

	UserAgentOriginal  = "user_agent.original"
	UserAgentName      = "user_agent.name"
	UserAgentVersion   = "user_agent.version"
	UserAgentOSName    = "user_agent.os_name"
	UserAgentOSVersion = "user_agent.os_version"
	UserAgentDevice    = "user_agent.device"
	UserAgentIsBot     = "user_agent.is_bot"

	DBSystem    = "db.system"
	DBName      = "db.name"
	DBStatement = "db.statement"
	DBOperation = "db.operation"
	DBSqlTable  = "db.sql.table"

	EnduserID    = "enduser.id"
	EnduserRole  = "enduser.role"
	EnduserScope = "enduser.scope"

	LogMessage        = "log.message"
	LogSeverity       = "log.severity"
	LogSeverityNumber = "log.severity_number"
	LogSource         = "log.source"
	LogFilePath       = "log.file.path"
	LogFileName       = "log.file.name"

	ExceptionType       = "exception.type"
	ExceptionMessage    = "exception.message"
	ExceptionStacktrace = "exception.stacktrace"

	CodeFunction = "code.function"
	CodeFilepath = "code.filepath"

	MessagingSystem                            = "messaging.system"
	MessagingOperation                         = "messaging.operation"
	MessagingDestinationName                   = "messaging.destination.name"
	MessagingDestinationKind                   = "messaging.destination.kind"
	MessagingDestinationTemporary              = "messaging.destination.temporary"
	MessagingMessageID                         = "messaging.message.id"
	MessagingMessageType                       = "messaging.message.type" // TODO: remove
	MessagingMessagePayloadSizeBytes           = "messaging.message.payload_size_bytes"
	MessagingMessagePayloadCompressedSizeBytes = "messaging.message.payload_compressed_size_bytes"

	CloudProvider         = "cloud.provider"
	CloudAccountID        = "cloud.account.id"
	CloudRegion           = "cloud.region"
	CloudResourceID       = "cloud.resource_id"
	CloudAvailabilityZone = "cloud.availability_zone"
	CloudPlatform         = "cloud.platform"
)
