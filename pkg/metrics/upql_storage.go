package metrics

import (
	"context"
	"fmt"
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
	ProjectID uint32
	org.TimeFilter
	MetricMap map[string]*Metric

	TableName      ch.Ident
	GroupingPeriod time.Duration

	GroupByTime bool
	FillHoles   bool
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

func (s *CHStorage) MakeTimeseries() []upql.Timeseries {
	if !s.conf.GroupByTime {
		return []upql.Timeseries{{
			Value: []float64{0},
		}}
	}

	size := int(s.conf.TimeFilter.Duration() / s.conf.GroupingPeriod)
	return []upql.Timeseries{{
		Value: make([]float64, size),
		Time:  make([]time.Time, size),
	}}
}

func (s *CHStorage) SelectTimeseries(f *upql.TimeseriesFilter) ([]upql.Timeseries, error) {
	metric, ok := s.conf.MetricMap[f.Metric]
	if !ok {
		return nil, fmt.Errorf("can't find metric by alias %q", f.Metric)
	}

	subq, err := s.subquery(s.db.NewSelect(), metric, f)
	if err != nil {
		return nil, err
	}

	q := s.db.NewSelect().
		TableExpr("(?)", subq).
		ColumnExpr("metric").
		ColumnExpr("groupArray(value) AS value").
		GroupExpr("metric").
		Limit(10000)

	if len(f.Grouping) > 0 {
		for _, attrKey := range f.Grouping {
			q = q.Column(attrKey).Group(attrKey)
		}
	} else if f.GroupByAll {
		q = q.
			ColumnExpr("anyLast(attr_keys) AS attr_keys").
			ColumnExpr("anyLast(attr_values) AS attr_values").
			GroupExpr("attrs_hash")
	}
	if s.conf.GroupByTime {
		q = q.ColumnExpr("groupArray(time_) AS time")
	}

	var ms []map[string]any

	if err := q.Scan(s.ctx, &ms); err != nil {
		return nil, err
	}

	timeseries := make([]upql.Timeseries, 0, len(ms))

	for _, m := range ms {
		var metricName string
		if f.Func != "" {
			metricName = f.Func + "(" + f.Metric + ")"
		} else {
			metricName = f.Metric
		}

		timeseries = append(timeseries, upql.Timeseries{
			Metric:  metricName,
			Filters: f.Filters,
			Value:   m["value"].([]float64),
			Unit:    metricUnit(metric, f),
		})
		ts := &timeseries[len(timeseries)-1]

		if s.conf.GroupByTime {
			ts.Time = m["time"].([]time.Time)
		}

		if s.conf.GroupByTime && s.conf.FillHoles {
			ts.Value = bunutil.FillOrdered(
				ts.Value, ts.Time, s.conf.TimeGTE, s.conf.TimeLT, s.conf.GroupingPeriod)
			ts.Time = bunutil.FillTime(ts.Time, s.conf.TimeGTE, s.conf.TimeLT, s.conf.GroupingPeriod)
		}

		if len(f.Grouping) > 0 {
			delete(m, "metric")
			delete(m, "value")
			delete(m, "time")
			ts.Attrs = upql.AttrsFromMap(m)
		} else if f.GroupByAll {
			keys := m["attr_keys"].([]string)
			values := m["attr_values"].([]string)
			ts.Attrs = upql.AttrsFromKeysValues(keys, values)
		}
	}

	return timeseries, nil
}

func (s *CHStorage) subquery(
	q *ch.SelectQuery,
	metric *Metric,
	f *upql.TimeseriesFilter,
) (_ *ch.SelectQuery, err error) {
	q = q.
		ColumnExpr("metric").
		TableExpr("?", s.conf.TableName).
		Where("project_id = ?", s.conf.ProjectID).
		Where("metric = ?", metric.Name).
		Where("time >= ?", s.conf.TimeGTE).
		Where("time < ?", s.conf.TimeLT).
		GroupExpr("metric")

	if len(f.Filters) > 0 {
		q, err = s.filters(q, f.Filters)
		if err != nil {
			return nil, err
		}
	}
	for _, filters := range f.Where {
		q, err = s.filters(q, filters)
		if err != nil {
			return nil, err
		}
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
			ColumnExpr("anyLast(attr_keys) AS attr_keys").
			ColumnExpr("anyLast(attr_values) AS attr_values").
			GroupExpr("attrs_hash")
	}

	if s.conf.GroupByTime {
		q = q.
			ColumnExpr("toStartOfInterval(time, INTERVAL ? minute) AS time_",
				s.conf.GroupingPeriod.Minutes()).
			GroupExpr("time_").
			OrderExpr("time_")
	}

	isValueInstrument := isValueInstrument(metric.Instrument)
	if isValueInstrument {
		q = q.
			ColumnExpr("argMax(value, time) AS value").
			GroupExpr("attrs_hash")

		q = s.db.NewSelect().
			ColumnExpr("metric").
			TableExpr("(?) AS wrapper", q).
			GroupExpr("metric")
	}

	q, err = s.agg(q, metric, f)
	if err != nil {
		return nil, err
	}

	if isValueInstrument {
		if len(f.Grouping) > 0 {
			for _, attrKey := range f.Grouping {
				q = q.Column(attrKey).Group(attrKey)
			}
		} else if f.GroupByAll {
			q = q.
				ColumnExpr("attrs_hash").
				ColumnExpr("anyLast(attr_keys) AS attr_keys").
				ColumnExpr("anyLast(attr_values) AS attr_values").
				GroupExpr("attrs_hash")
		}

		if s.conf.GroupByTime {
			q = q.ColumnExpr("time_").GroupExpr("time_").OrderExpr("time_ ASC")
		}
	}

	return q, nil
}

func (s *CHStorage) filters(q *ch.SelectQuery, filters []ast.Filter) (*ch.SelectQuery, error) {
	var b []byte
	for i := range filters {
		filter := &filters[i]

		col := CHColumn(filter.LHS)

		if i > 0 {
			b = append(b, ' ')
			if filter.Sep != "" {
				b = append(b, filter.Sep...)
			} else {
				b = append(b, ast.BoolAnd...)
			}
			b = append(b, ' ')
		}

		switch filter.Op {
		case ast.FilterEqual:
			b = chschema.AppendQuery(b, "? = ?", col, filter.RHS)
		case ast.FilterNotEqual:
			b = chschema.AppendQuery(b, "? != ?", col, filter.RHS)
		case ast.FilterRegexp:
			b = chschema.AppendQuery(b, "match(?, ?)", col, filter.RHS)
		case ast.FilterNotRegexp:
			b = chschema.AppendQuery(b, "NOT match(?, ?)", col, filter.RHS)
		case ast.FilterLike:
			b = chschema.AppendQuery(b, "? LIKE ?", col, filter.RHS)
		case ast.FilterNotLike:
			b = chschema.AppendQuery(b, "? NOT LIKE ?", col, filter.RHS)
		default:
			return nil, fmt.Errorf("unsupported op: %s", filter.Op)
		}
	}
	return q.Where(unsafeconv.String(b)), nil
}

func (s *CHStorage) agg(
	q *ch.SelectQuery,
	metric *Metric,
	f *upql.TimeseriesFilter,
) (*ch.SelectQuery, error) {
	if f.Func == "rate" {
		f.Func = "per_min"
	}

	switch metric.Instrument {
	case InvalidInstrument:
		return nil, fmt.Errorf("metric %q not found", metric.Name)

	case CounterInstrument:
		if f.Func == "" {
			f.Func = "per_min"
		}

		switch f.Func {
		case "per_min", "per_minute":
			q = q.ColumnExpr("sum(sum) / ? AS value",
				s.conf.GroupingPeriod.Minutes())
			return q, nil
		case "per_sec", "per_second":
			q = q.ColumnExpr("sum(sum) / ? AS value",
				s.conf.GroupingPeriod.Seconds())
			return q, nil
		case "sum":
			q = q.ColumnExpr("sum(sum) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.Func)
		}

	case GaugeInstrument:
		switch f.Func {
		case "", "avg":
			q = q.ColumnExpr("avg(value) AS value")
			return q, nil
		case "sum":
			q = q.ColumnExpr("sum(value) AS value")
			return q, nil
		case "min", "max":
			q = q.ColumnExpr("?(value) AS value", ch.Safe(f.Func))
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.Func)
		}

	case AdditiveInstrument:
		switch f.Func {
		case "", "sum":
			q = q.ColumnExpr("sum(value) AS value")
			return q, nil
		case "avg":
			q = q.ColumnExpr("avg(value) AS value")
			return q, nil
		case "min", "max":
			q = q.ColumnExpr("?(value) AS value", ch.Safe(f.Func))
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.Func)
		}

	case HistogramInstrument:
		if f.Func == "" {
			f.Func = "p50"
		}

		switch f.Func {
		case "count":
			q = q.ColumnExpr("sum(count) AS value")
			return q, nil
		case "per_min", "per_minute":
			q = q.ColumnExpr("sum(count) / ? AS value",
				s.conf.GroupingPeriod.Minutes())
			return q, nil
		case "per_sec", "per_second":
			q = q.ColumnExpr("sum(count) / ? AS value",
				s.conf.GroupingPeriod.Seconds())
			return q, nil
		case "min":
			q = quantileColumn(q, 0)
			return q, nil
		case "max":
			q = quantileColumn(q, 1)
			return q, nil
		case "avg":
			q = q.ColumnExpr("sum(sum) / sum(count) AS value")
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

func quantileColumn(q *ch.SelectQuery, quantile float64) *ch.SelectQuery {
	return q.ColumnExpr("quantilesBFloat16Merge(?)(histogram)[1] AS value", quantile)
}

func metricUnit(metric *Metric, f *upql.TimeseriesFilter) string {
	switch metric.Instrument {
	case CounterInstrument, GaugeInstrument, AdditiveInstrument:
		return metric.Unit
	case HistogramInstrument:
		switch f.Func {
		case "count", "per_min", "per_minute", "per_sec", "per_second":
			return bununit.None
		default:
			return metric.Unit
		}
	default:
		return bununit.None
	}
}

func isValueInstrument(instrument string) bool {
	switch instrument {
	case GaugeInstrument, AdditiveInstrument:
		return true
	default:
		return false
	}
}

func unsupportedInstrumentFunc(instrument, fn string) error {
	return fmt.Errorf("%s does not support %s", instrument, fn)
}

func CHColumn(key string) ch.Safe {
	return ch.Safe(AppendCHColumn(nil, key))
}

func AppendCHColumn(b []byte, key string) []byte {
	return chschema.AppendQuery(b, "attr_values[indexOf(attr_keys, ?)]", key)
}
