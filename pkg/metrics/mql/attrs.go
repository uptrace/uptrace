package mql

import (
	"strings"

	"github.com/segmentio/encoding/json"

	"github.com/uptrace/uptrace/pkg/unsafeconv"
	"golang.org/x/exp/slices"
)

type Attrs []KeyValue

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (kv KeyValue) AppendString(b []byte) []byte {
	b = append(b, kv.Key...)
	b = append(b, '=')
	b = append(b, kv.Value...)
	return b
}

func NewAttrs(ss ...string) Attrs {
	attrs := make([]KeyValue, 0, len(ss)/2)
	for i := 0; i < len(ss); i += 2 {
		attrs = append(attrs, KeyValue{
			Key:   ss[i],
			Value: ss[i+1],
		})
	}
	return attrs
}

func AttrsFromMap(m map[string]string) Attrs {
	if len(m) == 0 {
		return nil
	}

	attrs := make([]KeyValue, 0, len(m))

	for k, v := range m {
		attrs = append(attrs, KeyValue{
			Key:   k,
			Value: v,
		})
	}

	SortAttrs(attrs)
	return attrs
}

func AttrsFromKeysValues(keys, values []string) Attrs {
	if len(keys) == 0 {
		return nil
	}

	attrs := make([]KeyValue, 0, len(keys))

	for i, key := range keys {
		attrs = append(attrs, KeyValue{
			Key:   key,
			Value: values[i],
		})
	}

	SortAttrs(attrs)
	return attrs
}

func (attrs Attrs) String() string {
	b := make([]byte, 0, len(attrs)*30)
	b = attrs.AppendString(b, ", ")
	return unsafeconv.String(b)
}

func (attrs Attrs) AppendString(b []byte, sep string) []byte {
	for i, kv := range attrs {
		if i > 0 {
			b = append(b, sep...)
		}
		b = kv.AppendString(b)
	}
	return b
}

func (attrs Attrs) SortedKeys() []string {
	keys := make([]string, 0, len(attrs))
	for _, kv := range attrs {
		keys = append(keys, kv.Key)
	}
	return keys
}

func (attrs Attrs) Pick(keys ...string) Attrs {
	clone := make(Attrs, 0, len(keys))

	i, j := 0, 0
	for i < len(attrs) && j < len(keys) {
		if keys[j] < attrs[i].Key {
			j++
		} else if attrs[i].Key < keys[j] {
			i++
		} else {
			clone = append(clone, attrs[i])
			i++
			j++
		}
	}

	return clone
}

func (attrs Attrs) Bytes(buf []byte) []byte {
	if buf == nil {
		buf = make([]byte, 0, len(attrs)*20)
	}

	for _, kv := range attrs {
		buf = append(buf, kv.Key...)
		buf = append(buf, sep)
		buf = append(buf, kv.Value...)
		buf = append(buf, sep)
	}
	return buf
}

func (attrs Attrs) BytesWithKeys(buf []byte, keys ...string) []byte {
	if len(keys) == 0 {
		return buf
	}
	if buf == nil {
		buf = make([]byte, 0, len(keys)*20)
	}

	i, j := 0, 0
	for i < len(attrs) && j < len(keys) {
		kv := attrs[i]
		if keys[j] < kv.Key {
			j++
		} else if kv.Key < keys[j] {
			i++
		} else {
			buf = append(buf, kv.Key...)
			buf = append(buf, sep)
			buf = append(buf, kv.Value...)
			buf = append(buf, sep)
			i++
			j++
		}
	}
	return buf
}

func (attrs Attrs) Intersect(other Attrs) Attrs {
	set := make(Attrs, 0, min(len(attrs), len(other)))
	for _, kv := range attrs {
		if _, ok := slices.BinarySearchFunc(other, kv, func(a, b KeyValue) int {
			return strings.Compare(a.Key, b.Key)
		}); ok {
			set = append(set, kv)
		}
	}
	return set
}

func (attrs Attrs) Map() map[string]string {
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		m[kv.Key] = kv.Value
	}
	return m
}

func (attrs Attrs) MarshalJSON() ([]byte, error) {
	return json.Marshal(attrs.Map())
}

func (attrs *Attrs) UnmarshalJSON(b []byte) error {
	m := make(map[string]string)

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	kvs := make(Attrs, 0, len(m))

	for k, v := range m {
		kvs = append(kvs, KeyValue{
			Key:   k,
			Value: v,
		})
	}

	SortAttrs(kvs)
	*attrs = kvs

	return nil
}

func SortAttrs(attrs Attrs) {
	slices.SortFunc(attrs, func(a, b KeyValue) bool {
		return strings.Compare(a.Key, b.Key) == -1
	})
}
