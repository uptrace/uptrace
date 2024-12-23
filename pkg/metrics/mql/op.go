package mql

import (
	"math"
	"slices"
	"time"

	"github.com/uptrace/pkg/unixtime"
	"gonum.org/v1/gonum/stat"
)

const (
	CHAggNone = "_"

	CHAggMin    = "min"
	CHAggMax    = "max"
	CHAggSum    = "sum"
	CHAggAvg    = "avg"
	CHAggMedian = "median"

	CHAggUniq = "uniq"

	// Histograms only.
	CHAggCount = "count"
	CHAggP50   = "p50"
	CHAggP75   = "p75"
	CHAggP90   = "p90"
	CHAggP95   = "p95"
	CHAggP99   = "p99"
)

const (
	GoAggMin    = "min"
	GoAggMax    = "max"
	GoAggSum    = "sum"
	GoAggAvg    = "avg"
	GoAggMedian = "median"
)

const (
	TableFuncMin    = "min"
	TableFuncMax    = "max"
	TableFuncSum    = "sum"
	TableFuncAvg    = "avg"
	TableFuncMedian = "median"
	TableFuncLast   = "last"
)

//------------------------------------------------------------------------------

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

const (
	TransformPerMin = "per_min"
	TransformPerSec = "per_sec"
	TransformAbs    = "abs"
	TransformCeil   = "ceil"
	TransformFloor  = "floor"
	TransformTrunc  = "trunc"
	TransformCos    = "cos"
	TransformCosh   = "cosh"
	TransformAcos   = "acos"
	TransformAcosh  = "acosh"
	TransformSin    = "sin"
	TransformSinh   = "sinh"
	TransformAsin   = "asin"
	TransformAsinh  = "asinh"
	TransformTan    = "tan"
	TransformTanh   = "tanh"
	TransformAtan   = "atan"
	TransformAtanh  = "atanh"
	TransformExp    = "exp"
	TransformExp2   = "exp2"
	TransformLog    = "log"
	TransformLog2   = "log2"
	TransformLog10  = "log10"
	TransformLn     = "ln"
)

type transformFunc func(value []float64)

func (e *Engine) perMinTransform(value []float64) {
	minutes := e.interval.Minutes()
	for i, num := range value {
		value[i] = num / minutes
	}
}

func (e *Engine) perSecTransform(value []float64) {
	seconds := e.interval.Seconds()
	for i, num := range value {
		value[i] = num / seconds
	}
}

func absTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Abs(num)
	}
}

func ceilTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Ceil(num)
	}
}

func floorTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Floor(num)
	}
}

func truncTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Trunc(num)
	}
}

func cosTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Cos(num)
	}
}

func coshTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Cosh(num)
	}
}

func acosTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Acos(num)
	}
}

func acoshTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Acosh(num)
	}
}

func sinTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Sin(num)
	}
}

func sinhTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Sinh(num)
	}
}

func asinTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Asin(num)
	}
}

func asinhTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Asinh(num)
	}
}

func tanTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Tan(num)
	}
}

func tanhTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Tanh(num)
	}
}

func atanTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Atan(num)
	}
}

func atanhTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Atanh(num)
	}
}

func expTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Exp(num)
	}
}

func exp2Transform(value []float64) {
	for i, num := range value {
		value[i] = math.Exp2(num)
	}
}

func logTransform(value []float64) {
	for i, num := range value {
		value[i] = math.Log(num)
	}
}

func log2Transform(value []float64) {
	for i, num := range value {
		value[i] = math.Log2(num)
	}
}

func log10Transform(value []float64) {
	for i, num := range value {
		value[i] = math.Log10(num)
	}
}

//------------------------------------------------------------------------------

const (
	RollupIncrease       = "increase"
	RollupDelta          = "delta"
	RollupRate           = "rate"
	RollupIRate          = "irate"
	RollupMinOverTime    = "min_over_time"
	RollupMaxOverTime    = "max_over_time"
	RollupSumOverTime    = "sum_over_time"
	RollupAvgOverTime    = "avg_over_time"
	RollupMedianOverTime = "median_over_time"
)

type rollupFunc func(dest, src []float64, tm []unixtime.Nano, window time.Duration)

func (e *Engine) increaseRollup(
	dest, src []float64, timeSlice []unixtime.Nano, windowDur time.Duration,
) {
	rateWindow := max(windowDur, 5*e.interval, minRateWindow)
	rateRollup(dest, src, timeSlice, rateWindow)

	windowSeconds := windowDur.Seconds()
	for i, currValue := range dest {
		dest[i] = currValue * windowSeconds
	}
}

