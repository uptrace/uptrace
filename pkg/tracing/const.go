package tracing

const emptyPlaceholder = "<empty>"

const (
	SpanTypeFuncs      = "funcs"
	SpanTypeHTTPServer = "httpserver"
	SpanTypeHTTPClient = "httpclient"
	SpanTypeDB         = "db"
	SpanTypeRPC        = "rpc"
	SpanTypeMessaging  = "messaging"
	SpanTypeFAAS       = "faas"

	EventTypeLog     = "log"
	EventTypeMessage = "message"
	EventTypeOther   = "other-events"
)

const (
	SystemUnknown = "unknown"

	SystemAll       = "all"
	SystemEventsAll = "events:all"
	SystemSpansAll  = "spans:all"

	SystemLogAll   = "log:all"
	SystemLogError = "log:error"
	SystemLogFatal = "log:fatal"
	SystemLogPanic = "log:panic"
)

const (
	otelEventLog       = "log"
	otelEventException = "exception"
	otelEventMessage   = "message"
	otelEventError     = "error"
)

var (
	EventTypes       = []string{EventTypeMessage, EventTypeOther}
	LogAndEventTypes = []string{EventTypeLog, EventTypeMessage, EventTypeOther}
	ErrorTypes       = []string{EventTypeLog}
	ErrorSystems     = []string{SystemLogError, SystemLogFatal, SystemLogPanic}
)

const (
	StatusCodeUnset = "unset"
	StatusCodeError = "error"
	StatusCodeOK    = "ok"
)

const (
	SpanKindInternal = "internal"
	SpanKindServer   = "server"
	SpanKindClient   = "client"
	SpanKindProducer = "producer"
	SpanKindConsumer = "consumer"
)
