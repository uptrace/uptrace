package mql

import (
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
	"golang.org/x/exp/constraints"
)

const sep = '0'

type Timeseries struct {
	MetricName   string
	NameTemplate string
	Filters      []ast.Filter
	Unit         string

	Attrs       Attrs
	Annotations map[string]any

	Value []float64
	Time  []time.Time

	Grouping   []string
	GroupByAll bool
}

func newTimeseriesFrom(ts *Timeseries) Timeseries {
	return Timeseries{
		MetricName:   ts.MetricName,
		NameTemplate: ts.NameTemplate,
		Filters:      ts.Filters,
		Unit:         ts.Unit,

		Attrs:       ts.Attrs,
		Annotations: ts.Annotations,

		Value: make([]float64, len(ts.Value)),
		Time:  ts.Time,

		Grouping:   ts.Grouping,
		GroupByAll: ts.GroupByAll,
	}
}

func (ts *Timeseries) Name() string {
	return buildName(ts.NameTemplate, ts.Filters, ts.Attrs)
}

func buildName(template string, filters []ast.Filter, attrs Attrs) string {
	if len(filters) == 0 && len(attrs) == 0 {
		return strings.ReplaceAll(template, "$$", "")
	}

	b := make([]byte, 0, 10*(len(filters)+len(attrs)))
	b = append(b, '{')

	for i := range filters {
		if i > 0 {
			b = append(b, ',')
		}
		b = filters[i].AppendString(b)
	}

	if len(attrs) > 0 {
		if len(filters) > 0 {
			b = append(b, ',')
		}
		b = attrs.AppendString(b, ",")
	}

	b = append(b, '}')

	return strings.ReplaceAll(template, "$$", unsafeconv.String(b))
}

func (ts *Timeseries) WhereQuery() string {
	if len(ts.Attrs) == 0 {
		return ""
	}

	b := make([]byte, 0, len(ts.Attrs)*30)
	for i, kv := range ts.Attrs {
		if i > 0 {
			b = append(b, " | "...)
		}
		b = append(b, "where "...)
		b = append(b, kv.Key...)
		b = append(b, " = "...)
		b = strconv.AppendQuote(b, kv.Value)
	}
	return unsafeconv.String(b)
}

func (ts *Timeseries) Clone() *Timeseries {
	clone := *ts
	return &clone
}

type TimeseriesFilter struct {
	Metric     string
	AggFunc    string
	TableFunc  string
	Uniq       []string
	Filters    []ast.Filter
	Where      [][]ast.Filter
	Grouping   []string
	GroupByAll bool
}

func min[T constraints.Ordered](a, b T) T {
	if a <= b {
		return a
	}
	return b
}
