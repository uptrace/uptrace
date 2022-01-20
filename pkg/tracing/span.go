package tracing

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/tracing/xattr"
	"github.com/vmihailenco/msgpack"
	"golang.org/x/exp/slices"
)

type Span struct {
	ProjectID uint32 `json:"projectId"`
	System    string `json:"system" ch:"span.system,lc"`
	GroupID   uint64 `json:"groupId,string" ch:"span.group_id"`

	TraceID  uuid.UUID `json:"traceId" ch:"span.trace_id,type:UUID"`
	ID       uint64    `json:"id,string" ch:"span.id"`
	ParentID uint64    `json:"parentId,string,omitempty" ch:"-"`

	Name string `json:"name" ch:"span.name,lc"`
	Kind string `json:"kind" ch:"span.kind,lc"`

	Time         time.Time     `json:"time" ch:"span.time"`
	Duration     time.Duration `json:"duration" ch:"span.duration"`
	DurationSelf time.Duration `json:"durationSelf" ch:"-"`
	StartPct     float64       `json:"startPct" ch:"-"`

	StatusCode    string `json:"statusCode" ch:"span.status_code,lc"`
	StatusMessage string `json:"statusMessage" ch:"span.status_message"`

	Attrs  AttrMap      `json:"attrs" ch:"-"`
	Events []*SpanEvent `json:"events" ch:"-"`
	Links  []*SpanLink  `json:"links" ch:"-"`

	Children []*Span `json:"children,omitempty" msgpack:"-" ch:"-"`
}

type SpanLink struct {
	TraceID uuid.UUID `json:"traceId"`
	SpanID  uint64    `json:"spanId"`
	Attrs   AttrMap   `json:"attrs"`
}

type SpanEvent struct {
	Name  string    `json:"name"`
	Time  time.Time `json:"time"`
	Attrs AttrMap   `json:"attrs"`
}

var _ ch.AfterScanRowHook = (*Span)(nil)

func (s *Span) AfterScanRow(ctx context.Context) error {
	return nil
}

func (s *Span) EndTime() time.Time {
	return s.Time.Add(time.Duration(s.Duration))
}

func (s *Span) TreeEndTime() time.Time {
	endTime := s.EndTime()
	for _, child := range s.Children {
		tm := child.TreeEndTime()
		if tm.After(endTime) {
			endTime = tm
		}
	}
	return endTime
}

func (s *Span) Walk(fn func(child, parent *Span) error) error {
	if err := fn(s, nil); err != nil {
		return err
	}
	for _, child := range s.Children {
		if err := child.Walk(fn); err != nil {
			return err
		}
	}
	return nil
}

func (s *Span) AddChild(child *Span) {
	s.Children = addSpanSorted(s.Children, child)
}

func addSpanSorted(a []*Span, x *Span) []*Span {
	if len(a) == 0 {
		return []*Span{x}
	}

	for i := len(a) - 1; i >= 0; i-- {
		el := a[i]
		if x.Time.After(el.Time) {
			return slices.Insert(a, i+1, x)
		}
	}

	return slices.Insert(a, 0, x)
}

func (s *Span) AdjustDurationSelf(child *Span) {
	if child.Attrs.Text(xattr.SpanKind) == consumerSpanKind { // async span
		return
	}
	if child.Time.After(s.EndTime()) {
		return
	}

	if child.EndTime().Before(s.EndTime()) {
		s.subtractDurationSelf(child.Duration)
		return
	}

	endTime := minTime(s.EndTime(), child.EndTime())
	s.subtractDurationSelf(endTime.Sub(child.Time))
}

func minTime(a, b time.Time) time.Time {
	if b.Before(a) {
		return b
	}
	return a
}

func (s *Span) subtractDurationSelf(dur time.Duration) {
	if s.DurationSelf <= dur {
		s.DurationSelf = 0
	} else {
		s.DurationSelf -= dur
	}
}

func BuildSpanTree(spansPtr *[]*Span) *Span {
	spans := *spansPtr
	m := make(map[uint64]*Span, len(spans))
	var root *Span

	for _, s := range spans {
		if s.ParentID == 0 {
			root = s
		}

		s.DurationSelf = s.Duration
		m[s.ID] = s
	}

	if root == nil {
		root = newFakeRoot(spans)
	}

	if len(m) != len(spans) {
		spans = spans[:0]
		for _, s := range m {
			spans = append(spans, s)
		}

		*spansPtr = spans
	}

	for _, s := range spans {
		if s.ParentID == 0 {
			continue
		}

		parent := m[s.ParentID]
		if parent == nil {
			parent = root
		}

		parent.AddChild(s)
		parent.AdjustDurationSelf(s)
	}

	return root
}

func newFakeRoot(spans []*Span) *Span {
	sample := spans[0]
	minTime := time.Unix(0, math.MaxInt64)

	for _, s := range spans {
		if s.Time.Before(minTime) {
			minTime = s.Time
		}
	}

	span := new(Span)
	span.ID = rand.Uint64()
	span.TraceID = sample.TraceID
	span.Attrs = AttrMap{
		xattr.SpanTime:       minTime,
		xattr.SpanStatusCode: okStatusCode,
	}
	return span
}

//------------------------------------------------------------------------------

func marshalSpan(span *Span) []byte {
	b, err := msgpack.Marshal(span)
	if err != nil {
		panic(err)
	}
	return b
}

func unmarshalSpan(b []byte, span *Span) error {
	return msgpack.Unmarshal(b, span)
}
