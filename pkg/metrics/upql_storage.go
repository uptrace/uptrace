package metrics

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/uptrace/pkg/bununit"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"github.com/uptrace/uptrace/pkg/metrics/upql/ast"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

type CHStorageConfig struct {
	org.TimeFilter

	ProjectID uint32
	MetricMap map[string]*Metric

	TableName      ch.Ident
	TableMode      bool
	GroupingPeriod time.Duration
}

type CHStorage struct {
	ctx  context.Context
	conf *CHStorageConfig

	db *ch.DB
}

func NewCHStorage(ctx context.Context, db *ch.DB, conf *CHStorageConfig) *CHStorage {
	s := &CHStorage{
		ctx:  ctx,
		db:   db,
		conf: conf,
	}
	return s
}

var _ upql.Storage = (*CHStorage)(nil)

func (s *CHStorage) MakeTimeseries(f *upql.TimeseriesFilter) []upql.Timeseries {
	var ts upql.Timeseries

	if f != nil {
		ts.Metric = f.Metric
		ts.Func = f.Func
		ts.Filters = f.Filters
		ts.Grouping = f.Grouping
		ts.GroupByAll = f.GroupByAll
		if metric, ok := s.conf.MetricMap[f.Metric]; ok {
			ts.Unit = metric.Unit
		}
	}

	if s.conf.TableMode {
		ts.Value = []float64{0}
		return []upql.Timeseries{ts}
	}

	size := int(s.conf.TimeFilter.Duration() / s.conf.GroupingPeriod)
	ts.Value = make([]float64, size)
	ts.Time = make([]time.Time, size)
	for i := range ts.Time {
		ts.Time[i] = s.conf.TimeGTE.Add(time.Duration(i) * s.conf.GroupingPeriod)
	}

	return []upql.Timeseries{ts}
}

func (s *CHStorage) SelectTimeseries(f *upql.TimeseriesFilter) ([]upql.Timeseries, error) {
	metric, ok := s.conf.MetricMap[f.Metric]
	if !ok {
		return nil, fmt.Errorf("can't find metric with alias %q", f.Metric)
	}

	q := s.db.NewSelect().
		ColumnExpr("metric").
		TableExpr("?", s.conf.TableName).
		Where("project_id = ?", s.conf.ProjectID).
		Where("metric = ?", metric.Name).
		Where("time >= ?", s.conf.TimeGTE).
		Where("time < ?", s.conf.TimeLT).
		GroupExpr("metric")

	subq, err := s.subquery(q, metric, f)
	if err != nil {
		return nil, err
	}

	q = s.db.NewSelect().
		ColumnExpr("metric").
		ColumnExpr("groupArray(toFloat64(value)) AS value").
		ColumnExpr("groupArray(time_) AS time").
		TableExpr("(?)", subq).
		GroupExpr("metric").
		Limit(10000)

	if len(f.Grouping) > 0 {
		for _, attrKey := range f.Grouping {
			q = q.Column(attrKey).Group(attrKey)
		}
	} else if f.GroupByAll {
		q = q.
			ColumnExpr("anyLast(string_keys) AS string_keys").
			ColumnExpr("anyLast(string_values) AS string_values").
			GroupExpr("attrs_hash")
	}

	var items []map[string]any

	if err := q.Scan(s.ctx, &items); err != nil {
		return nil, err
	}

	return s.newTimeseries(metric, f, items)
}

