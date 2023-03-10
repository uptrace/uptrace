package tracing

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

type SentryHandler struct {
	*bunapp.App

	sp *SpanProcessor
}

func NewSentryHandler(app *bunapp.App, sp *SpanProcessor) *SentryHandler {
	return &SentryHandler{
		App: app,
		sp:  sp,
	}
}

func (h *SentryHandler) Store(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	b, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	// TODO: copy Event from https://pkg.go.dev/github.com/getsentry/sentry-go#Event
	// to this package and use it here to unmarshal the JSON.

	data := make(map[string]any)

	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	if false {
		span := new(Span)
		// TODO: map data into span
		h.sp.AddSpan(ctx, span)
	}

	return nil
}
