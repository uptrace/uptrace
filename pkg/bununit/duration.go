package bununit

import (
	"math"
)

func FormatMicroseconds(n float64) string {
	return microseconds(n, false)
}

func FormatMicrosecondsSign(n float64) string {
	return microseconds(n, true)
}

func microseconds(n float64, sign bool) string {
	if n == 0 {
		return "0"
	}

	abs := math.Abs(n)

	if abs < 1000 {
		return format(n, 0, sign) + "µs"
	}

	n /= 1000
	abs /= 1000

	if abs < 10 {
		return format(n, 1, sign) + "ms"
	}
	if abs < 1000 {
		return format(n, 0, sign) + "ms"
	}

	n /= 1000
	abs /= 1000

	if abs < 1 {
		return format(n, 2, sign) + "s"
	}
	if abs < 10 {
		return format(n, 1, sign) + "s"
	}
	return format(n, 0, sign) + "s"
}
