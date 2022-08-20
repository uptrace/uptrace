package upql

import (
	"strings"

	"github.com/segmentio/encoding/json"

	"github.com/uptrace/uptrace/pkg/metrics/upql/ast"
)

type ColumnInfo struct{}

type QueryPart struct {
	Query string `json:"query"`

	AST   any       `json:"-"`
	Error JSONError `json:"error"`
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

func Parse(query string) []*QueryPart {
	parts := make([]*QueryPart, 0)

	for _, query := range strings.Split(query, " | ") {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		part := &QueryPart{
			Query: query,
		}
		parts = append(parts, part)

		v, err := ast.Parse(query)
		if err != nil {
			part.Error.Wrapped = err
		} else {
			part.AST = v
		}
	}

	return parts
}
