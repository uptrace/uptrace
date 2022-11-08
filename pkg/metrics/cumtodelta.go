package metrics

import (
	"time"

	"github.com/zyedidia/generic/list"
)

type NumberPoint struct {
	Int    int64
	Double float64
}

func NewIntPoint(n int64) *NumberPoint {
	return &NumberPoint{
		Int: n,
	}
}

type HistogramPoint struct {
	Sum          float64   `msgpack:"id:0"`
	Count        uint64    `msgpack:"id:1"`
	Bounds       []float64 `msgpack:"id:2"`
	BucketCounts []uint64  `msgpack:"id:3"`
}

type ExpHistogramPoint struct {
	Sum       float64             `msgpack:"id:0"`
	Count     uint64              `msgpack:"id:1"`
	Scale     int32               `msgpack:"id:2"`
	ZeroCount uint64              `msgpack:"id:3"`
	Positive  ExpHistogramBuckets `msgpack:"id:4"`
	Negative  ExpHistogramBuckets `msgpack:"id:5"`
}

type ExpHistogramBuckets struct {
	Offset       int32    `msgpack:"id:0"`
	BucketCounts []uint64 `msgpack:"id:1"`
}

type MeasureKey struct {
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
	return len(c.mp)
}

func (c *CumToDeltaConv) Lookup(key MeasureKey, point any, time time.Time) any {
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
