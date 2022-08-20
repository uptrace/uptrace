package bunutil

import (
	"time"

	"golang.org/x/exp/constraints"
)

func FillHoles(m map[string]any, gte, lt time.Time, interval time.Duration) {
	if len(m) == 0 {
		return
	}

	times, ok := findTimeColumn(m)
	if !ok {
		return
	}

	for k, v := range m {
		switch v := v.(type) {
		case []uint32:
			m[k] = FillOrdered(v, times, gte, lt, interval)
		case []uint64:
			m[k] = FillOrdered(v, times, gte, lt, interval)
		case []int64:
			m[k] = FillOrdered(v, times, gte, lt, interval)
		case []float32:
			m[k] = FillOrdered(v, times, gte, lt, interval)
		case []float64:
			m[k] = FillOrdered(v, times, gte, lt, interval)
		case []time.Time:
			m[k] = FillTime(v, gte, lt, interval)
		}
	}
}

func findTimeColumn(m map[string]any) ([]time.Time, bool) {
	for _, key := range []string{"time", "item.time", "span.time"} {
		if v, ok := m[key].([]time.Time); ok {
			return v, true
		}
	}
	return nil, false
}

func FillOrdered[T constraints.Ordered](
	nums []T,
	times []time.Time,
	gte, lt time.Time,
	interval time.Duration,
) []T {
	numItem := numItem(gte, lt, interval)
	if len(nums) == numItem {
		return nums
	}

	filled := make([]T, numItem)

	for i, num := range nums {
		index := int(times[i].Sub(gte) / interval)
		if index < 0 || index >= numItem {
			return nums
		}
		filled[index] = num
	}

	return filled
}

func FillTime(
	times []time.Time,
	gte, lt time.Time,
	interval time.Duration,
) []time.Time {
	numItem := numItem(gte, lt, interval)
	if len(times) == numItem {
		return times
	}

	for _, tm := range times {
		if tm.Before(gte) || tm.After(lt) {
			return times
		}
	}

	filled := make([]time.Time, numItem)
	for i := range filled {
		filled[i] = gte.Add(time.Duration(i) * interval)
	}
	return filled
}

func numItem(gte, lt time.Time, interval time.Duration) int {
	return int(lt.Sub(gte) / interval)
}
