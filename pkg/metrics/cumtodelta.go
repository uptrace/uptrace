package metrics

import (
	"sync"
	"time"

	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/zyedidia/generic/cache"
)

type MeasureKey struct {
	ProjectID         uint32
	Metric            string
	AttrsHash         uint64
	StartTimeUnixNano uint64
}

type MeasureValue struct {
	Key   MeasureKey
	Point any
	Time  time.Time
}

type CumToDeltaConv struct {
	cap int

	mu    sync.Mutex
	cache *cache.Cache[MeasureKey, *MeasureValue]
}

func NewCumToDeltaConv(n int) *CumToDeltaConv {
	c := &CumToDeltaConv{
		cache: cache.New[MeasureKey, *MeasureValue](n),
	}
	return c
}

func (c *CumToDeltaConv) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.cache.Size()
}

func (c *CumToDeltaConv) SwapPoint(key MeasureKey, point any, time time.Time) any {
	c.mu.Lock()
	defer c.mu.Unlock()

	if value, ok := c.cache.Get(key); ok {
		if time.Before(value.Time) {
			return nil
		}

		prevPoint := value.Point
		value.Point = point
		value.Time = time
		return prevPoint
	}

	c.cache.Put(key, &MeasureValue{
		Point: point,
		Time:  time,
	})
	return nil
}

//------------------------------------------------------------------------------

type NumberPoint struct {
	Int    int64
	Double float64
}

func NewIntPoint(n int64) *NumberPoint {
	return &NumberPoint{
		Int: n,
	}
}

func NewDoublePoint(n float64) *NumberPoint {
	return &NumberPoint{
		Double: n,
	}
}

type HistogramPoint struct {
	Sum          float64
	Count        uint64
	Bounds       []float64
	BucketCounts []uint64
}

type ExpHistogramPoint struct {
	Sum       float64
	Count     uint64
	Histogram map[bfloat16.T]uint64
}
