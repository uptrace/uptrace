package tracing

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
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
	otlpSetAttrs(dest.Attrs, src.Attributes)

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
	otlpSetAttrs(span.Attrs, event.Attributes)

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
		return serverSpanKind
	case tracepb.Span_SPAN_KIND_CLIENT:
		return clientSpanKind
	case tracepb.Span_SPAN_KIND_PRODUCER:
		return producerSpanKind
	case tracepb.Span_SPAN_KIND_CONSUMER:
		return consumerSpanKind
	}
	return internalSpanKind
}

func otlpStatusCode(code tracepb.Status_StatusCode) string {
	switch code {
	case tracepb.Status_STATUS_CODE_ERROR:
		return errorStatusCode
	default:
		return okStatusCode
	}
}

func otlpAttrs(kvs []*commonpb.KeyValue) xotel.AttrMap {
	dest := make(xotel.AttrMap, len(kvs))
	otlpSetAttrs(dest, kvs)
	return dest
}

func otlpSetAttrs(dest xotel.AttrMap, kvs []*commonpb.KeyValue) {
	for _, kv := range kvs {
		if kv == nil || kv.Value == nil {
			continue
		}
		if value, ok := otlpValue(*kv.Value); ok {
			dest[kv.Key] = value
		}
	}
}

func otlpValue(v commonpb.AnyValue) (any, bool) {
	switch v := v.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		return v.StringValue, true
	case *commonpb.AnyValue_IntValue:
		return v.IntValue, true
	case *commonpb.AnyValue_DoubleValue:
		return v.DoubleValue, true
	case *commonpb.AnyValue_BoolValue:
		return v.BoolValue, true
	case *commonpb.AnyValue_ArrayValue:
		return otlpArray(v.ArrayValue.Values)
	case *commonpb.AnyValue_KvlistValue:
		return otlpAttrs(v.KvlistValue.Values), true
	}

	log.Printf("unsupported attribute value %T", v.Value)
	return nil, false
}

func otlpArray(vs []*commonpb.AnyValue) ([]string, bool) {
	if len(vs) == 0 {
		return nil, false
	}

	switch value := vs[0].Value; value.(type) {
	case *commonpb.AnyValue_StringValue:
		ss := make([]string, len(vs))
		for i, v := range vs {
			if v == nil {
				continue
			}
			if v, ok := v.Value.(*commonpb.AnyValue_StringValue); ok {
				ss[i] = v.StringValue
			}
		}
		return ss, true
	case *commonpb.AnyValue_IntValue:
		ss := make([]string, len(vs))
		for i, v := range vs {
			if v == nil {
				continue
			}
			if v, ok := v.Value.(*commonpb.AnyValue_IntValue); ok {
				ss[i] = strconv.FormatInt(v.IntValue, 10)
			}
		}
		return ss, true
	case *commonpb.AnyValue_DoubleValue:
		ss := make([]string, len(vs))
		for i, v := range vs {
			if v == nil {
				continue
			}
			if v, ok := v.Value.(*commonpb.AnyValue_DoubleValue); ok {
				ss[i] = strconv.FormatFloat(v.DoubleValue, 'f', -1, 64)
			}
		}
		return ss, true
	case *commonpb.AnyValue_BoolValue:
		ss := make([]string, len(vs))
		for i, v := range vs {
			if v == nil {
				continue
			}
			if v, ok := v.Value.(*commonpb.AnyValue_BoolValue); ok {
				ss[i] = strconv.FormatBool(v.BoolValue)
			}
		}
		return ss, true
	default:
		log.Printf("unsupported attribute value %T", value)
		return nil, false
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
	case internalSpanKind:
		return tracepb.Span_SPAN_KIND_INTERNAL
	case serverSpanKind:
		return tracepb.Span_SPAN_KIND_SERVER
	case clientSpanKind:
		return tracepb.Span_SPAN_KIND_CLIENT
	case producerSpanKind:
		return tracepb.Span_SPAN_KIND_PRODUCER
	case consumerSpanKind:
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
