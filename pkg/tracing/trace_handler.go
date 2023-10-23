package tracing

import (
	"net/http"
	"strconv"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/uuid"
	"golang.org/x/exp/slices"
)

type TraceHandler struct {
	*bunapp.App
}

func NewTraceHandler(app *bunapp.App) *TraceHandler {
	return &TraceHandler{
		App: app,
	}
}

func (h *TraceHandler) FindTrace(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	traceID, err := uuid.Parse(req.URL.Query().Get("trace_id"))
	if err != nil {
		return err
	}

	span := &Span{
		TraceID: traceID,
	}
	if err := SelectSpan(ctx, h.App, span); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"span": bunrouter.H{
			"projectId":  span.ProjectID,
			"traceId":    span.TraceID,
			"id":         strconv.FormatUint(span.ID, 10),
			"standalone": span.Standalone,
		},
	})
}

func (h *TraceHandler) ShowTrace(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	traceID, err := uuid.Parse(req.Param("trace_id"))
	if err != nil {
		return err
	}

	spans, hasMore, err := SelectTraceSpans(ctx, h.App, traceID)
	if err != nil {
		return err
	}

	if len(spans) == 0 {
		return httperror.NotFound("Trace %q not found. Try again later.", traceID)
	}

	root, numSpan := buildSpanTree(spans)
	traceInfo := NewTraceInfo(root)

	_ = root.Walk(func(span, parent *Span) error {
		span.StartPct = spanPct(traceInfo, span.Time)
		span.EndPct = spanPct(traceInfo, span.EndTime())

		prevEndTime := span.Time
		for _, child := range span.Children {
			span.UpdateDurationSelf(child, prevEndTime)
			if endTime := child.EndTime(); endTime.After(prevEndTime) {
				prevEndTime = endTime
			}
		}

		slices.SortFunc(span.Children, func(a, b *Span) bool { return a.Time.Before(b.Time) })
		slices.SortFunc(span.Events, func(a, b *SpanEvent) bool { return a.Time.Before(b.Time) })

		return nil
	})

	return httputil.JSON(w, bunrouter.H{
		"trace":   traceInfo,
		"root":    root,
		"numSpan": numSpan,
		"hasMore": hasMore,
	})
}

func (h *TraceHandler) ShowSpan(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	traceID, err := uuid.Parse(req.Param("trace_id"))
	if err != nil {
		return err
	}

	spanID, err := req.Params().Uint64("span_id")
	if err != nil {
		return err
	}

	span := new(Span)
	span.ID = spanID
	span.TraceID = traceID

	if err := SelectSpan(ctx, h.App, span); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"span": span,
	})
}

//------------------------------------------------------------------------------

type TraceInfo struct {
	ID       uuid.UUID     `json:"id"`
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

func spanPct(trace *TraceInfo, spanTime time.Time) float32 {
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
