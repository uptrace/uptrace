package uql

import (
	"fmt"
	"strings"
)

type Part struct {
	Query    string `json:"query"`
	Disabled bool   `json:"disabled,omitempty"`
	Error    string `json:"error,omitempty"`

	AST any `json:"-"`
}

func (p *Part) SetError(s string) {
	if p.Error == "" {
		p.Error = s
	}
}

func Parse(s string) []*Part {
	ss := splitQuery(s)
	parts := make([]*Part, len(ss))

	for i, s := range ss {
		part := &Part{Query: s}
		parts[i] = part

		v, err := ParsePart(s)
		if err != nil {
			part.Error = err.Error()
			continue
		}

		part.AST = v
	}

	return parts
}

func ParseName(s string) (Name, error) {
	v, err := ParsePart(s)
	if err != nil {
		return Name{}, nil
	}

	cols, ok := v.(*Columns)
	if !ok {
		return Name{}, fmt.Errorf("uql: got %T, wanted *Columns", v)
	}

	if len(cols.Names) != 1 {
		return Name{}, fmt.Errorf("uql: got %d names, wanted 1", len(cols.Names))
	}
	return cols.Names[0], nil
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
