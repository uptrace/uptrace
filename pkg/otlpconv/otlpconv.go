package otlpconv

import (
	"log"
	"strconv"

	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
)

func Map(kvs []*commonpb.KeyValue) map[string]any {
	dest := make(map[string]any, len(kvs))
	ForEachKeyValue(kvs, func(key string, value any) {
		dest[key] = value
	})
	return dest
}

func ForEachKeyValue(kvs []*commonpb.KeyValue, fn func(key string, value any)) {
	for _, kv := range kvs {
		if kv == nil || kv.Value == nil {
			continue
		}
		if value, ok := AnyValue(kv.Value); ok {
			fn(kv.Key, value)
		}
	}
}

func AnyValue(v *commonpb.AnyValue) (any, bool) {
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
		return Array(v.ArrayValue.Values)
	case *commonpb.AnyValue_KvlistValue:
		return Map(v.KvlistValue.Values), true
	}

	log.Printf("unsupported attribute value %T", v.Value)
	return nil, false
}

func Array(vs []*commonpb.AnyValue) ([]string, bool) {
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
