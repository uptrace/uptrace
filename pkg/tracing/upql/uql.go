package upql

import (
	"fmt"
	"strings"
)

type QueryPart struct {
	Query    string `json:"query"`
	Disabled bool   `json:"disabled"`
	Error    string `json:"error"`

	AST any `json:"-"`
}

func (p *QueryPart) SetError(s string, args ...any) {
	if p.Error == "" {
		p.Error = fmt.Sprintf(s, args...)
	}
}

func Parse(s string) []*QueryPart {
	ss := splitQuery(s)
	parts := make([]*QueryPart, len(ss))

	for i, s := range ss {
		part := &QueryPart{Query: s}
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
		return Name{}, fmt.Errorf("upql: got %T, wanted *Columns", v)
	}

	if len(cols.Names) != 1 {
		return Name{}, fmt.Errorf("upql: got %d names, wanted 1", len(cols.Names))
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
