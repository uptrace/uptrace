package norm

import (
	"strings"

	"github.com/uptrace/uptrace/pkg/bunotel"
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
		return bunotel.TraceSeverity
	case "debug":
		return bunotel.DebugSeverity
	case "info", "information", "notice", "log":
		return bunotel.InfoSeverity
	case "warn", "warning":
		return bunotel.WarnSeverity
	case "error", "err", "alert":
		return bunotel.ErrorSeverity
	case "fatal", "crit", "critical", "emerg", "emergency", "panic":
		return bunotel.FatalSeverity
	default:
		return ""
	}
}
