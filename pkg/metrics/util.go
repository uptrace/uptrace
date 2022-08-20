package metrics

import (
	"github.com/uptrace/uptrace/pkg/bunlex"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
	"golang.org/x/exp/constraints"
)

func listToSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func min[T constraints.Ordered](a, b T) T {
	if a <= b {
		return a
	}
	return b
}

func max[T constraints.Ordered](a, b T) T {
	if a >= b {
		return a
	}
	return b
}

//------------------------------------------------------------------------------

func cleanAttrKey(s string) string {
	return cleanMetricName(s)
}

func cleanMetricName(s string) string {
	if isValidMetricName(s) {
		return s
	}

	r := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if isAllowedMetricNameChar(c) {
			r = append(r, c)
		} else {
			r = append(r, '_')
		}
	}
	return unsafeconv.String(r)
}

func isValidMetricName(s string) bool {
	for _, c := range []byte(s) {
		if !isAllowedMetricNameChar(c) {
			return false
		}
	}
	return true
}

func isAllowedMetricNameChar(c byte) bool {
	return bunlex.IsAlnum(c) || c == '_' || c == '.'
}
