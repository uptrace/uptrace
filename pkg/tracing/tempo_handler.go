package tracing

import (
	"fmt"
	"net/http"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/uuid"
)

var jsonMarshaler = &jsonpb.Marshaler{}

type TempoHandler struct {
	*bunapp.App
}

func NewTempoHandler(app *bunapp.App) *TempoHandler {
	return &TempoHandler{
		App: app,
	}
}

func (h *TempoHandler) Ready(w http.ResponseWriter, req bunrouter.Request) error {
	_, err := w.Write([]byte("ready\n"))
	return err
}

func (h *TempoHandler) Echo(w http.ResponseWriter, req bunrouter.Request) error {
	_, err := w.Write([]byte("echo\n"))
	return err
}

func (h *TempoHandler) QueryTrace(w http.ResponseWriter, req bunrouter.Request) error {
	contentType := req.Header.Get("Accept")
	if contentType == "" {
		contentType = protobufContentType
	}
	return h.queryTrace(w, req, contentType)
}

func (h *TempoHandler) QueryTraceJSON(w http.ResponseWriter, req bunrouter.Request) error {
	return h.queryTrace(w, req, jsonContentType)
}

func (h *TempoHandler) queryTrace(
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

	resp := newTempopbTrace(spans)
	w.Header().Set("Content-Type", contentType)

	switch contentType {
	case "*/*", jsonContentType:
		return jsonMarshaler.Marshal(w, resp)
	case protobufContentType:
		b, err := proto.Marshal(resp)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		return err
	default:
		return fmt.Errorf("unknown content type: %q", contentType)
	}
}
