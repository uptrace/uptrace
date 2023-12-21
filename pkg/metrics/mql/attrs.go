package mql

import (
	"cmp"

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

func (attrs Attrs) Pick(grouping map[string]struct{}) Attrs {
	dest := make(Attrs, 0, len(grouping))
	for _, kv := range attrs {
		if _, ok := grouping[kv.Key]; ok {
			dest = append(dest, kv)
		}
	}
	return dest
}

func (attrs Attrs) Bytes(buf []byte, pick map[string]struct{}) []byte {
	const sep = '0'

	if buf == nil {
		buf = make([]byte, 0, len(attrs)*20)
	}

	for _, kv := range attrs {
		if pick != nil {
			if _, ok := pick[kv.Key]; !ok {
				continue
			}
		}
		buf = append(buf, kv.Key...)
		buf = append(buf, sep)
		buf = append(buf, kv.Value...)
		buf = append(buf, sep)
	}
	return buf
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
	slices.SortFunc(attrs, func(a, b KeyValue) int {
		return cmp.Compare(a.Key, b.Key)
	})
}
