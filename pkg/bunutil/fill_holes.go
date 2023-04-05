package bunutil

import (
	"time"
)

func FillHoles(m map[string]any, gte, lt time.Time, interval time.Duration) {
	if len(m) == 0 {
		return
	}

	timeKey, timeCol := findTimeColumn(m)
	if timeKey == "" {
		return
	}

	for k, v := range m {
		switch v := v.(type) {
		case []uint32:
			m[k] = Fill(v, timeCol, 0, gte, lt, interval)
		case []uint64:
			m[k] = Fill(v, timeCol, 0, gte, lt, interval)
		case []int64:
			m[k] = Fill(v, timeCol, 0, gte, lt, interval)
		case []float32:
			m[k] = Fill(v, timeCol, 0, gte, lt, interval)
		case []float64:
			m[k] = Fill(v, timeCol, 0, gte, lt, interval)
		}
	}

	m[timeKey] = FillTime(timeCol, gte, lt, interval)
}

func findTimeColumn(m map[string]any) (string, []time.Time) {
	for _, key := range []string{"_time", "time", "item.time", "span.time"} {
		if v, ok := m[key].([]time.Time); ok {
			return key, v
		}
	}
	return "", nil
}

func Fill[T any](
	values []T,
	timeCol []time.Time,
	value T,
	gte, lt time.Time,
	interval time.Duration,
) []T {
	if len(values) != len(timeCol) {
		return values
	}

	numItem := numItem(gte, lt, interval)
	if len(values) == numItem {
		return values
	}

	filled := make([]T, numItem)
	for i := range filled {
		filled[i] = value
	}

	for i, num := range values {
		index := int(timeCol[i].Sub(gte) / interval)
		if index < 0 || index >= numItem {
			return values
		}
		filled[index] = num
	}

	return filled
}

func FillTime(
	timeCol []time.Time,
	gte, lt time.Time,
	interval time.Duration,
) []time.Time {
	numItem := numItem(gte, lt, interval)
	if len(timeCol) == numItem {
		return timeCol
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
