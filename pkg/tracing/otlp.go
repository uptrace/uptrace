package tracing

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/otlpconv"
	"github.com/uptrace/uptrace/pkg/tracing/attrkey"
	"github.com/uptrace/uptrace/pkg/uuid"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/zap"
)

func initSpanFromOTLP(dest *Span, resource AttrMap, src *tracepb.Span) {
	dest.ID = otlpSpanID(src.SpanId)
	dest.ParentID = otlpSpanID(src.ParentSpanId)
	dest.TraceID = otlpTraceID(src.TraceId)
	dest.Name = src.Name
	dest.Kind = otlpSpanKind(src.Kind)

	dest.Time = time.Unix(0, int64(src.StartTimeUnixNano))
	dest.Duration = time.Duration(src.EndTimeUnixNano - src.StartTimeUnixNano)

	if src.Status != nil {
		dest.StatusCode = otlpStatusCode(src.Status.Code)
		dest.StatusMessage = src.Status.Message
	}

	dest.Attrs = make(AttrMap, len(resource)+len(src.Attributes))
	for k, v := range resource {
		dest.Attrs[k] = v
	}
	otlpconv.ForEachKeyValue(src.Attributes, func(key string, value any) {
		dest.Attrs[key] = value
	})
	if rand.Float64() < 0.5 {
		dest.Attrs[attrkey.DeploymentEnvironment] = "stage"
	} else {
		dest.Attrs[attrkey.DeploymentEnvironment] = "prod"
	}

	dest.Events = make([]*Span, len(src.Events))
	for i, event := range src.Events {
		dest.Events[i] = newSpanFromOTLPEvent(event)
	}

	dest.Links = make([]*SpanLink, len(src.Links))
	for i, link := range src.Links {
		dest.Links[i] = newSpanLink(link)
	}
}

func newSpanFromOTLPEvent(event *tracepb.Span_Event) *Span {
	span := new(Span)
	span.EventName = event.Name
	span.Time = time.Unix(0, int64(event.TimeUnixNano))

	span.Attrs = make(AttrMap, len(event.Attributes))
	otlpconv.ForEachKeyValue(event.Attributes, func(key string, value any) {
		span.Attrs[key] = value
	})

	return span
}

func otlpSpanID(b []byte) uint64 {
	switch len(b) {
	case 0:
		return 0
	case 8:
		return binary.LittleEndian.Uint64(b)
	case 12:
		// continue below
	default:
		otelzap.L().Error("otlpSpanID failed", zap.Int("length", len(b)))
		return 0
	}

	s := base64.RawStdEncoding.EncodeToString(b)
	b, err := hex.DecodeString(s)
	if err != nil {
		otelzap.L().Error("otlpSpanID failed", zap.Error(err))
		return 0
	}

	if len(b) == 8 {
		return binary.LittleEndian.Uint64(b)
	}

	otelzap.L().Error("otlpSpanID failed", zap.Int("length", len(b)))
	return 0
}

func otlpTraceID(b []byte) uuid.UUID {
	switch len(b) {
	case 0:
		return uuid.UUID{}
	case 16:
		u, err := uuid.FromBytes(b)
		if err != nil {
			otelzap.L().Error("otlpTraceID failed", zap.Error(err))
		}
		return u
	case 24:
		// continue below
	default:
		otelzap.L().Error("otlpTraceID failed", zap.Int("length", len(b)))
		return uuid.UUID{}
	}

	s := base64.RawStdEncoding.EncodeToString(b)
	b, err := hex.DecodeString(s)
	if err != nil {
		otelzap.L().Error("otlpTraceID failed", zap.Error(err))
		return uuid.UUID{}
	}

	u, err := uuid.FromBytes(b)
	if err != nil {
		otelzap.L().Error("otlpTraceID failed", zap.Error(err))
	}
	return u
}

func otlpSpanKind(kind tracepb.Span_SpanKind) string {
	switch kind {
	case tracepb.Span_SPAN_KIND_SERVER:
		return ServerSpanKind
	case tracepb.Span_SPAN_KIND_CLIENT:
		return ClientSpanKind
	case tracepb.Span_SPAN_KIND_PRODUCER:
		return ProducerSpanKind
	case tracepb.Span_SPAN_KIND_CONSUMER:
		return ConsumerSpanKind
	}
	return InternalSpanKind
}

func otlpStatusCode(code tracepb.Status_StatusCode) string {
	switch code {
	case tracepb.Status_STATUS_CODE_ERROR:
		return ErrorStatusCode
	default:
		return OKStatusCode
	}
}
