package madalarm

import "github.com/uptrace/uptrace/pkg/bunutil"

type config struct {
	duration int
	bounds   *Bounds
}

type Option func(c *config)

func WithDuration(n int) Option {
	return func(c *config) {
		c.duration = n
	}
}

func WithMinValue(min float64) Option {
	return func(c *config) {
		if c.bounds == nil {
			c.bounds = NewBounds()
		}
		c.bounds.Min = bunutil.NullFloat64{
			Float64: min,
			Valid:   true,
		}
	}
}

func WithMaxValue(max float64) Option {
	return func(c *config) {
		if c.bounds == nil {
			c.bounds = NewBounds()
		}
		c.bounds.Max = bunutil.NullFloat64{
			Float64: max,
			Valid:   true,
		}
	}
}

type Bounds struct {
	Min bunutil.NullFloat64 `json:"min"`
	Max bunutil.NullFloat64 `json:"max"`
}

func NewBounds() *Bounds {
	return &Bounds{}
}

func (b *Bounds) IsOutlier(x float64) int {
	return isOutlier(x, b.Min, b.Max)
}
