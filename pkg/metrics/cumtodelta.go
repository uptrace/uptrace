package metrics

import (
	"sync"
	"time"

	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/zyedidia/generic/list"
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

	mu   sync.Mutex
	mp   map[MeasureKey]*list.Node[MeasureValue]
	list *list.List[MeasureValue]
}

func NewCumToDeltaConv(n int) *CumToDeltaConv {
	c := &CumToDeltaConv{
		cap: n,

		mp:   make(map[MeasureKey]*list.Node[MeasureValue], n),
		list: list.New[MeasureValue](),
	}
	return c
}

func (c *CumToDeltaConv) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.mp)
}

func (c *CumToDeltaConv) SwapPoint(key MeasureKey, point any, time time.Time) any {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, ok := c.mp[key]; ok {
		c.list.Remove(node)
		c.list.PushFrontNode(node)

		if time.Before(node.Value.Time) {
			return nil
		}

		prevPoint := node.Value.Point
		node.Value.Point = point
		node.Value.Time = time
		return prevPoint
	}

	if len(c.mp) < c.cap {
		c.list.PushFront(MeasureValue{
			Key:   key,
			Point: point,
			Time:  time,
		})
		c.mp[key] = c.list.Front
		return nil
	}

	back := c.list.Back

	c.list.Remove(back)
	c.list.PushFrontNode(back)

	delete(c.mp, back.Value.Key)

	back.Value.Key = key
	back.Value.Point = point
	back.Value.Time = time
	c.mp[key] = back

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
	Min          float64
	Max          float64
	Sum          float64
	Count        uint64
	Bounds       []float64
	BucketCounts []uint64
}

type ExpHistogramPoint struct {
	Min       float64
	Max       float64
	Sum       float64
	Count     uint64
	Histogram map[bfloat16.T]uint64
}
