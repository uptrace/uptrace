package bunconv

import (
	"math"
	"strings"
	"time"
)

func FormatMicroseconds(n float64) string {
	if n == 0 {
		return "0"
	}

	abs := math.Abs(n)

	if abs < 1000 {
		return format(n, 0) + "µs"
	}

	n /= 1000
	abs /= 1000

	if abs < 10 {
		return format(n, 1) + "ms"
	}
	if abs < 1000 {
		return format(n, 0) + "ms"
	}

	n /= 1000
	abs /= 1000

	if abs < 1 {
		return format(n, 2) + "s"
	}
	if abs < 10 {
		return format(n, 1) + "s"
	}
	return format(n, 0) + "s"
}

func ShortDuration(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}
