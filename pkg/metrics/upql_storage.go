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
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconv"
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

	TableName        string
	TableMode        bool
	GroupingInterval time.Duration
}

type CHStorage struct {
	ctx  context.Context
	app  *bunapp.App
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
		"_seconds": s.conf.GroupingInterval.Seconds(),
		"_minutes": s.conf.GroupingInterval.Minutes(),
	}
}

func (s *CHStorage) MakeTimeseries(f *mql.TimeseriesFilter) *mql.Timeseries {
	var ts mql.Timeseries

	if f != nil {
		ts.Filters = f.Filters
		ts.Grouping = f.Grouping.Attrs()
		if metric, ok := s.conf.MetricMap[f.Metric]; ok {
			ts.Unit = metric.Unit
		}
	}

	size := int(s.conf.TimeFilter.Duration() / s.conf.GroupingInterval)
	ts.Value = make([]float64, size)
	ts.Time = make([]time.Time, size)
	for i := range ts.Time {
		ts.Time[i] = s.conf.TimeGTE.Add(time.Duration(i) * s.conf.GroupingInterval)
	}

	return &ts
}

func (s *CHStorage) SelectTimeseries(f *mql.TimeseriesFilter) ([]*mql.Timeseries, error) {
	metric, ok := s.conf.MetricMap[f.Metric]
	if !ok {
		return nil, fmt.Errorf("can't find metric with alias %q", f.Metric)
	}

	q, err := s.compileQuery(metric, f)
	if err != nil {
		return nil, err
	}

	var items []map[string]any
	if err := q.Scan(s.ctx, &items); err != nil {
		return nil, err
	}

	return s.newTimeseries(metric, f, items)
}

func (s *CHStorage) compileQuery(metric *Metric, f *mql.TimeseriesFilter) (*ch.SelectQuery, error) {
	q := s.db.NewSelect().
		ColumnExpr("d.metric").
		ColumnExpr("max(d.annotations) AS annotations").
		TableExpr("? AS d", ch.Name(s.conf.TableName)).
		Where("d.project_id = ?", s.conf.ProjectID).
		Where("d.metric = ?", metric.Name).
		Where("d.time >= ?", s.conf.TimeGTE).
		Where("d.time < ?", s.conf.TimeLT).
		GroupExpr("d.metric")

	if s.conf.Search != "" {
		values := strings.Split(s.conf.Search, "|")
		q = q.WhereGroup(" AND ", func(q *ch.SelectQuery) *ch.SelectQuery {
			for _, elem := range f.Grouping {
				chExpr := CHExpr(elem.Name)
				q = q.WhereOr("multiSearchAnyCaseInsensitiveUTF8(?, ?) != 0",
					chExpr, ch.Array(values))
			}
			return q
		})
	}

	subq, err := s.subquery(q, metric, f)
	if err != nil {
		return nil, err
	}

	q = s.db.NewSelect().
		ColumnExpr("d.metric").
		ColumnExpr("max(d.annotations) AS annotations").
		ColumnExpr("groupArray(toFloat64(d.value)) AS value").
		ColumnExpr("groupArray(d.time) AS time").
		TableExpr("(?) AS d", subq).
		GroupExpr("d.metric").
		Limit(10000)

	if len(f.Grouping) > 0 {
		for _, elem := range f.Grouping {
			q = q.Column(elem.Alias).Group(elem.Alias)
		}
	}

	return q, nil
}

