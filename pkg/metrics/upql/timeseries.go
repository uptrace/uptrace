package upql

import (
	"strconv"
	"time"

	"github.com/uptrace/uptrace/pkg/metrics/upql/ast"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
	"golang.org/x/exp/constraints"
)

const sep = '0'

type Timeseries struct {
	Metric  string
	Func    string
	Filters []ast.Filter
	Unit    string

	Attrs       Attrs
	Annotations map[string]any

	Value []float64
	Time  []time.Time

	Grouping   []string
	GroupByAll bool
}

func newTimeseriesFrom(ts *Timeseries) Timeseries {
	return Timeseries{
		Metric:  ts.Metric,
		Func:    ts.Func,
		Filters: ts.Filters,
		Unit:    ts.Unit,

		Attrs:       ts.Attrs,
		Annotations: ts.Annotations,

		Value: make([]float64, len(ts.Value)),
		Time:  ts.Time,

		Grouping:   ts.Grouping,
		GroupByAll: ts.GroupByAll,
	}
}

func (ts *Timeseries) Name() string {
	b := appendName(nil, ts.Func, ts.Metric, ts.Filters, ts.Attrs)
	return unsafeconv.String(b)
}

func (ts *Timeseries) MetricName() string {
	b := appendName(nil, ts.Func, ts.Metric, ts.Filters, nil)
	return unsafeconv.String(b)
}

func appendName(b []byte, funcName, metric string, filters []ast.Filter, attrs Attrs) []byte {
	if funcName != "" {
		b = append(b, funcName...)
		b = append(b, '(')
	}

	b = append(b, metric...)

	if len(filters) > 0 || len(attrs) > 0 {
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
			b = attrs.AppendString(b)
		}

		b = append(b, '}')
	}

	if funcName != "" {
		b = append(b, ')')
	}

	return b
}

func (ts *Timeseries) WhereQuery() string {
	if len(ts.Attrs) == 0 {
		return ""
	}

	b := make([]byte, 0, len(ts.Attrs)*30)
	b = append(b, "where "...)
	for i, kv := range ts.Attrs {
		if i > 0 {
			b = append(b, " and "...)
		}
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
	Func       string
	Attr       string
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
