package tql

import (
	"fmt"
	"strings"

	"github.com/segmentio/encoding/json"
)

func ParseQueryError(query string) ([]*QueryPart, error) {
	parts := ParseQuery(query)
	for _, part := range parts {
		if part.Error.Wrapped != nil {
			return nil, part.Error.Wrapped
		}
	}
	return parts, nil
}

func ParseQuery(s string) []*QueryPart {
	ss := splitQuery(s)
	parts := make([]*QueryPart, len(ss))

	for i, s := range ss {
		part := &QueryPart{Query: s}
		parts[i] = part

		v, err := ParsePart(s)
		if err != nil {
			part.Error.Wrapped = err
			continue
		}

		part.AST = v
	}

	return parts
}

func ParseColumn(s string) (*Column, error) {
	v, err := ParsePart(s)
	if err != nil {
		return nil, err
	}

	sel, ok := v.(*Selector)
	if !ok {
		return nil, fmt.Errorf("tql: expected *Selector, got %T", v)
	}

	if len(sel.Columns) != 1 {
		return nil, fmt.Errorf("tql: expected 1 column, got %d", len(sel.Columns))
	}
	return &sel.Columns[0], nil
}

type QueryPart struct {
	Query    string    `json:"query"`
	Error    JSONError `json:"error,omitempty"`
	Disabled bool      `json:"disabled,omitempty"`

	AST any `json:"-"`
}

func splitQuery(s string) []string {
	ss := strings.Split(s, " | ")
	for i := len(ss) - 1; i >= 0; i-- {
		s := strings.TrimSpace(ss[i])
		if s == "" {
			ss = append(ss[:i], ss[i+1:]...)
		} else {
			ss[i] = s
		}
	}
	return ss
}

type JSONError struct {
	Wrapped error
}

func (e JSONError) MarshalJSON() ([]byte, error) {
	if e.Wrapped == nil {
		return []byte(`""`), nil
	}
	return json.Marshal(e.Wrapped.Error())
}
