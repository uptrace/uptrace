package attrkey

import (
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

const (
	AlertType = "alert.type"

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

	ServiceName    = "service.name"
	ServiceVersion = "service.version"
	PeerService    = "peer.service"
	HostName       = "host.name"

	OtelLibraryName      = "otel.library.name"
	OtelLibraryVersion   = "otel.library.version"
	TelemetrySDKName     = "telemetry.sdk.name"
	TelemetrySDKVersion  = "telemetry.sdk.version"
	TelemetrySDKLanguage = "telemetry.sdk.language"

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
)

func Clean(s string) string {
	if isClean(s) {
		return s
	}
	return clean(s)
}

func isClean(s string) bool {
	for _, c := range []byte(s) {
		if !isAllowed(c) {
			return false
		}
	}
	return true
}

func clean(s string) string {
	b := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if isAllowed(c) {
			b = append(b, c)
			continue
		}
		if c >= 'A' && c <= 'Z' {
			b = append(b, c+32)
		}
	}
	return unsafeconv.String(b)
}

func isAllowed(c byte) bool {
	return isAlnum(c) || c == '.' || c == '_'
}

func isAlnum(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return c >= 'a' && c <= 'z'
}
