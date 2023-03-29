package tracing

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/segmentio/encoding/json"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/anyconv"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const vectorSDK = "vector"

type VectorHandler struct {
	*bunapp.App

	sp *SpanProcessor
}

func NewVectorHandler(app *bunapp.App, sp *SpanProcessor) *VectorHandler {
	return &VectorHandler{
		App: app,
		sp:  sp,
	}
}

func (h *VectorHandler) Create(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn := dsnFromRequest(req)
	if dsn == "" {
		return errors.New("uptrace-dsn header is empty or missing")
	}

	project, err := org.SelectProjectByDSN(ctx, h.App, dsn)
	if err != nil {
		return err
	}

	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.SetAttributes(attribute.Int64("project", int64(project.ID)))
	}

	switch ct := req.Header.Get("Content-Type"); ct {
	case "application/x-ndjson", "application/json":
		// ok
	default:
		// TODO: update error and improve check
		return fmt.Errorf(
			`got %q, wanted %q (use encoding.codec = "json" and framing.method = "newline_delimited")`,
			ct, "application/json")
	}

	dec := json.NewDecoder(http.MaxBytesReader(w, req.Body, 10<<20))
	dec.DisallowUnknownFields()
	dec.DontMatchCaseInsensitiveStructFields()

	for {
		var m map[string]any

		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		span := new(Span)
		h.spanFromVector(ctx, span, m)
		span.ProjectID = project.ID
		h.sp.AddSpan(ctx, span)
	}

	return nil
}

func (h *VectorHandler) spanFromVector(ctx context.Context, span *Span, params AttrMap) {
	// Can be overridden later with the information parsed from the log message.
	span.ID = rand.Uint64()

	span.Kind = InternalSpanKind
	span.EventName = otelEventLog
	span.StatusCode = OKStatusCode

	span.Attrs = make(AttrMap, len(params)+2)
	span.Attrs[attrkey.TelemetrySDKName] = vectorSDK

	if msg, _ := params[attrkey.LogMessage].(string); msg != "" {
		span.Attrs[attrkey.LogMessage] = msg
	} else if msg := popLogMessageParam(params); msg != "" {
		span.Attrs[attrkey.LogMessage] = msg
	}

	span.Time = time.Now()
	for _, key := range []string{"timestamp", "datetime", "time", "date"} {
		value, _ := params[key].(string)
		if value == "" {
			continue
		}

		tm := anyconv.Time(value)
		if tm.IsZero() {
			continue
		}

		span.Time = tm
		delete(params, key)
		break
	}

	populateSpanFromParams(span, params)
}
