package mql

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/unixtime"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

type Timeseries struct {
	MetricName   string
	NameTemplate string

	Unit     string
	Filters  []ast.Filter
	Grouping []string

	Attrs       Attrs
	Annotations map[string]any

	Value []float64
	Time  []unixtime.Seconds
}

func (ts *Timeseries) DeepClone() *Timeseries {
	clone := ts.Clone()
	clone.Value = slices.Clone(ts.Value)
	return clone
}

func (ts *Timeseries) Clone() *Timeseries {
	clone := *ts
	return &clone
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
		b = filters[i].AppendString(b, false)
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

type TimeseriesFilter struct {
	Metric string

	TimeGTE  time.Time
	TimeLT   time.Time
	Interval time.Duration

	CHFunc string
	Attr   string

	Uniq []string

	Filters  []ast.Filter
	Where    [][]ast.Filter
	Grouping ast.GroupingElems
}
