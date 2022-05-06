package tracing

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type TempoHandler struct {
	*bunapp.App
}

func NewTempoHandler(app *bunapp.App) *TempoHandler {
	return &TempoHandler{
		App: app,
	}
}

func (h *TempoHandler) ShowTraceProto(w http.ResponseWriter, req bunrouter.Request) error {
	return h.showTrace(w, req, protobufContentType)
}

func (h *TempoHandler) ShowTraceJSON(w http.ResponseWriter, req bunrouter.Request) error {
	return h.showTrace(w, req, jsonContentType)
}

func (h *TempoHandler) showTrace(
	w http.ResponseWriter, req bunrouter.Request, contentType string,
) error {
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

	var data []byte

	switch contentType {
	case jsonContentType:
		data, err = protojson.Marshal(resp)
	case protobufContentType:
		data, err = proto.Marshal(resp)
	}

	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}
