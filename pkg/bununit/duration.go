package bununit

import (
	"math"
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
