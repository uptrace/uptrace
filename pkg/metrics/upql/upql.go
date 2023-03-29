package upql

import (
	"errors"
	"strings"

	"github.com/segmentio/encoding/json"

	"github.com/uptrace/uptrace/pkg/metrics/upql/ast"
)

type MetricAlias struct {
	Name  string `yaml:"name" json:"name"`
	Alias string `yaml:"alias" json:"alias"`
}

func (m *MetricAlias) String() string {
	return m.Name + " as $" + m.Alias
}

func (m *MetricAlias) Validate() error {
	if m.Name == "" {
		return errors.New("metric name can't be empty")
	}
	if m.Alias == "" {
		return errors.New("metric alias can't be empty")
	}
	return nil
}

type ParsedQuery struct {
	Parts   []*QueryPart  `json:"parts"`
	Columns []*ColumnInfo `json:"columns"`
}

type QueryPart struct {
	Query    string    `json:"query"`
	Error    JSONError `json:"error,omitempty"`
	Disabled bool      `json:"disabled,omitempty"`

	AST any `json:"-"`
}

type ColumnInfo struct{}

type JSONError struct {
	Wrapped error
}

func (e JSONError) MarshalJSON() ([]byte, error) {
	if e.Wrapped == nil {
		return []byte(`""`), nil
	}
	return json.Marshal(e.Wrapped.Error())
}

func ParseError(query string) (*ParsedQuery, error) {
	parsedQuery := Parse(query)
	for _, part := range parsedQuery.Parts {
		if part.Error.Wrapped != nil {
			return nil, part.Error.Wrapped
		}
	}
	return parsedQuery, nil
}

func Parse(query string) *ParsedQuery {
	parts := make([]*QueryPart, 0)

	for _, query := range SplitQuery(query) {
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

	return &ParsedQuery{
		Parts: parts,
	}
}

func SplitQuery(query string) []string {
	return strings.Split(query, " | ")
}
