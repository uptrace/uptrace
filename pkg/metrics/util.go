package metrics

import (
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
