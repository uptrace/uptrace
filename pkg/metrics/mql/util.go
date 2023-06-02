package mql

import (
	"math"

	"golang.org/x/exp/constraints"
)

func max[T constraints.Ordered](a, b T) T {
	if a >= b {
		return a
	}
	return b
}

func nan(f float64) float64 {
	if math.IsNaN(f) {
		return 0
	}
	return f
}
