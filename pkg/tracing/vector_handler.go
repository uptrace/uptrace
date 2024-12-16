package tracing

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/segmentio/encoding/json"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/org"
)

const vectorSDK = "vector"

type VectorHandlerParams struct {
	fx.In

	Logger   *otelzap.Logger
	PG       *bun.DB
	PS       *org.ProjectGateway
	Consumer *LogConsumer
}

type VectorHandler struct {
	*VectorHandlerParams
}

func NewVectorHandler(p VectorHandlerParams) *VectorHandler {
	return &VectorHandler{&p}
}

func registerVectorHandler(h *VectorHandler, p bunapp.RouterParams) {
	p.Router.WithGroup("/api/v1", func(g *bunrouter.Group) {
		g.POST("/vector-logs", h.Create)
		g.POST("/vector/logs", h.Create)
	})
}

func (h *VectorHandler) Create(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn, err := org.DSNFromRequest(req)
	if err != nil {
		return err
	}

	project, err := h.PS.SelectByDSN(ctx, dsn)
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
		return fmt.Errorf(
			`got content-type %q, wanted %q`+
				` (use encoding.codec = "json" and framing.method = "newline_delimited")`,
			ct, "application/json")
	}

	p := new(vectorLogProcessor)

	dec := json.NewDecoder(req.Body)
	m := make(map[string]any)
	for {
		clear(m)

		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		span := new(Span)
		p.spanFromVector(ctx, span, m)
		span.ProjectID = project.ID
		h.Consumer.AddSpan(ctx, span)
	}

	return nil
}

//------------------------------------------------------------------------------

type vectorLogProcessor struct {
	baseLogProcessor
}

func (p *vectorLogProcessor) spanFromVector(ctx context.Context, span *Span, params AttrMap) {
	span.ID = idgen.RandSpanID()
	span.Kind = InternalSpanKind
	span.EventName = otelEventLog
	span.StatusCode = OKStatusCode

	span.Attrs = make(AttrMap, len(params)+2)
	span.Attrs[attrkey.TelemetrySDKName] = vectorSDK

	msg, ok := params[attrkey.LogMessage]
	if !ok {
		msg = popLogMessageParam(params)
	}

	switch msg := msg.(type) {
	case string:
		if msg == "" {
			break
		}
		if params, ok := bunutil.IsJSON(msg); ok {
			p.parseJSONLogMessage(span, params)
		} else {
			span.Attrs[attrkey.LogMessage] = msg
		}
	case map[string]any:
		populateSpanFromParams(span, msg)
	default:
		span.Attrs[attrkey.LogMessage] = msg
	}

	if spanName, _ := params["span_name"].(string); spanName != "" {
		span.Name = spanName
		span.EventName = ""
		delete(params, "span_name")
		delete(params, "span_event_name")

		if dur, ok := params["span_duration"].(float64); ok {
			span.Duration = time.Duration(dur)
			delete(params, "span_duration")
		}
	}

	if kind, ok := params["span_kind"].(string); ok {
		span.Kind = kind
		delete(params, "span_kind")
	}
	if code, ok := params["span_status_code"].(string); ok {
		span.StatusCode = code
		delete(params, "span_status_code")
	}
	if msg, ok := params["span_status_message"].(string); ok {
		span.StatusMessage = msg
		delete(params, "span_status_message")
	}

	populateSpanFromParams(span, params)
}
