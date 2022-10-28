package tracing

import (
	"context"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/tracing/attrkey"
	"github.com/uptrace/uptrace/pkg/uuid"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

const (
	InternalSpanKind = "internal"
	ServerSpanKind   = "server"
	ClientSpanKind   = "client"
	ProducerSpanKind = "producer"
	ConsumerSpanKind = "consumer"
)

const (
	OKStatusCode    = "ok"
	ErrorStatusCode = "error"
)

type Span struct {
	ProjectID uint32 `json:"projectId" ch:"project_id"`
	System    string `json:"system" ch:"system,lc"`
	GroupID   uint64 `json:"groupId,string" ch:"group_id"`

	TraceID  uuid.UUID `json:"traceId" ch:"trace_id,type:UUID"`
	ID       uint64    `json:"id,string" ch:"id"`
	ParentID uint64    `json:"parentId,string,omitempty" ch:"parent_id"`

	Name       string `json:"name" ch:"name,lc"`
	EventName  string `json:"eventName,omitempty" ch:"event_name,lc"`
	Kind       string `json:"kind" ch:"kind,lc"`
	Standalone bool   `json:"standalone,omitempty" ch:"-"`

	Time         time.Time     `json:"time" ch:"time"`
	Duration     time.Duration `json:"duration" ch:"duration"`
	DurationSelf time.Duration `json:"durationSelf" msgpack:"-" ch:"-"`
	StartPct     float64       `json:"startPct" msgpack:"-" ch:"-"`

	StatusCode    string `json:"statusCode" ch:"status_code,lc"`
	StatusMessage string `json:"statusMessage" ch:"status_message"`

	Attrs  AttrMap     `json:"attrs" ch:"-"`
	Events []*Span     `json:"events,omitempty" msgpack:"-" ch:"-"`
	Links  []*SpanLink `json:"links,omitempty" ch:"-"`

	Children []*Span `json:"children,omitempty" msgpack:"-" ch:"-"`

	logMessageHash uint64
}

type SpanLink struct {
	TraceID uuid.UUID `json:"traceId"`
	SpanID  uint64    `json:"spanId"`
	Attrs   AttrMap   `json:"attrs"`
}

func (s *Span) IsEvent() bool {
	return isEventSystem(s.System)
}

func (s *Span) IsError() bool {
	return isErrorSystem(s.System)
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

func (s *Span) AddEvent(event *Span) {
	s.Events = addSpanSorted(s.Events, event)
}

func (s *Span) AdjustDurationSelf(child *Span) {
	if child.Attrs.Text(attrkey.SpanKind) == ConsumerSpanKind { // async span
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

//------------------------------------------------------------------------------

func buildSpanTree(ctx context.Context, app *bunapp.App, spansPtr *[]*Span) *Span {
	spans := *spansPtr
	m := make(map[uint64]*Span, len(spans))
	var root *Span

	for _, s := range spans {
		if s.IsEvent() {
			continue
		}

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
		for _, s := range spans {
			if !s.IsEvent() {
				continue
			}
			if span, ok := m[s.ParentID]; ok {
				span.AddEvent(s)
			} else {
				root.AddEvent(s)
			}
		}

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
			app.Zap(ctx).Error("can't find parent span", zap.Uint64("parent_id", s.ParentID))
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
		attrkey.SpanTime:       minTime,
		attrkey.SpanStatusCode: OKStatusCode,
	}
	return span
}

//------------------------------------------------------------------------------

func isEventSystem(s string) bool {
	if idx := strings.IndexByte(s, ':'); idx >= 0 {
		s = s[:idx]
	}
	switch s {
	case SystemEventOther,
		SystemEventLog,
		SystemEventExceptions,
		SystemEventMessage:
		return true
	default:
		return false
	}
}

func isLogSystem(s string) bool {
	return strings.HasPrefix(s, "log:")
}

func isErrorSystem(s string) bool {
	switch s {
	case SystemEventExceptions, "log:error", "log:fatal", "log:panic":
		return true
	default:
		return false
	}
}

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
