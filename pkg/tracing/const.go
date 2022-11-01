package tracing

import (
	"strings"

	"github.com/uptrace/uptrace/pkg/bunotel"
)

const (
	SystemAll       = "all"
	SystemEventsAll = "events:all"
	SystemSpansAll  = "spans:all"

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

var eventSystems = []string{
	SystemEventLog + ":" + strings.ToLower(bunotel.TraceSeverity),
	SystemEventLog + ":" + strings.ToLower(bunotel.DebugSeverity),
	SystemEventLog + ":" + strings.ToLower(bunotel.InfoSeverity),
	SystemEventLog + ":" + strings.ToLower(bunotel.WarnSeverity),
	SystemEventLog + ":" + strings.ToLower(bunotel.ErrorSeverity),
	SystemEventLog + ":" + strings.ToLower(bunotel.FatalSeverity),
	SystemEventLog + ":" + strings.ToLower(bunotel.PanicSeverity),
	SystemEventExceptions,
	SystemEventOther,
}
