package bununit

import "math"

func FormatPercents(n float64) string {
	switch {
	case math.IsNaN(n):
		return "0%"
	case math.IsInf(n, +1):
		return "+Inf"
	case math.IsInf(n, -1):
		return "-Inf"
	}

	abs := math.Abs(n)
	if abs < 0.001 {
		return "0%"
	}

	if abs < 1 {
		return format(n, 2) + "%"
	}
	if abs < 10 {
		return format(n, 1) + "%"
	}
	return format(n, 0) + "%"
}

func FormatUtilization(n float64) string {
	switch {
	case math.IsNaN(n):
		return "0%"
	case math.IsInf(n, +1):
		return "+Inf"
	case math.IsInf(n, -1):
		return "-Inf"
	}

	abs := math.Abs(n)
	if abs < 0.001 {
		return "0%"
	}

	n *= 100
	abs *= 100

	if abs < 1 {
		return format(n, 2) + "%"
	}
	if abs < 10 {
		return format(n, 1) + "%"
	}
	return format(n, 0) + "%"
}
