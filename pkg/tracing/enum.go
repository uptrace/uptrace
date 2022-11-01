package tracing

const (
	SystemAll       = "all"
	SystemAllEvents = "all:events"
	SystemAllSpans  = "all:spans"

	SystemSpanFuncs     = "funcs"
	SystemSpanHTTP      = "http"
	SystemSpanDB        = "db"
	SystemSpanRPC       = "rpc"
	SystemSpanMessaging = "messaging"

	SystemEventLog        = "log"
	SystemEventExceptions = "exceptions"
	SystemEventMessage    = "message"
	SystemEventOther      = "other-events"
)
