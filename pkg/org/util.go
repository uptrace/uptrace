package org

import "fmt"

type AttrMatcherOp string

const (
	AttrEqual    = "="
	AttrNotEqual = "!="
)

type AttrMatcher struct {
	Attr  string        `json:"attr"`
	Op    AttrMatcherOp `json:"op"`
	Value string        `json:"value"`
}

func (m *AttrMatcher) Matches(attrs map[string]any) bool {
	valueAny, ok := attrs[m.Attr]
	if !ok {
		return false
	}

	value := fmt.Sprint(valueAny)

	switch m.Op {
	case "=":
		return value == m.Value
	case "!=":
		return value != m.Value
	default:
		return false
	}
}
