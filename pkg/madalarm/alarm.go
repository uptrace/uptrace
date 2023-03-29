package madalarm

import (
	"fmt"
	"math"

	"github.com/uptrace/uptrace/pkg/bunutil"
)

func Check(in []float64, opts ...Option) (*CheckResult, error) {
	conf := &config{
		duration: 1,
	}
	for _, opt := range opts {
		opt(conf)
	}

	if conf.duration == 0 {
		return nil, fmt.Errorf("duration can't be zero")
	}

	check := &CheckResult{
		Input: in,
	}

	if len(in) < conf.duration {
		return check, nil
	}

	var numHole int
	for i, n := range in {
		if math.IsNaN(n) {
			in[i] = 0
			numHole++
		}
	}

	check.Values = in[len(in)-conf.duration:]

	if !conf.minValue.Valid && !conf.maxValue.Valid {
		return nil, fmt.Errorf("at least min or max value is required")
	}

	check.Bounds.Min = conf.minValue
	check.Bounds.Max = conf.maxValue

	check.Firing, check.FiringFor = checkValues(
		check.Values, check.Bounds.Min, check.Bounds.Max)
	if check.Firing != 0 {
		check.Outlier = check.Values[0]
	}

	return check, nil
}

type CheckResult struct {
	Input  []float64 `json:"input"`
	Values []float64 `json:"values"`

	Bounds Bounds `json:"bounds"`

	Firing    int     `json:"firing"`
	FiringFor int     `json:"firingFor"`
	Outlier   float64 `json:"outlier"`
}

type Bounds struct {
	Min bunutil.NullFloat64 `json:"min"`
	Max bunutil.NullFloat64 `json:"max"`
}

func (r *CheckResult) IsOutlier(x float64) int {
	return isOutlier(x, r.Bounds.Min, r.Bounds.Max)
}

func isOutlier(x float64, min, max bunutil.NullFloat64) int {
	if min.Valid && x < min.Float64 {
		return -1 // outlier to the left
	}
	if max.Valid && x > max.Float64 {
		return 1 // outlier to the right
	}
	return 0 // not an outlier
}

func checkValues(values []float64, min, max bunutil.NullFloat64) (firing, duration int) {
	for i := len(values) - 1; i >= 0; i-- {
		value := values[i]

		flag := isOutlier(value, min, max)
		if flag == 0 {
			return 0, duration
		}

		duration++

		if firing == 0 {
			firing = flag
			continue
		}
		if firing != flag {
			return 0, duration
		}
	}
	return firing, duration
}
