package metrics

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/uptrace/pkg/bununit"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

type CHStorageConfig struct {
	org.TimeFilter

	ProjectID uint32
	MetricMap map[string]*Metric
	Search    string

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

var _ mql.Storage = (*CHStorage)(nil)

func (s *CHStorage) Consts() map[string]float64 {
	return map[string]float64{
		"_seconds": s.conf.GroupingPeriod.Seconds(),
		"_minutes": s.conf.GroupingPeriod.Minutes(),
	}
}

func (s *CHStorage) MakeTimeseries(f *mql.TimeseriesFilter) []mql.Timeseries {
	var ts mql.Timeseries

	if f != nil {
		ts.Filters = f.Filters
		ts.Grouping = f.Grouping
		ts.GroupByAll = f.GroupByAll
		if metric, ok := s.conf.MetricMap[f.Metric]; ok {
			ts.Unit = metric.Unit
		}
	}

	if s.conf.TableMode {
		ts.Value = []float64{0}
		return []mql.Timeseries{ts}
	}

	size := int(s.conf.TimeFilter.Duration() / s.conf.GroupingPeriod)
	ts.Value = make([]float64, size)
	ts.Time = make([]time.Time, size)
	for i := range ts.Time {
		ts.Time[i] = s.conf.TimeGTE.Add(time.Duration(i) * s.conf.GroupingPeriod)
	}

	return []mql.Timeseries{ts}
}

func (s *CHStorage) SelectTimeseries(f *mql.TimeseriesFilter) ([]mql.Timeseries, error) {
	metric, ok := s.conf.MetricMap[f.Metric]
	if !ok {
		return nil, fmt.Errorf("can't find metric with alias %q", f.Metric)
	}

	q := s.db.NewSelect().
		ColumnExpr("metric").
		ColumnExpr("max(annotations) AS annotations").
		TableExpr("?", s.conf.TableName).
		Where("project_id = ?", s.conf.ProjectID).
		Where("metric = ?", metric.Name).
		Where("time >= ?", s.conf.TimeGTE).
		Where("time < ?", s.conf.TimeLT).
		GroupExpr("metric")

	if s.conf.Search != "" {
		values := strings.Split(s.conf.Search, "|")
		q = q.Where("arrayExists(x -> multiSearchAnyCaseInsensitiveUTF8(x, ?) != 0, string_values)",
			ch.Array(values))
	}

	subq, err := s.subquery(q, metric, f)
	if err != nil {
		return nil, err
	}

	q = s.db.NewSelect().
		ColumnExpr("metric").
		ColumnExpr("max(annotations) AS annotations").
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
	f *mql.TimeseriesFilter,
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

	shouldDedup := f.AggFunc != "uniq" && isValueInstrument(metric.Instrument)

	if shouldDedup {
		switch f.AggFunc {
		case "min":
			q = q.ColumnExpr("min(gauge) AS gauge")
		case "max":
			q = q.ColumnExpr("max(gauge) AS gauge")
		default:
			q = q.ColumnExpr("argMax(gauge, time) AS gauge")
		}

		q = q.GroupExpr("attrs_hash")

		q = s.db.NewSelect().
			ColumnExpr("metric").
			ColumnExpr("max(annotations) AS annotations").
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
		var val any

		switch rhs := filter.RHS.(type) {
		case *ast.Number:
			val = rhs.Text
		case ast.StringValue:
			val = rhs.Text
		case ast.StringValues:
			var b []byte
			for i, text := range rhs.Texts {
				if i > 0 {
					b = append(b, ", "...)
				}
				b = chschema.AppendString(b, text)
			}
			val = ch.Safe(b)
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
		case ast.FilterIn:
			b = chschema.AppendQuery(b, "? IN (?)", col, val)
		case ast.FilterNotIn:
			b = chschema.AppendQuery(b, "? NOT IN (?)", col, val)
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
	f *mql.TimeseriesFilter,
) (*ch.SelectQuery, error) {
	if f.AggFunc == mql.AggUniq {
		var b []byte
		b = append(b, "uniqCombined64("...)

		if len(f.Uniq) == 0 {
			b = append(b, "attrs_hash"...)
		}

		for i, attr := range f.Uniq {
			if i > 0 {
				b = append(b, ',')
			}
			b = chschema.AppendQuery(b, "string_values[indexOf(string_keys, ?)]", attr)
		}

		b = append(b, ") AS value"...)
		q = q.ColumnExpr(unsafeconv.String(b))
		return q, nil
	}

	switch metric.Instrument {
	case InstrumentDeleted:
		return nil, fmt.Errorf("metric %q not found", metric.Name)

	case InstrumentCounter:
		switch f.AggFunc {
		case "", mql.AggSum:
			q = q.ColumnExpr("sumWithOverflow(sum) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.AggFunc)
		}

	case InstrumentGauge:
		switch f.AggFunc {
		case "", mql.AggAvg:
			q = q.ColumnExpr("avg(gauge) AS value")
			return q, nil
		case mql.AggSum: // may be okay
			q = q.ColumnExpr("sumWithOverflow(gauge) AS value")
			return q, nil
		case mql.AggMin:
			q = q.ColumnExpr("min(gauge) AS value")
			return q, nil
		case mql.AggMax:
			q = q.ColumnExpr("max(gauge) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.AggFunc)
		}

	case InstrumentAdditive:
		switch f.AggFunc {
		case "", mql.AggSum:
			// Sum last values with different attributes, for example,
			// fs.usage{state="free"} + fs.usage{state="reserved"}.
			q = q.ColumnExpr("sumWithOverflow(gauge) AS value")
			return q, nil
		case mql.AggAvg: // may be okay
			q = q.ColumnExpr("avg(gauge) AS value")
			return q, nil
		case mql.AggMin:
			q = q.ColumnExpr("min(gauge) AS value")
			return q, nil
		case mql.AggMax:
			q = q.ColumnExpr("max(gauge) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.AggFunc)
		}

	case InstrumentSummary:
		switch f.AggFunc {
		case mql.AggCount:
			q = q.ColumnExpr("sumWithOverflow(count) AS value")
			return q, nil
		case mql.AggAvg:
			q = q.ColumnExpr("sumWithOverflow(sum) / sumWithOverflow(count) AS value")
			return q, nil
		case mql.AggSum:
			q = q.ColumnExpr("sumWithOverflow(sum) AS value")
			return q, nil
		case mql.AggMin:
			q = q.ColumnExpr("min(min) AS value")
			return q, nil
		case mql.AggMax:
			q = q.ColumnExpr("max(max) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.AggFunc)
		}

	case InstrumentHistogram:
		switch f.AggFunc {
		case mql.AggAvg:
			q = q.ColumnExpr("sumWithOverflow(sum) / sumWithOverflow(count) AS value")
			return q, nil
		case mql.AggMin:
			q = q.ColumnExpr("min(min) AS value")
			return q, nil
		case mql.AggMax:
			q = q.ColumnExpr("max(max) AS value")
			return q, nil
		case mql.AggP50:
			q = quantileColumn(q, 0.5)
			return q, nil
		case mql.AggP75:
			q = quantileColumn(q, 0.75)
			return q, nil
		case mql.AggP90:
			q = quantileColumn(q, 0.9)
			return q, nil
		case mql.AggP95:
			q = quantileColumn(q, 0.95)
			return q, nil
		case mql.AggP99:
			q = quantileColumn(q, 0.99)
			return q, nil
		case mql.AggCount:
			q = q.ColumnExpr("sumWithOverflow(count) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.AggFunc)
		}

	default:
		return nil, fmt.Errorf("unsupported instrument %q", metric.Instrument)
	}
}

func (s *CHStorage) newTimeseries(
	metric *Metric, f *mql.TimeseriesFilter, items []map[string]any,
) ([]mql.Timeseries, error) {
	timeseries := make([]mql.Timeseries, 0, len(items))

	for _, m := range items {
		timeseries = append(timeseries, mql.Timeseries{
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

		if annotations, _ := m["annotations"].(string); annotations != "" {
			if err := json.Unmarshal([]byte(annotations), &ts.Annotations); err != nil {
				return nil, err
			}
		}

		if s.conf.TableMode {
			ts.Time = nil
			ts.Value = []float64{tableValue(ts.Value, metric.Instrument, f.AggFunc, f.TableFunc)}
		}

		switch {
		case len(f.Grouping) > 0:
			attrs := make([]string, 0, 2*len(f.Grouping))
			for _, attrKey := range f.Grouping {
				attrs = append(attrs, attrKey, fmt.Sprint(m[attrKey]))
				delete(m, attrKey)
			}
			ts.Attrs = mql.NewAttrs(attrs...)
		case f.GroupByAll:
			keys := m["string_keys"].([]string)
			values := m["string_values"].([]string)
			ts.Attrs = mql.AttrsFromKeysValues(keys, values)
			delete(m, "string_keys")
			delete(m, "string_values")
		}
	}

	return timeseries, nil
}

func quantileColumn(q *ch.SelectQuery, quantile float64) *ch.SelectQuery {
	return q.ColumnExpr("quantileBFloat16Merge(?)(histogram) AS value", quantile)
}

func metricUnit(metric *Metric, f *mql.TimeseriesFilter) string {
	switch f.AggFunc {
	case mql.AggCount, mql.AggUniq:
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

//------------------------------------------------------------------------------

func tableValue(
	value []float64, instrument Instrument, aggFunc, tableFunc string,
) float64 {
	if aggFunc == mql.FuncDelta {
		switch tableFunc {
		// TODO: support min, max, sum, avg
		}
	}

	var funcName string
	if tableFunc != "" {
		funcName = tableFunc
	} else {
		funcName = aggFunc
	}

	switch funcName {
	case "":
		// continue below
	case mql.TableMin:
		return minTableValue(value)
	case mql.TableMax:
		return maxTableValue(value)
	case mql.TableAvg,
		mql.FuncPerMin, mql.FuncPerSec,
		mql.AggP50, mql.AggP75, mql.AggP90, mql.AggP95, mql.AggP99:
		return avgTableValue(value)
	case mql.TableSum, mql.AggCount:
		return sumTableValue(value)
	case mql.FuncDelta:
		return deltaTableValue(value)
	case mql.TableLast, mql.AggUniq:
		return lastTableValue(value)
	default:
		return lastTableValue(value)
	}

	switch instrument {
	case InstrumentCounter:
		return sumTableValue(value)
	default:
		return lastTableValue(value)
	}
}

func minTableValue(ns []float64) float64 {
	min := math.MaxFloat64
	for _, n := range ns {
		if math.IsNaN(n) {
			continue
		}
		if n < min {
			min = n
		}
	}
	if min != math.MaxFloat64 {
		return min
	}
	return 0
}

func maxTableValue(ns []float64) float64 {
	var max float64
	for _, n := range ns {
		if math.IsNaN(n) {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max
}

func lastTableValue(ns []float64) float64 {
	end := len(ns) - 3
	if end < 0 {
		end = 0
	}

	for i := len(ns) - 1; i >= end; i-- {
		n := ns[i]
		if !math.IsNaN(n) {
			return n
		}
	}
	return 0
}

func avgTableValue(ns []float64) float64 {
	sum, count := sumCount(ns)
	return sum / float64(count)
}

func sumTableValue(ns []float64) float64 {
	sum, _ := sumCount(ns)
	return sum
}

func sumCount(ns []float64) (float64, int) {
	var sum float64
	var count int
	for _, n := range ns {
		if !math.IsNaN(n) {
			sum += n
			count++
		}
	}
	return sum, count
}

func deltaTableValue(value []float64) float64 {
	for i, num := range value {
		if !math.IsNaN(num) {
			value = value[i:]
			break
		}
	}

	if len(value) == 0 {
		return 0
	}

	prevNum := value[0]
	value = value[1:]
	var sum float64

	for _, num := range value {
		if math.IsNaN(num) {
			continue
		}
		sum += num - prevNum
		prevNum = num
	}

	return sum
}
