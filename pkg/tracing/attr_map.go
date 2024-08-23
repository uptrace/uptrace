package tracing

import (
	"strconv"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/tracing/anyconv"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

type AttrMap map[string]any

func NewAttrMap() AttrMap {
	return make(AttrMap)
}

func (m AttrMap) Clone() AttrMap {
	clone := make(AttrMap, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

func (m AttrMap) Merge(other AttrMap) {
	for k, v := range other {
		m[k] = v
	}
}

func (m AttrMap) Delete(key string) {
	delete(m, key)
}

func (m AttrMap) Exists(key string) bool {
	_, ok := m[key]
	return ok
}

func (m AttrMap) SetDefault(key string, value any) {
	if _, ok := m[key]; !ok {
		m[key] = value
	}
}

func (m AttrMap) SetClashingKeys(key string, value any) {
	if !m.Exists(key) {
		m[key] = value
		return
	}

	for _, suffix := range []string{"1", "2", "3"} {
		key := key + suffix
		if !m.Exists(key) {
			m[key] = value
			return
		}
	}
}

func (m AttrMap) Text(key string) string {
	s, _ := m[key].(string)
	return s
}

func (m AttrMap) Get(key string) (any, bool) {
	val, ok := m[key]
	return val, ok
}

func (m AttrMap) GetString(key string) string {
	s, _ := m[key].(string)
	return s
}

func (m AttrMap) GetAsLCString(key string) string {
	if s, ok := m[key].(string); ok {
		return utf8util.TruncLC(s)
	}
	return ""
}

func (m AttrMap) Int64(key string) int64 {
	return anyconv.Int64(key)
}

func (m AttrMap) Uint64(key string) uint64 {
	return anyconv.Uint64(m[key])
}

func (m AttrMap) Time(key string) time.Time {
	return anyconv.Time(m[key])
}

func (m AttrMap) Duration(key string) time.Duration {
	switch v := m[key].(type) {
	case time.Duration:
		return v
	case json.Number:
		n, _ := strconv.ParseInt(string(v), 10, 64)
		return time.Duration(n)
	case string:
		dur, _ := time.ParseDuration(v)
		return dur
	default:
		return 0
	}
}

func (m AttrMap) ServiceName() string {
	s, _ := m[attrkey.ServiceName].(string)
	return s
}

func (m AttrMap) ServiceNameOrUnknown() string {
	if service := m.ServiceName(); service != "" {
		return service
	}
	return "unknown"
}

func (m AttrMap) HostName() string {
	s, _ := m[attrkey.HostName].(string)
	return s
}

func (m AttrMap) PutString(key, val string) {
	m[key] = val
}
