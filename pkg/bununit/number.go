package bununit

import (
	"fmt"
	"math"
	"strconv"
)

func FormatNumber(n float64) string {
	if math.IsNaN(n) || math.IsInf(n, 0) || n == 0 {
		return "0"
	}

	abs := math.Abs(n)

	if abs < 0.01 {
		return round(n, 3)
	}
	if abs < 0.1 {
		return round(n, 2)
	}
	if abs < 100 {
		return round(n, 1)
	}
	if abs < 1000 {
		return round(n, 0)
	}

	n /= 1000
	abs /= 1000

	if abs < 100 {
		return round(n, 1) + "k"
	}
	if abs < 1000 {
		return round(n, 0) + "k"
	}

	n /= 1000
	abs /= 1000

	if abs < 100 {
		return round(n, 1) + "m"
	}
	if abs < 1000 {
		return round(n, 0) + "m"
	}

	n /= 1000
	return round(n, 1) + "b"
}

//------------------------------------------------------------------------------

func FormatFloat(n float64) string {
	if n == 0 {
		return "0"
	}

	abs := math.Abs(n)

	if abs < 0.01 {
		return round(n, 3)
	}
	if abs < 0.1 {
		return round(n, 2)
	}
	if abs < 100 {
		return round(n, 1)
	}
	return round(n, 0)
}

//------------------------------------------------------------------------------

func round(f float64, mantissa int) string {
	f = roundFloat(f, mantissa)
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func format(n float64, mantissa int) string {
	format := getFormat(mantissa)
	return fmt.Sprintf(format, n)
}

func getFormat(mantissa int) string {
	b := make([]byte, 0, 8)
	b = append(b, '%')
	b = append(b, '.')
	b = strconv.AppendInt(b, int64(mantissa), 10)
	b = append(b, 'f')
	return string(b)
}

func roundFloat(f float64, mantissa int) float64 {
	pow := math.Pow(10, float64(mantissa))
	return math.Round(f*pow) / pow
}
