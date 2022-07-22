package tracing

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"reflect"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/tracing/otlpconv"
	"github.com/uptrace/uptrace/pkg/tracing/xotel"
	"github.com/uptrace/uptrace/pkg/uuid"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/zap"
)

func initSpanFromOTLP(dest *Span, resource xotel.AttrMap, src *tracepb.Span) {
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

	dest.Attrs = make(xotel.AttrMap, len(resource)+len(src.Attributes))
	for k, v := range resource {
		dest.Attrs[k] = v
	}
	otlpconv.ForEachAttr(src.Attributes, func(key string, value any) {
		dest.Attrs[key] = value
	})

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

	span.Attrs = make(xotel.AttrMap, len(event.Attributes))
	otlpconv.ForEachAttr(event.Attributes, func(key string, value any) {
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

//------------------------------------------------------------------------------

func toOTLPSpanID(n uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	return b
}

func toOTLPStatusCode(s string) tracepb.Status_StatusCode {
	switch s {
	case "ok":
		return tracepb.Status_STATUS_CODE_OK
	case "error":
		return tracepb.Status_STATUS_CODE_ERROR
	default:
		return tracepb.Status_STATUS_CODE_UNSET
	}
}

func toOTLPSpanKind(s string) tracepb.Span_SpanKind {
	switch s {
	case InternalSpanKind:
		return tracepb.Span_SPAN_KIND_INTERNAL
	case ServerSpanKind:
		return tracepb.Span_SPAN_KIND_SERVER
	case ClientSpanKind:
		return tracepb.Span_SPAN_KIND_CLIENT
	case ProducerSpanKind:
		return tracepb.Span_SPAN_KIND_PRODUCER
	case ConsumerSpanKind:
		return tracepb.Span_SPAN_KIND_CONSUMER
	default:
		return tracepb.Span_SPAN_KIND_UNSPECIFIED
	}
}

func toOTLPAttributes(m xotel.AttrMap) []*commonpb.KeyValue {
	kvs := make([]*commonpb.KeyValue, 0, len(m))
	for k, v := range m {
		if av := toOTLPAnyValue(v); av != nil {
			kvs = append(kvs, &commonpb.KeyValue{
				Key:   k,
				Value: av,
			})
		}
	}
	return kvs
}

func toOTLPAnyValue(v any) *commonpb.AnyValue {
	switch v := v.(type) {
	case string:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: v,
			},
		}
	case int64:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_IntValue{
				IntValue: v,
			},
		}
	case uint64:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_IntValue{
				IntValue: int64(v),
			},
		}
	case float64:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_DoubleValue{
				DoubleValue: v,
			},
		}
	case bool:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_BoolValue{
				BoolValue: v,
			},
		}
	case []byte:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_BytesValue{
				BytesValue: v,
			},
		}
	case []string:
		values := make([]*commonpb.AnyValue, len(v))
		for i, el := range v {
			values[i] = toOTLPAnyValue(el)
		}
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_ArrayValue{
				ArrayValue: &commonpb.ArrayValue{
					Values: values,
				},
			},
		}
	case []int64:
		values := make([]*commonpb.AnyValue, len(v))
		for i, el := range v {
			values[i] = toOTLPAnyValue(el)
		}
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_ArrayValue{
				ArrayValue: &commonpb.ArrayValue{
					Values: values,
				},
			},
		}
	case []float64:
		values := make([]*commonpb.AnyValue, len(v))
		for i, el := range v {
			values[i] = toOTLPAnyValue(el)
		}
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_ArrayValue{
				ArrayValue: &commonpb.ArrayValue{
					Values: values,
				},
			},
		}
	default:
		otelzap.L().Error("unsupported attribute type",
			zap.String("type", reflect.TypeOf(v).String()))
		return nil
	}
}