func rateRollup(dest, src []float64, timeSlice []unixtime.Nano, windowDur time.Duration) {
	windowSeconds := unixtime.Nano(windowDur.Seconds())
	for i, currValue := range src {
		if math.IsNaN(currValue) {
			dest[i] = math.NaN()
			continue
		}
		currTime := timeSlice[i]

		{
			before := currTime - windowSeconds
			prevValue, prevTime := prevPoint(src[:i], timeSlice[:i], currValue, before)
			if prevTime > 0 {
				dest[i] = (currValue - prevValue) / float64(currTime-prevTime)
				continue
			}
		}

		{
			after := currTime + windowSeconds
			nextValue, nextTime := nextPoint(src[i+1:], timeSlice[i+1:], currValue, after)
			if nextTime > 0 {
				dest[i] = (nextValue - currValue) / float64(nextTime-currTime)
				continue
			}
		}

		dest[i] = 0
	}
}

func prevPoint(
	src []float64,
	timeSlice []unixtime.Nano,
	currValue float64,
	before unixtime.Nano,
) (float64, unixtime.Nano) {
	lastValue := currValue
	var lastTime unixtime.Nano
	for i := len(src) - 1; i >= 0; i-- {
		pointValue := src[i]
		pointTime := timeSlice[i]

		if math.IsNaN(pointValue) || pointValue > lastValue {
			// Counter reset.
			return lastValue, lastTime
		}

		if pointTime <= before {
			return pointValue, pointTime
		}
		lastValue = pointValue
		lastTime = pointTime
	}
	return lastValue, lastTime
}

func nextPoint(
	src []float64,
	timeSlice []unixtime.Nano,
	currValue float64,
	after unixtime.Nano,
) (float64, unixtime.Nano) {
	lastValue := currValue
	var lastTime unixtime.Nano
	for i, pointValue := range src {
		pointTime := timeSlice[i]

		if math.IsNaN(pointValue) || pointValue < lastValue {
			// Counter reset.
			return lastValue, lastTime
		}

		if pointTime >= after {
			return pointValue, pointTime
		}
		lastValue = pointValue
		lastTime = pointTime
	}
	return lastValue, lastTime
}

func minRollup(dest, src []float64, timeSlice []unixtime.Nano, windowDur time.Duration) {
	overTime(dest, src, timeSlice, windowDur, minAgg)
}

func maxRollup(dest, src []float64, timeSlice []unixtime.Nano, windowDur time.Duration) {
	overTime(dest, src, timeSlice, windowDur, maxAgg)
}

func sumRollup(dest, src []float64, timeSlice []unixtime.Nano, windowDur time.Duration) {
	overTime(dest, src, timeSlice, windowDur, sumAgg)
}

func avgRollup(dest, src []float64, timeSlice []unixtime.Nano, windowDur time.Duration) {
	overTime(dest, src, timeSlice, windowDur, avgAgg)
}

func medianRollup(dest, src []float64, timeSlice []unixtime.Nano, windowDur time.Duration) {
	overTime(dest, src, timeSlice, windowDur, Median)
}

func overTime(
	dest, src []float64, timeSlice []unixtime.Nano, windowDur time.Duration, fn aggFunc,
) {
	windowSeconds := unixtime.Nano(windowDur.Seconds())
	for i := len(timeSlice) - 1; i >= 0; i-- {
		currTime := timeSlice[i]
		window := window(src[:i+1], timeSlice[:i+1], currTime-windowSeconds)
		dest[i] = fn(window)
	}
}

func window(
	src []float64, timeSlice []unixtime.Nano, wantedTime unixtime.Nano,
) []float64 {
	for i := len(timeSlice) - 2; i >= 0; i-- {
		currTime := timeSlice[i]
		if currTime <= wantedTime {
			return src[i:]
		}
	}
	return src
}

//------------------------------------------------------------------------------

type aggFunc func(value []float64) float64

func minAgg(value []float64) float64 {
	return slices.Min(value)
}

func maxAgg(value []float64) float64 {
	return slices.Max(value)
}

func sumAgg(value []float64) float64 {
	sum, _ := sumCount(value)
	return sum
}

func avgAgg(value []float64) float64 {
	sum, count := sumCount(value)
	return sum / float64(count)
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

func Median(value []float64) float64 {
	value = slices.Clone(value)

	// Zero nans because stat.Quantile does not support them.
	for i, num := range value {
		if math.IsNaN(num) {
			value[i] = 0
		}
	}

	slices.Sort(value)
	return stat.Quantile(0.5, stat.Empirical, value, nil)
}

//------------------------------------------------------------------------------

func nan(f float64) float64 {
	if math.IsNaN(f) {
		return 0
	}
	return f
}
