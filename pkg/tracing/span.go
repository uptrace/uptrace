package tracing

import (
	"math/rand"
	"strings"
	"time"

	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/uuid"
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
	ProjectID uint32 `json:"projectId" msgpack:"-"`
	Type      string `json:"-" msgpack:"-" ch:",lc"`
	System    string `json:"system" ch:",lc"`
	GroupID   uint64 `json:"groupId,string"`

	TraceID  uuid.UUID `json:"traceId" msgpack:"-" ch:"type:UUID"`
	ID       uint64    `json:"id,string" msgpack:"-" ch:"id"`
	ParentID uint64    `json:"parentId,string,omitempty" msgpack:"-"`

	Name       string `json:"name" ch:",lc"`
	EventName  string `json:"eventName,omitempty" ch:",lc"`
	Kind       string `json:"kind" ch:",lc"`
	Standalone bool   `json:"standalone,omitempty" ch:"-"`

	Time         time.Time     `json:"time"`
	Duration     time.Duration `json:"duration"`
	DurationSelf time.Duration `json:"durationSelf" msgpack:"-" ch:"-"`
	StartPct     float64       `json:"startPct" msgpack:"-" ch:"-"`

	StatusCode    string `json:"statusCode" ch:",lc"`
	StatusMessage string `json:"statusMessage"`

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
	s.Children = append(s.Children, child)
}

func (s *Span) AddEvent(event *Span) {
	s.Events = append(s.Events, event)
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

func buildSpanTree(spans []*Span) (*Span, int) {
	var root *Span
	m := make(map[uint64]*Span, len(spans))

	for _, s := range spans {
		if s.IsEvent() {
			continue
		}

		if s.ParentID == 0 {
			root = s
			continue
		}

		m[s.ID] = s
	}

	if root == nil {
		root = newFakeRoot(spans[0])
	}

	for _, s := range spans {
		if s.IsEvent() {
			if span, ok := m[s.ParentID]; ok {
				span.AddEvent(s)
			} else {
				root.AddEvent(s)
			}
			continue
		}

		if s.ParentID == 0 {
			if s.ID != root.ID {
				s.ParentID = root.ID
				root.AddChild(s)
			}
			continue
		}

		parent := m[s.ParentID]
		if parent == nil {
			parent = root
		}
		parent.AddChild(s)
	}

	return root, len(m) + 1
}

func newFakeRoot(sample *Span) *Span {
	span := &Span{
		ID:      rand.Uint64(),
		TraceID: sample.TraceID,

		ProjectID: sample.ProjectID,
		Type:      "service",
		System:    "service:" + SystemUnknown,
		Kind:      SpanKindInternal,

		Name:       "The span is missing. Make sure to configure the upstream service to report to Uptrace, end spans on all conditions, and shut down OpenTelemetry when the app exits.",
		Time:       sample.Time,
		StatusCode: StatusCodeUnset,
		Attrs:      make(AttrMap),
	}
	return span
}

//------------------------------------------------------------------------------

func isEventSystem(s string) bool {
	if s == SystemEventsAll {
		return true
	}
	if idx := strings.IndexByte(s, ':'); idx >= 0 {
		s = s[:idx]
	}
	switch s {
	case EventTypeOther,
		EventTypeLog,
		EventTypeExceptions,
		EventTypeMessage:
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
	case EventTypeExceptions, "log:error", "log:fatal", "log:panic":
		return true
	default:
		return false
	}
}