func (s *CHStorage) subquery(
	q *ch.SelectQuery,
	metric *Metric,
	f *mql.TimeseriesFilter,
) (_ *ch.SelectQuery, err error) {
	if len(f.Filters) > 0 {
		if err := compileFilters(q, metric.Instrument, f.Filters); err != nil {
			return nil, err
		}
	}
	for _, filters := range f.Where {
		if err := compileFilters(q, metric.Instrument, filters); err != nil {
			return nil, err
		}
	}

	for _, elem := range f.Grouping {
		chExpr, err := chGrouping(elem, CHExpr(elem.Name))
		if err != nil {
			return nil, err
		}
		q.ColumnExpr("? AS ?", chExpr, ch.Name(elem.Alias)).
			GroupExpr("?", chExpr)
	}

	q = q.
		ColumnExpr("toStartOfInterval(d.time, INTERVAL ? SECOND) AS time",
			s.conf.GroupingInterval.Seconds()).
		GroupExpr("time").
		OrderExpr("time")

	shouldDedup := isValueInstrument(metric.Instrument)

	if shouldDedup {
		switch f.CHFunc {
		case mql.CHAggMin:
			q = q.ColumnExpr("min(d.gauge) AS value")
		case mql.CHAggMax:
			q = q.ColumnExpr("max(d.gauge) AS value")
		case mql.CHAggUniq:
			if len(f.Uniq) == 0 {
				q = q.ColumnExpr("d.attrs_hash").GroupExpr("d.attrs_hash")
			} else {
				for _, attrKey := range f.Uniq {
					chExpr := CHExpr(attrKey)
					q = q.ColumnExpr("? AS ?", chExpr, ch.Name(attrKey)).
						GroupExpr(string(chExpr))
				}
			}
			q = q.ColumnExpr("argMax(d.gauge, d.time) AS value")
		case "", "_", mql.CHAggSum, mql.CHAggAvg:
			q = q.ColumnExpr("argMax(d.gauge, d.time) AS value").
				GroupExpr("d.attrs_hash")
		default:
			q = q.ColumnExpr("argMax(d.gauge, d.time) AS value")
		}

		q = s.db.NewSelect().
			ColumnExpr("d.metric").
			ColumnExpr("max(d.annotations) AS annotations").
			TableExpr("(?) AS d", q).
			GroupExpr("d.metric")
	}

	q, err = s.agg(q, metric, f)
	if err != nil {
		return nil, err
	}

	if shouldDedup {
		for _, elem := range f.Grouping {
			q = q.Column(elem.Alias).Group(elem.Alias)
		}

		q = q.ColumnExpr("time").GroupExpr("time").OrderExpr("time ASC")
	}

	return q, nil
}

func compileFilters(
	q *ch.SelectQuery, instrument Instrument, filters []ast.Filter,
) error {
	var where []byte
	var having []byte

	for i := range filters {
		filter := &filters[i]

		switch filter.LHS {
		case "", ".value", "_value":
			colVal, err := resolveNumberValue(filter.RHS)
			if err != nil {
				return err
			}

			having, err = appendFilter(having, filter, ch.Safe("value"), colVal)
			if err != nil {
				return err
			}

		default:
			if filter.Op == ast.FilterExists {
				where = appendExistsFilter(where, filter, filter.LHS)
				break
			}

			colVal, err := resolveStringValue(filter.RHS)
			if err != nil {
				return err
			}

			chExpr := CHExpr(filter.LHS)
			where, err = appendFilter(where, filter, chExpr, colVal)
			if err != nil {
				return err
			}
		}
	}

	if len(where) > 0 {
		q = q.Where(unsafeconv.String(where))
	}
	if len(having) > 0 {
		q = q.Having(unsafeconv.String(having))
	}
	return nil
}

func appendExistsFilter(b []byte, filter *ast.Filter, attrKey string) []byte {
	b = appendFilterSep(b, filter)
	b = chschema.AppendQuery(b, "has(d.string_keys, ?)", attrKey)
	return b
}

func appendFilter(b []byte, filter *ast.Filter, colName, colVal any) ([]byte, error) {
	b = appendFilterSep(b, filter)

	switch filter.Op {
	case ast.FilterEqual, ast.FilterNotEqual,
		ast.FilterLT, ast.FilterLTE,
		ast.FilterGT, ast.FilterGTE,
		ast.FilterLike, ast.FilterNotLike:
		b = chschema.AppendQuery(b, "? ? ?", colName, ch.Safe(filter.Op), colVal)
		return b, nil
	case ast.FilterIn, ast.FilterNotIn:
		b = chschema.AppendQuery(b, "? ? (?)", colName, ch.Safe(filter.Op), colVal)
		return b, nil
	case ast.FilterRegexp:
		b = chschema.AppendQuery(b, "match(?, ?)", colName, colVal)
		return b, nil
	case ast.FilterNotRegexp:
		b = chschema.AppendQuery(b, "NOT match(?, ?)", colName, colVal)
		return b, nil
	default:
		return nil, fmt.Errorf("unsupported op: %s", filter.Op)
	}
}

func appendFilterSep(b []byte, filter *ast.Filter) []byte {
	if len(b) == 0 {
		return b
	}
	b = append(b, ' ')
	if filter.BoolOp != "" {
		b = append(b, filter.BoolOp...)
	} else {
		b = append(b, ast.BoolAnd...)
	}
	b = append(b, ' ')
	return b
}

