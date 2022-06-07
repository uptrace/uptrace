package tracing

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/utf8string"
)

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

func listToSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func quantileLevel(fn string) float64 {
	n, err := strconv.ParseInt(fn[1:], 10, 64)
	if err != nil {
		panic(err)
	}
	return float64(n) / 100
}

//------------------------------------------------------------------------------

func fillHoles(m map[string]any, gte, lt time.Time, interval time.Duration) {
	if len(m) == 0 {
		return
	}

	times, ok := findTimeColumn(m)
	if !ok {
		return
	}

	for k, v := range m {
		switch v := v.(type) {
		case []uint32:
			m[k] = fillOrdered(v, times, gte, lt, interval)
		case []uint64:
			m[k] = fillOrdered(v, times, gte, lt, interval)
		case []int64:
			m[k] = fillOrdered(v, times, gte, lt, interval)
		case []float32:
			m[k] = fillOrdered(v, times, gte, lt, interval)
		case []float64:
			m[k] = fillOrdered(v, times, gte, lt, interval)
		case []time.Time:
			m[k] = fillTime(v, gte, lt, interval)
		}
	}
}

func findTimeColumn(m map[string]any) ([]time.Time, bool) {
	for _, key := range []string{"time", "item.time", "span.time"} {
		if v, ok := m[key].([]time.Time); ok {
			return v, true
		}
	}
	return nil, false
}

func fillOrdered[T constraints.Ordered](
	nums []T,
	times []time.Time,
	gte, lt time.Time,
	interval time.Duration,
) []T {
	numItem := numItem(gte, lt, interval)
	if len(nums) == numItem {
		return nums
	}

	filled := make([]T, numItem)

	for i, num := range nums {
		index := int(times[i].Sub(gte) / interval)
		if index < 0 || index >= numItem {
			return nums
		}
		filled[index] = num
	}

	return filled
}

func fillTime(
	times []time.Time,
	gte, lt time.Time,
	interval time.Duration,
) []time.Time {
	numItem := numItem(gte, lt, interval)
	if len(times) == numItem {
		return times
	}

	for _, tm := range times {
		if tm.Before(gte) || tm.After(lt) {
			return times
		}
	}

	filled := make([]time.Time, numItem)
	for i := range filled {
		filled[i] = gte.Add(time.Duration(i) * interval)
	}
	return filled
}

func numItem(gte, lt time.Time, interval time.Duration) int {
	return int(lt.Sub(gte) / interval)
}

//------------------------------------------------------------------------------

func formatSQL(query string) string {
	cmd := exec.Command("clickhouse-format")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return ""
	}

	if _, err := stdin.Write([]byte(query)); err != nil {
		stdin.Close()
		return ""
	}
	stdin.Close()

	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	return string(out)
}
