package unixtime

import (
	"fmt"
	"github.com/segmentio/encoding/iso8601"
	"strings"
	"time"
)

func ParseTime(s string) (time.Time, error) {
	format := DiscoverFormat(s)
	if format == "" {
		return time.Time{}, unknownFormatError{s: s}
	}
	switch format {
	case time.RFC3339:
		return iso8601.Parse(s)
	case goFormat:
		if i := strings.Index(s, " m="); i >= 0 {
			s = s[:i]
		}
	}
	return time.Parse(format, s)
}

type unknownFormatError struct{ s string }

func (e unknownFormatError) Error() string { return fmt.Sprintf("unknown time format: %q", e.s) }
func DiscoverFormat(s string) string {
	pos, format := ReadTime(s)
	if pos == len(s) {
		return format
	}
	return ""
}
