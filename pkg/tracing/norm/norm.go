package norm

import (
	"strings"
)

const (
	SeverityTrace  = "TRACE"
	SeverityTrace2 = "TRACE2"
	SeverityTrace3 = "TRACE3"
	SeverityTrace4 = "TRACE4"
	SeverityDebug  = "DEBUG"
	SeverityDebug2 = "DEBUG2"
	SeverityDebug3 = "DEBUG3"
	SeverityDebug4 = "DEBUG4"
	SeverityInfo   = "INFO"
	SeverityInfo2  = "INFO2"
	SeverityInfo3  = "INFO3"
	SeverityInfo4  = "INFO4"
	SeverityWarn   = "WARN"
	SeverityWarn2  = "WARN2"
	SeverityWarn3  = "WARN3"
	SeverityWarn4  = "WARN4"
	SeverityError  = "ERROR"
	SeverityError2 = "ERROR2"
	SeverityError3 = "ERROR3"
	SeverityError4 = "ERROR4"
	SeverityFatal  = "FATAL"
	SeverityFatal2 = "FATAL2"
	SeverityFatal3 = "FATAL3"
	SeverityFatal4 = "FATAL4"
	SeverityPanic  = "PANIC"
	SeverityPanic2 = "PANIC2"
	SeverityPanic3 = "PANIC3"
	SeverityPanic4 = "PANIC4"
)

func LogSeverity(s string) string {
	if s := logSeverity(s); s != "" {
		return s
	}
	return logSeverity(strings.ToLower(s))
}

func logSeverity(s string) string {
	switch s {
	case "trace":
		return SeverityTrace
	case "debug":
		return SeverityDebug
	case "info", "information", "notice", "log", "usage":
		return SeverityInfo
	case "warn", "warning":
		return SeverityWarn
	case "error", "err", "alert":
		return SeverityError
	case "fatal", "crit", "critical", "emerg", "emergency", "panic":
		return SeverityFatal

	case "trace2":
		return SeverityTrace2
	case "trace3":
		return SeverityTrace3
	case "trace4":
		return SeverityTrace4
	case "debug2":
		return SeverityDebug2
	case "debug3":
		return SeverityDebug3
	case "debug4":
		return SeverityDebug4
	case "info2":
		return SeverityInfo2
	case "info3":
		return SeverityInfo3
	case "info4":
		return SeverityInfo4
	case "warn2":
		return SeverityWarn2
	case "warn3":
		return SeverityWarn3
	case "warn4":
		return SeverityWarn4
	case "error2":
		return SeverityError2
	case "error3":
		return SeverityError3
	case "error4":
		return SeverityError4
	case "fatal2":
		return SeverityFatal2
	case "fatal3":
		return SeverityFatal3
	case "fatal4":
		return SeverityFatal4
	case "panic2":
		return SeverityPanic2
	case "panic3":
		return SeverityPanic3
	case "panic4":
		return SeverityPanic4

	default:
		return ""
	}
}
