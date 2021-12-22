package tracing

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"golang.org/x/exp/utf8string"
)

func scaleWithCPU(min, max int) int {
	if min == 0 {
		panic("min == 0")
	}
	if max == 0 {
		panic("max == 0")
	}

	n := runtime.GOMAXPROCS(0) * min
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func asString(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return fmt.Sprint(v)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return utf8string.NewString(s).Slice(0, n)
}

const jsMaxSafeInt = 1<<53 - 1

func fixJSBigInt(m map[string]any) {
	for k, v := range m {
		switch v := v.(type) {
		case json.Number:
			n, err := v.Int64()
			if err == nil && n > jsMaxSafeInt {
				m[k] = strconv.FormatInt(n, 10)
			}
		case int64:
			if v > jsMaxSafeInt {
				m[k] = strconv.FormatInt(v, 10)
			}
		case uint64:
			if v > jsMaxSafeInt {
				m[k] = strconv.FormatUint(v, 10)
			}
		}
	}
}
