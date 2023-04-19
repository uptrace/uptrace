package metrics

import (
	"math"

	"golang.org/x/exp/constraints"
)

func listToSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func min[T constraints.Ordered](a, b T) T {
	if a <= b {
		return a
	}
	return b
}

func max[T constraints.Ordered](a, b T) T {
	if a >= b {
		return a
	}
	return b
}

func minValue(ns []float64) float64 {
	min := math.MaxFloat64
	for _, n := range ns {
		if math.IsNaN(n) {
			continue
		}
		if n < min {
			min = n
		}
	}
	if min != math.MaxFloat64 {
		return min
	}
	return 0
}

func maxValue(ns []float64) float64 {
	var max float64
	for _, n := range ns {
		if math.IsNaN(n) {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max
}

func last(ns []float64) float64 {
	end := len(ns) - 3
	if end < 0 {
		end = 0
	}

	for i := len(ns) - 1; i >= end; i-- {
		n := ns[i]
		if !math.IsNaN(n) {
			return n
		}
	}
	return 0
}

func avg(ns []float64) float64 {
	sum, count := sumCount(ns)
	return sum / float64(count)
}

func sum(ns []float64) float64 {
	sum, _ := sumCount(ns)
	return sum
}

func sumCount(ns []float64) (float64, int) {
	var sum float64
	var count int
	for _, n := range ns {
		if !math.IsNaN(n) {
			sum += n
			count++
		}
	}
	return sum, count
}

func delta(value []float64) float64 {
	for i, num := range value {
		if !math.IsNaN(num) {
			value = value[i:]
			break
		}
	}

	if len(value) == 0 {
		return 0
	}

	prevNum := value[0]
	value = value[1:]
	var sum float64

	for _, num := range value {
		if math.IsNaN(num) {
			continue
		}
		sum += num - prevNum
		prevNum = num
	}

	return sum
}
