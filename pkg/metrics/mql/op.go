package mql

import (
	"math"
	"slices"
)

type binaryOpFunc func(v1, v2 float64) float64

func addOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) && math.IsNaN(v2) {
		return math.NaN()
	}
	return nan(v1) + nan(v2)
}

func subtractOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) && math.IsNaN(v2) {
		return math.NaN()
	}
	return nan(v1) - nan(v2)
}

func multiplyOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return math.NaN()
	}
	return nan(v1) * nan(v2)
}

func divideOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return math.NaN()
	}
	if v2 == 0 {
		return math.Inf(1)
	}
	return v1 / v2
}

func remOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return math.NaN()
	}
	return float64(int64(v1) % int64(v2))
}

func equalOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return math.NaN()
	}
	if v1 == v2 {
		return 1
	}
	return 0
}

func gtOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return math.NaN()
	}
	if v1 > v2 {
		return 1
	}
	return 0
}

func notEqualOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return math.NaN()
	}
	if v1 != v2 {
		return 1
	}
	return 0
}

func gteOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return math.NaN()
	}
	if v1 >= v2 {
		return 1
	}
	return 0
}

func ltOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return math.NaN()
	}
	if v1 < v2 {
		return 1
	}
	return 0
}

func lteOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return math.NaN()
	}
	if v1 < v2 {
		return 1
	}
	return 0
}

func andOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return math.NaN()
	}
	if v1 != 0 && v2 != 0 {
		return v2
	}
	return 0
}

func orOp(v1, v2 float64) float64 {
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return math.NaN()
	}
	if v1 != 0 {
		return v1
	}
	if v2 != 0 {
		return v2
	}
	return 0
}

//------------------------------------------------------------------------------

type FuncOp func(value []float64, consts map[string]float64)

func deltaFunc(value []float64, consts map[string]float64) {
	for i, num := range value {
		if math.IsNaN(num) {
			value[i] = 0
			continue
		}
		value = value[i:]
		break
	}

	if len(value) == 1 { // table mode
		return
	}

	prevNum := value[0]
	value[0] = 0
	value = value[1:]

	for i, num := range value {
		if math.IsNaN(num) {
			value[i] = 0
			continue
		}

		if delta := num - prevNum; delta >= 0 {
			value[i] = delta
		} else {
			value[i] = 0
		}
		prevNum = num
	}
}

func perMinFunc(value []float64, consts map[string]float64) {
	period, ok := consts["_minutes"]
	if !ok {
		return
	}
	for i, num := range value {
		value[i] = num / period
	}
}

func perSecFunc(value []float64, consts map[string]float64) {
	period, ok := consts["_seconds"]
	if !ok {
		return
	}
	for i, num := range value {
		value[i] = num / period
	}
}

func irateFunc(value []float64, consts map[string]float64) {
	deltaFunc(value, consts)
	perSecFunc(value, consts)
}

func noop(value []float64, consts map[string]float64) {}

func nan(f float64) float64 {
	if math.IsNaN(f) {
		return 0
	}
	return f
}

//------------------------------------------------------------------------------

type aggFunc func(value []float64) float64

func minAgg(value []float64) float64 {
	return slices.Min(value)
}

func maxAgg(value []float64) float64 {
	return slices.Max(value)
}

func avgAgg(value []float64) float64 {
	sum, count := sumCount(value)
	return sum / float64(count)
}

func sumAgg(value []float64) float64 {
	sum, _ := sumCount(value)
	return sum
}

func sumCount(value []float64) (sum float64, count int) {
	for _, f := range value {
		if !math.IsNaN(f) {
			sum += f
			count++
		}
	}
	return sum, count
}
