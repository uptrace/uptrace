package tracing

import (
	"math/rand"
	"strings"
	"time"

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
	ID         uint64    `json:"id,string" msgpack:"-" ch:"id"`
	ParentID   uint64    `json:"parentId,string,omitempty" msgpack:"-"`
	TraceID    uuid.UUID `json:"traceId" msgpack:"-" ch:"type:UUID"`
	Standalone bool      `json:"standalone,omitempty" ch:"-"`

	ProjectID uint32 `json:"projectId" msgpack:"-"`
	Type      string `json:"-" msgpack:"-" ch:",lc"`
	System    string `json:"system" ch:",lc"`
	GroupID   uint64 `json:"groupId,string"`
	Kind      string `json:"kind" ch:",lc"`

	Name        string `json:"name" ch:",lc"`
	EventName   string `json:"eventName,omitempty" ch:",lc"`
	DisplayName string `json:"displayName"`

	Time         time.Time     `json:"time" msgpack:"-"`
	Duration     time.Duration `json:"duration"`
	DurationSelf time.Duration `json:"durationSelf" msgpack:"-" ch:"-"`

	StartPct float32 `json:"startPct" msgpack:"-" ch:"-"`
	EndPct   float32 `json:"endPct" msgpack:"-" ch:"-"`

	StatusCode    string `json:"statusCode" ch:",lc"`
	StatusMessage string `json:"statusMessage"`

	Attrs  AttrMap      `json:"attrs" ch:"-"`
	Events []*SpanEvent `json:"events,omitempty" ch:"-"`
	Links  []*SpanLink  `json:"links,omitempty" ch:"-"`

	Children []*Span `json:"children,omitempty" msgpack:"-" ch:"-"`

	logMessageHash uint64
}

type SpanEvent struct {
	Name  string    `json:"name"`
	Time  time.Time `json:"time"`
	Attrs AttrMap   `json:"attrs"`

	System  string `json:"system,omitempty"`
	GroupID uint64 `json:"groupId,omitempty"`
}

type SpanLink struct {
	TraceID uuid.UUID `json:"traceId"`
	SpanID  uint64    `json:"spanId"`
	Attrs   AttrMap   `json:"attrs"`
}

func (s *Span) EventOrSpanName() string {
	if s.EventName != "" {
		return s.EventName
	}
	if s.Name != "" {
		return s.Name
	}
	return emptyPlaceholder
}

func (s *Span) IsEvent() bool {
	return s.EventName != ""
}

func (s *Span) IsError() bool {
	return isErrorSystem(s.System)
}

func (s *Span) Event() *SpanEvent {
	return &SpanEvent{
		Name:  s.DisplayName,
		Time:  s.Time,
		Attrs: s.Attrs,

		System:  s.System,
		GroupID: s.GroupID,
	}
}

func (s *Span) TreeStartEndTime() (time.Time, time.Time) {
	startTime := s.Time
	endTime := s.EndTime()
	for _, child := range s.Children {
		s, e := child.TreeStartEndTime()
		if s.Before(startTime) {
			startTime = s
		}
		if e.After(endTime) {
			endTime = e
		}
	}
	return startTime, endTime
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

func (s *Span) AddEvent(event *SpanEvent) {
	s.Events = append(s.Events, event)
}

func (s *Span) UpdateDurationSelf(child *Span, prevEndTime time.Time) {
	spanEndTime := s.EndTime()
	childEndTime := child.EndTime()

	if child.Time.After(spanEndTime) {
		return
	}

	startTime := maxTime(child.Time, prevEndTime)
	endTime := minTime(childEndTime, spanEndTime)
	if endTime.After(startTime) {
		dur := endTime.Sub(startTime)
		if dur < s.DurationSelf {
			s.DurationSelf -= dur
		} else {
			s.DurationSelf = 0
		}
	}
}

func maxTime(a, b time.Time) time.Time {
	if b.Before(a) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if b.Before(a) {
		return b
	}
	return a
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
				span.AddEvent(s.Event())
			} else {
				root.AddEvent(s.Event())
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
		Type:      SpanTypeFuncs,
		System:    SpanTypeFuncs + ":" + SystemUnknown,
		Kind:      SpanKindInternal,

		Name:        "unknown",
		DisplayName: "The span is missing. Make sure to configure the upstream service to report to Uptrace, end spans on all conditions, and shut down OpenTelemetry when the app exits.",
		Time:        sample.Time,
		StatusCode:  StatusCodeUnset,
		Attrs:       make(AttrMap),
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
	case "log:error", "log:fatal", "log:panic":
		return true
	default:
		return false
	}
}
