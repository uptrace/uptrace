package madalarm

import "github.com/uptrace/uptrace/pkg/bunutil"

type config struct {
	duration int

	// Manual
	minValue bunutil.NullFloat64
	maxValue bunutil.NullFloat64
}

type Option func(c *config)

func WithDuration(n int) Option {
	return func(c *config) {
		c.duration = n
	}
}

func WithMinValue(min float64) Option {
	return func(c *config) {
		c.minValue.Float64 = min
		c.minValue.Valid = true
	}
}

func WithMaxValue(max float64) Option {
	return func(c *config) {
		c.maxValue.Float64 = max
		c.maxValue.Valid = true
	}
}
