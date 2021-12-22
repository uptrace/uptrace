package tracing

import (
	"context"
	"encoding/json"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/tracing/xattr"
	"golang.org/x/exp/slices"
)

type SpanIndex struct {
	ch.BaseModel `ch:"table:spans_index_buffer,alias:s"`

	System  string `ch:"span.system,lc"`
	GroupID uint64 `ch:"span.group_id"`

	ID       uint64    `ch:"span.id"`
	ParentID uint64    `ch:"-"`
	TraceID  uuid.UUID `ch:"span.trace_id,type:UUID"`
	Name     string    `ch:"span.name,lc"`
	Kind     string    `ch:"span.kind,lc"`

	Time     time.Time     `ch:"span.time"`
	Duration time.Duration `ch:"span.duration"`

	StatusCode    string `ch:"span.status_code,lc"`
	StatusMessage string `ch:"span.status_message"`

	Attrs      AttrMap  `ch:"-"`
	AttrKeys   []string `ch:",lc"`
	AttrValues []string `ch:",lc"`

	ServiceName string `ch:"service.name,lc"`
	HostName    string `ch:"host.name,lc"`
}

type SpanData struct {
	ch.BaseModel `ch:"table:spans_data_buffer,alias:s"`

	TraceID  uuid.UUID    `json:"traceId"`
	ID       uint64       `json:"id,string"`
	ParentID uint64       `json:"parentId,string,omitempty"`
	Time     time.Time    `json:"time"`
	Attrs    AttrMap      `json:"attrs"`
	Events   []*SpanEvent `json:"events"`
	Links    []*SpanLink  `json:"links"`
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

type AttrMap map[string]any

func (m AttrMap) Clone() AttrMap {
	clone := make(AttrMap, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

func (m AttrMap) Has(key string) bool {
	_, ok := m[key]
	return ok
}

func (m AttrMap) Text(key string) string {
	s, _ := m[key].(string)
	return s
}

func (m AttrMap) Int64(key string) int64 {
	switch v := m[key].(type) {
	case int64:
		return v
	case json.Number:
		n, _ := v.Int64()
		return n
	default:
		return 0
	}
}

func (m AttrMap) Uint64(key string) uint64 {
	switch v := m[key].(type) {
	case uint64:
		return v
	case json.Number:
		n, _ := strconv.ParseUint(string(v), 10, 64)
		return n
	default:
		return 0
	}
}

func (m AttrMap) Time(key string) time.Time {
	switch v := m[key].(type) {
	case time.Time:
		return v
	case string:
		tm, _ := time.Parse(time.RFC3339Nano, v)
		return tm
	default:
		return time.Time{}
	}
}

func (m AttrMap) Duration(key string) time.Duration {
	switch v := m[key].(type) {
	case time.Duration:
		return v
	case json.Number:
		n, _ := strconv.ParseInt(string(v), 10, 64)
		return time.Duration(n)
	case string:
		dur, _ := time.ParseDuration(v)
		return dur
	default:
		return 0
	}
}

func (m AttrMap) ServiceName() string {
	return m.Text(xattr.ServiceName)
}

func SelectSpan(ctx context.Context, app *bunapp.App, span *Span) error {
	if err := app.CH().NewSelect().
		Model(span).
		Column("parent_id", "attrs", "events", "links").
		Where("trace_id = ?", span.TraceID).
		Where("id = ?", span.ID).
		Limit(1).
		Scan(ctx); err != nil {
		return err
	}
	return nil
}

//------------------------------------------------------------------------------

type Span struct {
	SpanData `ch:",inherit"`

	System  string `json:"system"`
	GroupID uint64 `json:"groupId,string"`

	Name string `json:"name"`
	Kind string `json:"kind"`

	StatusCode    string `json:"statusCode"`
	StatusMessage string `json:"statusMessage"`

	Duration     time.Duration `json:"duration"`
	DurationSelf time.Duration `json:"durationSelf"`
	StartPct     float64       `json:"startPct"`

	parent   *Span
	Children []*Span `json:"children,omitempty"`
}

var _ ch.AfterScanRowHook = (*Span)(nil)

func (s *Span) AfterScanRow(ctx context.Context) error {
	s.System = s.Attrs.Text(xattr.SpanSystem)
	s.GroupID = s.Attrs.Uint64(xattr.SpanGroupID)

	s.Name = s.Attrs.Text(xattr.SpanName)
	s.Kind = s.Attrs.Text(xattr.SpanKind)

	s.StatusCode = s.Attrs.Text(xattr.SpanStatusCode)
	s.StatusMessage = s.Attrs.Text(xattr.SpanStatusMessage)

	s.Duration = s.Attrs.Duration(xattr.SpanDuration)
	s.Time = s.Attrs.Time(xattr.SpanTime)

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

		s.parent = parent
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
