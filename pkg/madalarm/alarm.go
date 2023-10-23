package madalarm

import (
	"errors"
	"fmt"
	"math"

	"github.com/uptrace/uptrace/pkg/bunutil"
)

type CheckResult struct {
	Input      []float64 `json:"input"`
	Values     []float64 `json:"values"`
	FirstValue float64   `json:"value"`

	Bounds Bounds `json:"bounds"`

	Firing    int `json:"firing"`
	FiringFor int `json:"firingFor"`
}

func Check(input []float64, opts ...Option) (*CheckResult, error) {
	checker, err := NewChecker(opts...)
	if err != nil {
		return nil, err
	}
	return checker.Check(input, nil)
}

type Checker struct {
	conf *config
}

func NewChecker(opts ...Option) (*Checker, error) {
	conf := &config{
		duration: 1,
	}
	for _, opt := range opts {
		opt(conf)
	}

	if conf.duration == 0 {
		return nil, fmt.Errorf("duration can't be zero")
	}

	if conf.bounds == nil {
		return nil, fmt.Errorf("at least min or max value is required")
	}

	return &Checker{
		conf: conf,
	}, nil
}

func (c *Checker) Bounds() *Bounds {
	return c.conf.bounds
}

func (c *Checker) Check(in []float64, bounds *Bounds) (*CheckResult, error) {
	check := &CheckResult{
		Input: in,
	}

	if len(in) < c.conf.duration {
		return check, nil
	}

	var numHole int
	for i, n := range in {
		if math.IsNaN(n) {
			in[i] = 0
			numHole++
		}
	}

	check.Values = in[len(in)-c.conf.duration:]
	check.FirstValue = check.Values[0]

	if bounds != nil {
		check.Bounds = *bounds
	} else if c.conf.bounds != nil {
		check.Bounds = *c.conf.bounds
	} else {
		return nil, errors.New("madalarm: bounds can't be nil")
	}

	for i := len(check.Values) - 1; i >= 0; i-- {
		x := check.Values[i]
		if check.Bounds.IsOutlier(x) != 0 {
			check.FiringFor++
		}
	}

	if check.FiringFor == c.conf.duration {
		check.Firing = check.Bounds.IsOutlier(check.FirstValue)
	}
	return check, nil
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
