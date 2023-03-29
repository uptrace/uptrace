package bununit

import "math"

func FormatPercents(n float64) string {
	return percents(n, false)
}

func FormatPercentsSign(n float64) string {
	return percents(n, true)
}

func percents(n float64, sign bool) string {
	switch {
	case math.IsNaN(n):
		return "0%"
	case math.IsInf(n, +1):
		n = 1
	case math.IsInf(n, -1):
		n = -1
	case n <= -10:
		return format(-999, 0, sign) + "%"
	case n >= 10:
		return format(999, 0, sign) + "%"
	}

	abs := math.Abs(n)
	if abs < 0.001 {
		return "0%"
	}

	n *= 100
	abs *= 100

	if abs < 1 {
		return format(n, 2, sign) + "%"
	}
	if abs < 10 {
		return format(n, 1, sign) + "%"
	}
	return format(n, 0, sign) + "%"
}
