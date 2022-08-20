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
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/attrkey"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
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
	const contentType = "application/x-ndjson"

	ctx := req.Context()

	dsn := req.Header.Get("uptrace-dsn")
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

	if ct := req.Header.Get("Content-Type"); ct != contentType {
		return fmt.Errorf(`got %q, wanted %q (use encoding.codec = "ndjson")`,
			ct, contentType)
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

func (h *VectorHandler) spanFromVector(ctx context.Context, span *Span, vector AttrMap) {
	// Can be overridden later with the information parsed from the log message.
	span.ID = rand.Uint64()

	span.Kind = InternalSpanKind
	span.EventName = LogEventType
	span.StatusCode = OKStatusCode

	attrs := make(AttrMap, len(vector)+2)
	span.Attrs = attrs
	attrs[attrkey.TelemetrySDKName] = vectorSDK

	for key, value := range vector {
		switch key {
		case "level", "severity":
			found, _ := value.(string)
			if found == "" {
				break
			}

			if normalized := normalizeLogSeverity(found); normalized != "" {
				attrs.SetDefault(attrkey.LogSeverity, normalized)
			}
		case "message":
			if s, _ := value.(string); s != "" {
				attrs.SetDefault(attrkey.LogMessage, s)
			}
		case "timestamp":
			if s, _ := value.(string); s != "" {
				tm, err := time.Parse(time.RFC3339Nano, s)
				if err != nil {
					h.Zap(ctx).Error("time.Parse failed", zap.Error(err))
					span.Time = time.Now()
				} else {
					span.Time = tm
				}
			}
		case "file":
			if s, _ := value.(string); s != "" {
				attrs.SetDefault(attrkey.LogFilepath, s)
			}
		case "host", "hostname":
			if s, _ := value.(string); s != "" {
				attrs.SetDefault(attrkey.HostName, s)
			}
		case "source_type":
			if s, _ := value.(string); s != "" {
				attrs.SetDefault(attrkey.LogSource, s)
			}
		default:
			// Plain keys have a priority over discovered keys.
			attrs[key] = value
		}
	}
}
