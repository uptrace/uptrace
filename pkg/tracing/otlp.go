package tracing

import (
	"encoding/binary"
	"log"
	"strconv"

	"github.com/google/uuid"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

func otlpSpanID(b []byte) uint64 {
	if len(b) != 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(b)
}

func otlpTraceID(b []byte) uuid.UUID {
	u, _ := uuid.FromBytes(b)
	return u
}

const (
	internalSpanKind = "internal"
	serverSpanKind   = "server"
	clientSpanKind   = "client"
	producerSpanKind = "producer"
	consumerSpanKind = "consumer"
)

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

const (
	okStatusCode    = "ok"
	errorStatusCode = "error"
)

func otlpStatusCode(code tracepb.Status_StatusCode) string {
	switch code {
	case tracepb.Status_STATUS_CODE_ERROR:
		return errorStatusCode
	default:
		return okStatusCode
	}
}

func otlpAttrs(kvs []*commonpb.KeyValue) AttrMap {
	dest := make(AttrMap, len(kvs))
	otlpSetAttrs(dest, kvs)
	return dest
}

func otlpSetAttrs(dest AttrMap, kvs []*commonpb.KeyValue) {
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
