package tracing

import (
	"cmp"
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/org"
	"golang.org/x/exp/slices"
)

type TraceHandler struct {
	logger *otelzap.Logger
	ch     *ch.DB
}

func NewTraceHandler(logger *otelzap.Logger, ch *ch.DB) *TraceHandler {
	return &TraceHandler{
		logger: logger,
		ch:     ch,
	}
}

func (h *TraceHandler) FindTrace(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	traceID, err := idgen.ParseTraceID(req.URL.Query().Get("trace_id"))
	if err != nil {
		return err
	}

	var projectID uint32
	var spanID idgen.SpanID
	if err := h.ch.NewSelect().
		Model((*SpanData)(nil)).
		ColumnExpr("project_id, id").
		Where("trace_id = ?", traceID).
		OrderExpr("id DESC").
		Limit(1).
		Scan(ctx, &projectID, &spanID); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"span": map[string]any{
			"projectId":  projectID,
			"traceId":    traceID,
			"id":         spanID,
			"standalone": spanID == 0,
		},
	})
}

func (h *TraceHandler) ShowTrace(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	traceID, err := idgen.ParseTraceID(req.Param("trace_id"))
	if err != nil {
		return err
	}

	fakeApp := &bunapp.App{CH: h.ch}
	spans, hasMore, err := SelectTraceSpans(ctx, fakeApp, traceID)
	if err != nil {
		return err
	}

	if len(spans) == 0 {
		return httperror.NotFound("Trace %q not found. Try again later.", traceID)
	}

	root, numSpan := buildSpanTree(spans)

	if rootSpanIDStr := req.URL.Query().Get("root_span_id"); rootSpanIDStr != "" {
		rootSpanID, err := idgen.ParseSpanID(rootSpanIDStr)
		if err != nil {
			return err
		}

		_ = root.Walk(func(span, parent *Span) error {
			if span.ID == rootSpanID {
				span.ParentID = 0
				root = span
				return walkBreak
			}
			return nil
		})
	}

	traceInfo := NewTraceInfo(root)
	_ = root.Walk(func(span, parent *Span) error {
		setSpanSelfDuration(span)
		span.StartPct = traceInfo.spanPct(span.Time)
		span.EndPct = traceInfo.spanPct(span.EndTime())

		slices.SortFunc(span.Children, func(a, b *Span) int {
			return cmp.Compare(a.Time.UnixNano(), b.Time.UnixNano())
		})
		slices.SortFunc(span.Events, func(a, b *SpanEvent) int {
			return cmp.Compare(a.Time.UnixNano(), b.Time.UnixNano())
		})

		return nil
	})

	return httputil.JSON(w, bunrouter.H{
		"trace":   traceInfo,
		"root":    root,
		"numSpan": numSpan,
		"hasMore": hasMore,
	})
}

func setSpanSelfDuration(span *Span) {
	span.DurationSelf = span.Duration
	prevEndTime := span.Time
	for _, child := range span.Children {
		updateSpanSelfDuration(span, child, prevEndTime)
		if endTime := child.EndTime(); endTime.After(prevEndTime) {
			prevEndTime = endTime
		}
	}
}

func updateSpanSelfDuration(parent, child *Span, prevEndTime time.Time) {
	spanEndTime := parent.EndTime()
	childEndTime := child.EndTime()

	if child.Time.After(spanEndTime) {
		return
	}

	startTime := maxTime(child.Time, prevEndTime)
	endTime := minTime(childEndTime, spanEndTime)
	if endTime.After(startTime) {
		dur := endTime.Sub(startTime)
		if dur < parent.DurationSelf {
			parent.DurationSelf -= dur
		} else {
			parent.DurationSelf = 0
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

func (h *TraceHandler) ShowSpan(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	traceID, err := idgen.ParseTraceID(req.Param("trace_id"))
	if err != nil {
		return err
	}

	spanID, err := idgen.ParseSpanID(req.Param("span_id"))
	if err != nil {
		return err
	}

	fakeApp := &bunapp.App{CH: h.ch}
	span, err := SelectSpan(ctx, fakeApp, project.ID, traceID, spanID)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"span": span,
	})
}

//------------------------------------------------------------------------------

type TraceInfo struct {
	ID       idgen.TraceID `json:"id"`
	Time     time.Time     `json:"time"`
	Duration time.Duration `json:"duration"`
}

func NewTraceInfo(root *Span) *TraceInfo {
	startTime, endTime := root.TreeStartEndTime()
	return &TraceInfo{
		ID:       root.TraceID,
		Time:     startTime,
		Duration: endTime.Sub(startTime),
	}
}

func (trace *TraceInfo) spanPct(spanTime time.Time) float32 {
	if trace.Duration == 0 {
		return 0
	}

	dur := spanTime.Sub(trace.Time)
	pct := float64(dur) / float64(trace.Duration)
	if pct > 1 {
		pct = 1
	}
	return float32(pct)
}
