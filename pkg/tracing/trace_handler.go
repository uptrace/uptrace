package tracing

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
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

	traceID, err := uuid.Parse(req.Param("trace_id"))
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
		"trace": map[string]any{
			"id":        span.TraceID,
			"projectId": span.ProjectID,
		},
	})
}

func (h *TraceHandler) ShowTrace(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	traceID, err := uuid.Parse(req.Param("trace_id"))
	if err != nil {
		return err
	}

	spans, err := SelectTraceSpans(ctx, h.App, traceID)
	if err != nil {
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

	return httputil.JSON(w, bunrouter.H{
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

	return httputil.JSON(w, bunrouter.H{
		"span": span,
	})
}

//------------------------------------------------------------------------------

func (h *TraceHandler) ShowTraceJSON(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	traceID, err := uuid.Parse(req.Param("trace_id"))
	if err != nil {
		return err
	}

	spans, err := SelectTraceSpans(ctx, h.App, traceID)
	if err != nil {
		return err
	}

	if len(spans) == 0 {
		return httperror.NotFound("Trace %q not found. Try again later.", traceID)
	}

	tracepbSpans := make([]*tracepb.Span, len(spans))
	for i, span := range spans {
		tracepbSpans[i] = span.TracepbSpan()
	}

	resp := &tracepb.TracesData{
		ResourceSpans: []*tracepb.ResourceSpans{{
			// Here we should have resource attributes that are common for all spans.
			// But in the database, we mix resource attributes and normal attributes
			// together so the information about resource attributes is lost.
			//
			// Using nil here should work too.
			Resource: nil,

			ScopeSpans: []*tracepb.ScopeSpans{{
				Scope: nil,
				Spans: tracepbSpans,
			}},

			// InstrumentationLibrarySpans field is deprecated in favor of ScopeSpans.
			// It will be removed in next versions.
			InstrumentationLibrarySpans: []*tracepb.InstrumentationLibrarySpans{{
				InstrumentationLibrary: nil,
				Spans:                  tracepbSpans,
			}},
		}},
	}
	b, err := protojson.Marshal(resp)
	if err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}
