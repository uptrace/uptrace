package tracing

import "slices"

const emptyPlaceholder = "<empty>"

const (
	TypeSpanFuncs      = "funcs"
	TypeSpanHTTPServer = "httpserver"
	TypeSpanHTTPClient = "httpclient"
	TypeSpanDB         = "db"
	TypeSpanRPC        = "rpc"
	TypeSpanMessaging  = "messaging"
	TypeSpanFAAS       = "faas"

	TypeLog          = "log"
	TypeEventMessage = "message"
	TypeEventOther   = "other-events"
)

const (
	SystemUnknown = "unknown"

	SystemAll       = "all"
	SystemSpansAll  = "spans:all"
	SystemEventsAll = "events:all"

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

var spanTypeEnum = []string{
	TypeSpanFuncs,
	TypeSpanHTTPServer,
	TypeSpanDB,
	TypeSpanRPC,
	TypeSpanMessaging,
	TypeSpanFAAS,
	TypeSpanHTTPClient,

	TypeLog,
	TypeEventMessage,
	TypeEventOther,
}

var (
	LogTypes   = []string{TypeLog}
	EventTypes = []string{TypeEventMessage, TypeEventOther}
	SpanTypes  []string // filled in init
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

func init() {
	for _, typ := range spanTypeEnum {
		if slices.Contains(LogTypes, typ) {
			continue
		}
		if slices.Contains(EventTypes, typ) {
			continue
		}
		SpanTypes = append(SpanTypes, typ)
	}
}