func resolveStringValue(value any) (any, error) {
	switch value := value.(type) {
	case ast.Number:
		return value.Text, nil
	case ast.StringValue:
		return value.Text, nil
	case ast.StringValues:
		var b []byte
		for i, text := range value.Values {
			if i > 0 {
				b = append(b, ", "...)
			}
			b = chschema.AppendString(b, text)
		}
		return ch.Safe(b), nil
	default:
		return "", fmt.Errorf("unsupported string value type: %T", value)
	}
}

func resolveNumberValue(value any) (float64, error) {
	switch value := value.(type) {
	case ast.Number:
		return value.Float64(), nil
	default:
		return 0, fmt.Errorf("unsupported number value type: %T", value)
	}
}

func (s *CHStorage) agg(
	q *ch.SelectQuery,
	metric *Metric,
	f *mql.TimeseriesFilter,
) (*ch.SelectQuery, error) {
	switch f.CHFunc {
	case mql.CHAggUniq:
		var b []byte
		b = append(b, "uniqCombined64("...)

		if len(f.Uniq) == 0 {
			b = append(b, "d.attrs_hash"...)
		} else {
			isValue := isValueInstrument(metric.Instrument)
			for i, attrKey := range f.Uniq {
				if i > 0 {
					b = append(b, ", "...)
				}
				if isValue {
					b = chschema.AppendName(b, attrKey)
				} else {
					b = AppendCHExpr(b, attrKey)
				}
			}
		}

		b = append(b, ") AS value"...)
		q = q.ColumnExpr(unsafeconv.String(b))
		return q, nil
	}

	switch f.Attr {
	case "":
		// continue below
	case ".time", "_time":
		switch f.CHFunc {
		case mql.CHAggMin:
			q = q.ColumnExpr("min(d.time) AS value")
			return q, nil
		case mql.CHAggMax:
			q = q.ColumnExpr("max(d.time) AS value")
			return q, nil
		default:
			return nil, unsupportedAttrFunc(f.Attr, f.CHFunc)
		}
	default:
		return nil, unsupportedAttrFunc(f.Attr, f.CHFunc)
	}

	switch metric.Instrument {
	case InstrumentDeleted:
		return nil, fmt.Errorf("metric %q not found", metric.Name)

	case InstrumentCounter:
		switch f.CHFunc {
		case "", "_", mql.CHAggSum:
			q = q.ColumnExpr("sumWithOverflow(d.sum) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.CHFunc)
		}

	case InstrumentGauge:
		switch f.CHFunc {
		case "", "_", mql.CHAggAvg:
			q = q.ColumnExpr("avg(d.value) AS value")
			return q, nil
		case mql.CHAggSum: // may be okay
			q = q.ColumnExpr("sumWithOverflow(d.value) AS value")
			return q, nil
		case mql.CHAggMin:
			q = q.ColumnExpr("min(d.value) AS value")
			return q, nil
		case mql.CHAggMax:
			q = q.ColumnExpr("max(d.value) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.CHFunc)
		}

	case InstrumentAdditive:
		switch f.CHFunc {
		case "", mql.CHAggSum:
			q = q.ColumnExpr("sumWithOverflow(d.value) AS value")
			return q, nil
		case mql.CHAggAvg: // may be okay
			q = q.ColumnExpr("avg(d.value) AS value")
			return q, nil
		case mql.CHAggMin:
			q = q.ColumnExpr("min(d.value) AS value")
			return q, nil
		case mql.CHAggMax:
			q = q.ColumnExpr("max(d.value) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.CHFunc)
		}

	case InstrumentSummary:
		switch f.CHFunc {
		case mql.CHAggCount:
			q = q.ColumnExpr("sumWithOverflow(d.count) AS value")
			return q, nil
		case mql.CHAggAvg:
			q = q.ColumnExpr("sumWithOverflow(d.sum) / sumWithOverflow(d.count) AS value")
			return q, nil
		case mql.CHAggSum:
			q = q.ColumnExpr("sumWithOverflow(d.sum) AS value")
			return q, nil
		case mql.CHAggMin:
			q = q.ColumnExpr("min(d.min) AS value")
			return q, nil
		case mql.CHAggMax:
			q = q.ColumnExpr("max(d.max) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.CHFunc)
		}

	case InstrumentHistogram:
		switch f.CHFunc {
		case mql.CHAggAvg:
			q = q.ColumnExpr("sumWithOverflow(d.sum) / sumWithOverflow(d.count) AS value")
			return q, nil
		case mql.CHAggMin:
			q = q.ColumnExpr("min(d.min) AS value")
			return q, nil
		case mql.CHAggMax:
			q = q.ColumnExpr("max(d.max) AS value")
			return q, nil
		case mql.CHAggP50:
			q = quantileColumn(q, 0.5)
			return q, nil
		case mql.CHAggP75:
			q = quantileColumn(q, 0.75)
			return q, nil
		case mql.CHAggP90:
			q = quantileColumn(q, 0.9)
			return q, nil
		case mql.CHAggP95:
			q = quantileColumn(q, 0.95)
			return q, nil
		case mql.CHAggP99:
			q = quantileColumn(q, 0.99)
			return q, nil
		case mql.CHAggCount:
			q = q.ColumnExpr("sumWithOverflow(d.count) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.CHFunc)
		}

	default:
		return nil, fmt.Errorf("unsupported instrument %q", metric.Instrument)
	}
}

func (s *CHStorage) newTimeseries(
	metric *Metric, f *mql.TimeseriesFilter, items []map[string]any,
) ([]*mql.Timeseries, error) {
	timeseries := make([]*mql.Timeseries, 0, len(items))

	for _, m := range items {
		ts := &mql.Timeseries{
			Unit:     metricUnit(metric, f),
			Filters:  f.Filters,
			Grouping: f.Grouping.Attrs(),
		}
		timeseries = append(timeseries, ts)

		delete(m, "metric")

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
			s.conf.GroupingInterval,
		)
		ts.Time = bunutil.FillTime(ts.Time, s.conf.TimeGTE, s.conf.TimeLT, s.conf.GroupingInterval)

		if annotations, _ := m["annotations"].(string); annotations != "" {
			if err := json.Unmarshal([]byte(annotations), &ts.Annotations); err != nil {
				return nil, err
			}
		}

		if len(f.Grouping) > 0 {
			attrs := make([]string, 0, 2*len(f.Grouping))
			for _, elem := range f.Grouping {
				attrs = append(attrs, elem.Alias, fmt.Sprint(m[elem.Alias]))
				delete(m, elem.Alias)
			}
			ts.Attrs = mql.NewAttrs(attrs...)
		}
	}

	return timeseries, nil
}

func quantileColumn(q *ch.SelectQuery, quantile float64) *ch.SelectQuery {
	return q.ColumnExpr("quantileBFloat16Merge(?)(d.histogram) AS value", quantile)
}

func metricUnit(metric *Metric, f *mql.TimeseriesFilter) string {
	switch f.Attr {
	case "":
		// continue below
	case ".time", "_time":
		return bunconv.UnitTime
	default:
		return ""
	}

	switch f.CHFunc {
	case mql.CHAggCount, mql.CHAggUniq:
		return bunconv.UnitNone
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

func unsupportedAttrFunc(attrName, funcName string) error {
	if funcName == "" {
		return fmt.Errorf("%s attr requires a func", attrName)
	}
	return fmt.Errorf("%s attr does not support %s", attrName, funcName)

}

func CHExpr(key string) ch.Safe {
	return ch.Safe(AppendCHExpr(nil, key))
}

func AppendCHExpr(b []byte, key string) []byte {
	return chschema.AppendQuery(b, "d.string_values[indexOf(d.string_keys, ?)]", key)
}

func chGrouping(elem ast.GroupingElem, attr ch.Safe) (ch.Safe, error) {
	b, err := appendCHGrouping(nil, elem, attr)
	if err != nil {
		return "", err
	}
	return ch.Safe(b), nil
}

func appendCHGrouping(b []byte, elem ast.GroupingElem, attr ch.Safe) ([]byte, error) {
	switch elem.Func {
	case "":
		// nothing
	case "lower":
		b = append(b, "lowerUTF8("...)
	case "upper":
		b = append(b, "upperUTF8("...)
	default:
		return nil, fmt.Errorf("unsupported grouping func: %q", elem.Func)
	}

	b = append(b, attr...)

	if elem.Func != "" {
		b = append(b, ')')
	}

	return b, nil
}
