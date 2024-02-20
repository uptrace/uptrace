package anyconv

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/segmentio/encoding/json"
)

func Time(v any) time.Time {
	switch v := v.(type) {
	case time.Time:
		return v
	case int64:
		return time.Unix(0, v)
	case uint64:
		return time.Unix(0, int64(v))
	case string:
		tm, _ := time.Parse(time.RFC3339Nano, v)
		return tm
	case json.Number:
		n, _ := v.Int64()
		return time.Unix(0, n)
	default:
		return time.Time{}
	}
}

func Int64(v any) int64 {
	switch v := v.(type) {
	case int64:
		return v
	case json.Number:
		n, _ := v.Int64()
		return n
	case int:
		return int64(v)
	case int32:
		return int64(v)
	default:
		return 0
	}
}

func Uint64(v any) uint64 {
	switch v := v.(type) {
	case int:
		return uint64(v)
	case uint:
		return uint64(v)
	case int64:
		return uint64(v)
	case uint64:
		return v
	case int8:
		return uint64(v)
	case uint8:
		return uint64(v)
	case int16:
		return uint64(v)
	case uint16:
		return uint64(v)
	case int32:
		return uint64(v)
	case uint32:
		return uint64(v)
	case float64:
		return uint64(v)
	case float32:
		return uint64(v)
	case string:
		if len(v) == 16 {
			if b, err := hex.DecodeString(v); err == nil {
				return binary.BigEndian.Uint64(b)
			}
		}
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			return n
		}
		return 0
	case json.Number:
		n, _ := v.Int64()
		return uint64(n)
	default:
		return 0
	}
}