func (s *CHStorage) subquery(
	q *ch.SelectQuery,
	metric *Metric,
	f *upql.TimeseriesFilter,
) (_ *ch.SelectQuery, err error) {
	if len(f.Filters) > 0 {
		where, err := compileFilters(f.Filters)
		if err != nil {
			return nil, err
		}
		q = q.Where(where)
	}
	for _, filters := range f.Where {
		where, err := compileFilters(filters)
		if err != nil {
			return nil, err
		}
		q = q.Where(where)
	}

	if len(f.Grouping) > 0 {
		for _, attrKey := range f.Grouping {
			col := CHColumn(attrKey)
			q = q.
				ColumnExpr("? AS ?", col, ch.Ident(attrKey)).
				GroupExpr("?", col)
		}
	} else if f.GroupByAll {
		q = q.
			ColumnExpr("attrs_hash").
			ColumnExpr("anyLast(string_keys) AS string_keys").
			ColumnExpr("anyLast(string_values) AS string_values").
			GroupExpr("attrs_hash")
	}

	q = q.
		ColumnExpr("toStartOfInterval(time, INTERVAL ? minute) AS time_",
			s.conf.GroupingPeriod.Minutes()).
		GroupExpr("time_").
		OrderExpr("time_")

	shouldDedup := f.Func != "uniq" && isValueInstrument(metric.Instrument)

	if shouldDedup {
		switch f.Func {
		case "min":
			q = q.ColumnExpr("min(value) AS value")
		case "max":
			q = q.ColumnExpr("max(value) AS value")
		default:
			q = q.ColumnExpr("argMax(value, time) AS value")
		}

		q = q.GroupExpr("attrs_hash")

		q = s.db.NewSelect().
			ColumnExpr("metric").
			TableExpr("(?)", q).
			GroupExpr("metric")
	}

	q, err = s.agg(q, metric, f)
	if err != nil {
		return nil, err
	}

	if shouldDedup {
		if len(f.Grouping) > 0 {
			for _, attrKey := range f.Grouping {
				q = q.Column(attrKey).Group(attrKey)
			}
		} else if f.GroupByAll {
			q = q.
				ColumnExpr("attrs_hash").
				ColumnExpr("anyLast(string_keys) AS string_keys").
				ColumnExpr("anyLast(string_values) AS string_values").
				GroupExpr("attrs_hash")
		}

		q = q.ColumnExpr("time_").GroupExpr("time_").OrderExpr("time_ ASC")
	}

	return q, nil
}

func compileFilters(filters []ast.Filter) (string, error) {
	var b []byte
	for i := range filters {
		filter := &filters[i]

		attrKey := filter.LHS

		if filter.Op == ast.FilterExists {
			b = chschema.AppendQuery(b, "has(string_keys, ?)", attrKey)
			continue
		}

		col := CHColumn(filter.LHS)
		var val string

		switch rhs := filter.RHS.(type) {
		case *ast.Number:
			val = rhs.Text
		case ast.StringValue:
			val = rhs.Text
		default:
			return "", fmt.Errorf("unknown RHS: %T", rhs)
		}

		if i > 0 {
			b = append(b, ' ')
			if filter.BoolOp != "" {
				b = append(b, filter.BoolOp...)
			} else {
				b = append(b, ast.BoolAnd...)
			}
			b = append(b, ' ')
		}

		switch filter.Op {
		case ast.FilterEqual:
			b = chschema.AppendQuery(b, "? = ?", col, val)
		case ast.FilterNotEqual:
			b = chschema.AppendQuery(b, "? != ?", col, val)
		case ast.FilterRegexp:
			b = chschema.AppendQuery(b, "match(?, ?)", col, val)
		case ast.FilterNotRegexp:
			b = chschema.AppendQuery(b, "NOT match(?, ?)", col, val)
		case ast.FilterLike:
			b = chschema.AppendQuery(b, "? LIKE ?", col, val)
		case ast.FilterNotLike:
			b = chschema.AppendQuery(b, "? NOT LIKE ?", col, val)
		default:
			return "", fmt.Errorf("unsupported op: %s", filter.Op)
		}
	}
	return unsafeconv.String(b), nil
}

