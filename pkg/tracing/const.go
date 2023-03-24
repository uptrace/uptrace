package tracing

const (
	SpanTypeFuncs     = "funcs"
	SpanTypeHTTP      = "http"
	SpanTypeDB        = "db"
	SpanTypeRPC       = "rpc"
	SpanTypeMessaging = "messaging"
	SpanTypeFAAS      = "faas"

	EventTypeLog        = "log"
	EventTypeExceptions = "exceptions"
	EventTypeMessage    = "message"
	EventTypeOther      = "other-events"
)

const (
	SystemUnknown = "unknown"

	SystemAll       = "all"
	SystemEventsAll = "events:all"
	SystemSpansAll  = "spans:all"

	SystemLogError = "log:error"
	SystemLogFatal = "log:fatal"
	SystemLogPanic = "log:panic"
)

const (
	otelEventLog       = "log"
	otelEventException = "exception"
	otelEventMessage   = "message"
)

var (
	EventTypes   = []string{EventTypeLog, EventTypeExceptions, EventTypeMessage, EventTypeOther}
	ErrorTypes   = []string{EventTypeLog, EventTypeExceptions}
	ErrorSystems = []string{EventTypeExceptions, SystemLogError, SystemLogFatal, SystemLogPanic}
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
