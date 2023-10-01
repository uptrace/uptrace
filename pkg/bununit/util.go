package bununit

import "math"

func Round(f float64, mantissa int) float64 {
	pow := math.Pow(10, float64(mantissa))
	return math.Round(f*pow) / pow
}