func (s *CHStorage) agg(
	q *ch.SelectQuery,
	metric *Metric,
	f *upql.TimeseriesFilter,
) (*ch.SelectQuery, error) {
	if f.Func == "uniq" {
		q = q.ColumnExpr(
			"uniqCombined64(string_values[indexOf(string_keys, ?)]) AS value", f.Attr)
		return q, nil
	}

	if f.Attr != "" {
		return nil, fmt.Errorf("unexpected attribute: %s", f.Attr)
	}
	if f.Func == "rate" {
		f.Func = "per_min"
	}

	switch metric.Instrument {
	case InstrumentDeleted:
		return nil, fmt.Errorf("metric %q not found", metric.Name)

	case InstrumentCounter:
		switch f.Func {
		case "per_min", "per_minute":
			q = q.ColumnExpr("sumWithOverflow(sum) / ? AS value",
				s.conf.GroupingPeriod.Minutes())
			return q, nil
		case "per_sec", "per_second":
			q = q.ColumnExpr("sumWithOverflow(sum) / ? AS value",
				s.conf.GroupingPeriod.Seconds())
			return q, nil
		case "", "sum":
			q = q.ColumnExpr("sumWithOverflow(sum) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.Func)
		}

	case InstrumentGauge:
		switch f.Func {
		case "", "avg", "last", "delta":
			q = q.ColumnExpr("avg(value) AS value")
			return q, nil
		case "sum": // may be okay
			q = q.ColumnExpr("sumWithOverflow(value) AS value")
			return q, nil
		case "min":
			q = q.ColumnExpr("min(value) AS value")
			return q, nil
		case "max":
			q = q.ColumnExpr("max(value) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.Func)
		}

	case InstrumentAdditive:
		switch f.Func {
		case "", "sum", "delta":
			q = q.ColumnExpr("sumWithOverflow(value) AS value")
			return q, nil
		case "avg", "last": // may be okay
			q = q.ColumnExpr("avg(value) AS value")
			return q, nil
		case "min":
			q = q.ColumnExpr("min(value) AS value")
			return q, nil
		case "max":
			q = q.ColumnExpr("max(value) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.Func)
		}

	case InstrumentSummary:
		switch f.Func {
		case "avg", "last":
			q = q.ColumnExpr("sumWithOverflow(sum) / sumWithOverflow(count) AS value")
			return q, nil
		case "sum":
			q = q.ColumnExpr("sumWithOverflow(sum) AS value")
			return q, nil
		case "count":
			q = q.ColumnExpr("sumWithOverflow(count) AS value")
			return q, nil
		case "per_min", "per_minute":
			q = q.ColumnExpr("sumWithOverflow(count) / ? AS value",
				s.conf.GroupingPeriod.Minutes())
			return q, nil
		case "per_sec", "per_second":
			q = q.ColumnExpr("sumWithOverflow(count) / ? AS value",
				s.conf.GroupingPeriod.Seconds())
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.Func)
		}

	case InstrumentHistogram:
		switch f.Func {
		case "count":
			q = q.ColumnExpr("sumWithOverflow(count) AS value")
			return q, nil
		case "per_min", "per_minute":
			q = q.ColumnExpr("sumWithOverflow(count) / ? AS value",
				s.conf.GroupingPeriod.Minutes())
			return q, nil
		case "per_sec", "per_second":
			q = q.ColumnExpr("sumWithOverflow(count) / ? AS value",
				s.conf.GroupingPeriod.Seconds())
			return q, nil
		case "min":
			q = quantileColumn(q, 0)
			return q, nil
		case "max":
			q = quantileColumn(q, 1)
			return q, nil
		case "avg", "last":
			q = q.ColumnExpr("sumWithOverflow(sum) / sumWithOverflow(count) AS value")
			return q, nil
		case "p50":
			q = quantileColumn(q, 0.5)
			return q, nil
		case "p75":
			q = quantileColumn(q, 0.75)
			return q, nil
		case "p90":
			q = quantileColumn(q, 0.9)
			return q, nil
		case "p95":
			q = quantileColumn(q, 0.95)
			return q, nil
		case "p99":
			q = quantileColumn(q, 0.99)
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.Func)
		}

	default:
		return nil, fmt.Errorf("unknown instrument %q", metric.Instrument)
	}
}

func (s *CHStorage) newTimeseries(
	metric *Metric, f *upql.TimeseriesFilter, items []map[string]any,
) ([]upql.Timeseries, error) {
	timeseries := make([]upql.Timeseries, 0, len(items))

	for _, m := range items {
		timeseries = append(timeseries, upql.Timeseries{
			Metric:     f.Metric,
			Func:       f.Func,
			Filters:    f.Filters,
			Unit:       metricUnit(metric, f),
			Grouping:   f.Grouping,
			GroupByAll: f.GroupByAll,
		})
		ts := &timeseries[len(timeseries)-1]

		ts.Value = m["value"].([]float64)
		delete(m, "value")

		ts.Time = m["time"].([]time.Time)
		delete(m, "time")

		ts.Value = bunutil.Fill(
			ts.Value,
			ts.Time,
			math.NaN(),
			s.conf.TimeGTE,
			s.conf.TimeLT,
			s.conf.GroupingPeriod,
		)
		ts.Time = bunutil.FillTime(ts.Time, s.conf.TimeGTE, s.conf.TimeLT, s.conf.GroupingPeriod)

		if s.conf.TableMode {
			ts.Time = nil
			ts.Value = []float64{s.tableValue(metric, f, ts.Value)}
		}

		switch {
		case len(f.Grouping) > 0:
			attrs := make([]string, 0, 2*len(f.Grouping))
			for _, attrKey := range f.Grouping {
				attrs = append(attrs, attrKey, fmt.Sprint(m[attrKey]))
				delete(m, attrKey)
			}
			ts.Attrs = upql.NewAttrs(attrs...)
		case f.GroupByAll:
			keys := m["string_keys"].([]string)
			values := m["string_values"].([]string)
			ts.Attrs = upql.AttrsFromKeysValues(keys, values)
			delete(m, "string_keys")
			delete(m, "string_values")
		}

		if len(m) > 0 {
			ts.Annotations = m
		}
	}

	return timeseries, nil
}

func (s *CHStorage) tableValue(
	metric *Metric, f *upql.TimeseriesFilter, value []float64,
) float64 {
	switch f.Func {
	case "":
		// continue below
	case "min":
		return minValue(value)
	case "max":
		return maxValue(value)
	case "avg",
		"per_min", "per_minute", "per_sec", "per_second",
		"p50", "p75", "p90", "p95", "p99":
		return avg(value)
	case "count":
		return sum(value)
	case "delta":
		return delta(value)
	case "last", "uniq":
		return last(value)
	default:
		return last(value)
	}

	switch metric.Instrument {
	case InstrumentCounter:
		return sum(value)
	default:
		return last(value)
	}
}

func quantileColumn(q *ch.SelectQuery, quantile float64) *ch.SelectQuery {
	return q.ColumnExpr("quantileBFloat16Merge(?)(histogram) AS value", quantile)
}

func metricUnit(metric *Metric, f *upql.TimeseriesFilter) string {
	switch f.Func {
	case "count", "per_min", "per_minute", "per_sec", "per_second", "uniq":
		return bununit.None
	default:
		return metric.Unit
	}
}

func isValueInstrument(instrument Instrument) bool {
	switch instrument {
	case InstrumentGauge, InstrumentAdditive:
		return true
	default:
		return false
	}
}

func unsupportedInstrumentFunc(instrument Instrument, funcName string) error {
	if funcName == "" {
		return fmt.Errorf("%s instrument requires a func", instrument)
	}
	return fmt.Errorf("%s instrument does not support %s", instrument, funcName)
}

func CHColumn(key string) ch.Safe {
	return ch.Safe(AppendCHColumn(nil, key))
}

func AppendCHColumn(b []byte, key string) []byte {
	return chschema.AppendQuery(b, "string_values[indexOf(string_keys, ?)]", key)
}
