package otlpconv

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
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

		key := attrkey.Clean(kv.Key)
		if key == "" {
			continue
		}

		if value, ok := AnyValue(kv.Value); ok {
			fn(key, value)
		}
	}
}

func AnyValue(v *commonpb.AnyValue) (any, bool) {
	switch v := v.Value.(type) {
	case nil:
		return nil, false
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

	ss := make([]string, len(vs))
	for i, v := range vs {
		if str, ok := StringValue(v); ok {
			ss[i] = str
		}
	}
	return ss, true
}

func StringValue(v *commonpb.AnyValue) (string, bool) {
	switch v := v.Value.(type) {
	case nil:
		return "", false
	case *commonpb.AnyValue_StringValue:
		return v.StringValue, true
	case *commonpb.AnyValue_IntValue:
		return strconv.FormatInt(v.IntValue, 10), true
	case *commonpb.AnyValue_DoubleValue:
		return strconv.FormatFloat(v.DoubleValue, 'f', -1, 64), true
	case *commonpb.AnyValue_BoolValue:
		return strconv.FormatBool(v.BoolValue), true
	case *commonpb.AnyValue_ArrayValue:
		ss, ok := Array(v.ArrayValue.Values)
		if !ok {
			return "", false
		}

		b, err := json.Marshal(ss)
		if err != nil {
			return "", false
		}

		return unsafeconv.String(b), true
	case *commonpb.AnyValue_KvlistValue:
		attrs := Map(v.KvlistValue.Values)

		b, err := json.Marshal(attrs)
		if err != nil {
			return "", false
		}

		return unsafeconv.String(b), true
	default:
		return "", false
	}
}
