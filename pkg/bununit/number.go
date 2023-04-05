package bununit

import (
	"fmt"
	"math"
	"strconv"
)

func FormatNumber(n float64) string {
	return number(n, false)
}

func FormatNumberSign(n float64) string {
	return number(n, true)
}

func number(n float64, sign bool) string {
	if math.IsNaN(n) || math.IsInf(n, 0) || n == 0 {
		return "0"
	}

	abs := math.Abs(n)

	if abs < 0.01 {
		return round(n, 3, sign)
	}
	if abs < 0.1 {
		return round(n, 2, sign)
	}
	if abs < 100 {
		return round(n, 1, sign)
	}
	if abs < 1000 {
		return round(n, 0, sign)
	}

	n /= 1000
	abs /= 1000

	if abs < 100 {
		return round(n, 1, sign) + "k"
	}
	if abs < 1000 {
		return round(n, 0, sign) + "k"
	}

	n /= 1000
	abs /= 1000

	if abs < 100 {
		return round(n, 1, sign) + "m"
	}
	if abs < 1000 {
		return round(n, 0, sign) + "m"
	}

	n /= 1000
	return round(n, 1, sign) + "b"
}

//------------------------------------------------------------------------------

func FormatFloat(n float64) string {
	return float(n, false)
}

func float(n float64, sign bool) string {
	if n == 0 {
		return "0"
	}

	abs := math.Abs(n)

	if abs < 0.01 {
		return round(n, 3, sign)
	}
	if abs < 0.1 {
		return round(n, 2, sign)
	}
	if abs < 100 {
		return round(n, 1, sign)
	}
	return round(n, 0, sign)
}

//------------------------------------------------------------------------------

func round(f float64, mantissa int, sign bool) string {
	f = roundFloat(f, mantissa)
	if sign && f > 0 {
		return "+" + strconv.FormatFloat(f, 'f', -1, 64)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func format(n float64, mantissa int, sign bool) string {
	format := getFormat(mantissa, sign)
	return fmt.Sprintf(format, n)
}

func getFormat(mantissa int, sign bool) string {
	b := make([]byte, 0, 8)
	b = append(b, '%')
	if sign {
		b = append(b, '+')
	}
	b = append(b, '.')
	b = strconv.AppendInt(b, int64(mantissa), 10)
	b = append(b, 'f')
	return string(b)
}

func roundFloat(f float64, mantissa int) float64 {
	pow := math.Pow(10, float64(mantissa))
	return math.Round(f*pow) / pow
}
