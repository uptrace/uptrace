package upql

import "golang.org/x/exp/constraints"

func max[T constraints.Ordered](a, b T) T {
	if a >= b {
		return a
	}
	return b
}

func setTimeseriesMetric(timeseries []Timeseries, metric string) {
	for i := range timeseries {
		timeseries[i].Metric = metric
	}
}
