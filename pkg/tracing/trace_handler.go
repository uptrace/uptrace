package tracing

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
)

type TraceHandler struct {
	*bunapp.App
}

func NewTraceHandler(app *bunapp.App) *TraceHandler {
	return &TraceHandler{
		App: app,
	}
}

func (h *TraceHandler) ShowTrace(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	traceID, err := uuid.Parse(req.Param("trace_id"))
	if err != nil {
		return err
	}

	spans := make([]*Span, 0)

	if err := h.CH().NewSelect().
		Model(&spans).
		Column("id", "trace_id", "parent_id", "attrs", "events", "links").
		Where("trace_id = ?", traceID).
		Limit(10000).
		Scan(ctx); err != nil {
		return err
	}

	if len(spans) == 0 {
		return httperror.NotFound("Trace %q not found. Try again later.", traceID)
	}

	root := BuildSpanTree(&spans)
	traceDur := root.TreeEndTime().Sub(root.Time)

	_ = root.Walk(func(s, parent *Span) error {
		s.StartPct = spanStartPct(s, root.Time, traceDur)
		return nil
	})

	return bunrouter.JSON(w, bunrouter.H{
		"trace": bunrouter.H{
			"id":       traceID,
			"time":     root.Time,
			"duration": traceDur,
		},
		"root": root,
	})
}

func spanStartPct(span *Span, traceTime time.Time, traceDur time.Duration) float64 {
	dur := span.Time.Sub(traceTime)
	pct := float64(dur) / float64(traceDur)
	if pct > 1 {
		pct = 1
	}
	return pct
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

	return bunrouter.JSON(w, bunrouter.H{
		"span": span,
	})
}
