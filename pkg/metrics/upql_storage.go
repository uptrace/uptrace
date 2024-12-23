package metrics

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"

	"github.com/uptrace/pkg/clickhouse/ch"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"github.com/uptrace/pkg/unixtime"
	"github.com/uptrace/pkg/unsafeconv"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/chquery"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
)

type CHStorage struct {
	ctx  context.Context
	conf *CHStorageConfig

	db *ch.DB
}

type CHStorageConfig struct {
	ProjectID uint32
	MetricMap map[string]*Metric
	Search    []chquery.Token
	TableName string
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
	selectAll := f.CHFunc == mql.CHAggNone
	subq, err := s.subquery(metric, f, selectAll)
	if err != nil {
		return nil, err
	}

	q := s.db.NewSelect().
		ColumnExpr("groupArray(toFloat64(d.value)) AS value").
		ColumnExpr("groupArray(d.time) AS time").
		TableExpr("(?) AS d", subq).
		Limit(10000)

	if selectAll {
		q = q.ColumnExpr("d.attrs_hash").GroupExpr("d.attrs_hash")
	}
	if len(f.Grouping) > 0 {
		for i := range f.Grouping {
			elem := &f.Grouping[i]
			q = q.Column(elem.Alias).Group(elem.Alias)
		}
	}

	return q, nil
}

func (s *CHStorage) subquery(
	metric *Metric,
	f *mql.TimeseriesFilter,
	selectAll bool,
) (_ *ch.SelectQuery, err error) {
	if strings.HasPrefix(metric.Name, "uptrace_tracing_") {
		return s.tracingSubquery(metric, f)
	}
	return s.metricsSubquery(metric, f, selectAll)
}

func (s *CHStorage) metricsSubquery(
	metric *Metric,
	f *mql.TimeseriesFilter,
	selectAll bool,
) (_ *ch.SelectQuery, err error) {
	q := s.db.NewSelect().
		TableExpr("? AS d", ch.Name(s.conf.TableName)).
		Where("d.project_id = ?", s.conf.ProjectID).
		Where("d.metric = ?", metric.Name).
		Where("d.time >= ?", f.TimeGTE).
		Where("d.time < ?", f.TimeLT)

	q = q.
		ColumnExpr("toUnixTimestamp(toStartOfInterval(d.time, INTERVAL ? SECOND)) AS time",
			f.Interval.Seconds()).
		GroupExpr("time").
		OrderExpr("time")

	if len(s.conf.Search) > 0 && len(f.Grouping) > 0 {
		for _, token := range s.conf.Search {
			switch token.ID {
			case chquery.INCLUDE_TOKEN:
				q = q.WhereGroup(" AND ", func(q *ch.SelectQuery) *ch.SelectQuery {
					for _, elem := range f.Grouping {
						chExpr := CHExpr(elem.Name)
						q = q.WhereOr("multiSearchAnyCaseInsensitiveUTF8(?, ?) != 0",
							chExpr, ch.Array(token.Values))
					}
					return q
				})
			case chquery.EXCLUDE_TOKEN:
				q = q.WhereGroup(" AND ", func(q *ch.SelectQuery) *ch.SelectQuery {
					for _, elem := range f.Grouping {
						chExpr := CHExpr(elem.Name)
						q = q.Where("NOT multiSearchAnyCaseInsensitiveUTF8(?, ?) != 0",
							chExpr, ch.Array(token.Values))
					}
					return q
				})
			case chquery.REGEXP_TOKEN:
				q = q.WhereGroup(" AND ", func(q *ch.SelectQuery) *ch.SelectQuery {
					for _, elem := range f.Grouping {
						chExpr := CHExpr(elem.Name)
						q = q.WhereOr("match(?, ?)", chExpr, token.Values[0])
					}
					return q
				})
			}
		}
	}

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

	if selectAll {
		q = q.ColumnExpr("d.attrs_hash").GroupExpr("d.attrs_hash")
	}
	for i := range f.Grouping {
		elem := &f.Grouping[i]
		chExpr, err := chGrouping(elem, CHExpr(elem.Name))
		if err != nil {
			return nil, err
		}
		q = q.ColumnExpr("? AS ?", chExpr, ch.Name(elem.Alias)).
			GroupExpr("?", chExpr)
	}

	shouldDedup := isValueInstrument(metric.Instrument)

	if shouldDedup {
		switch f.CHFunc {
		case mql.CHAggMin:
			q = q.ColumnExpr("min(d.gauge) AS value")
		case mql.CHAggMax:
			q = q.ColumnExpr("max(d.gauge) AS value")
		case mql.CHAggNone, mql.CHAggSum, mql.CHAggAvg, mql.CHAggMedian:
			q = q.ColumnExpr("argMax(d.gauge, d.time) AS value")
			if selectAll {
				break
			}
			q = q.GroupExpr("d.attrs_hash")
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
		default:
			err := fmt.Errorf("unsupported ClickHouse func during deduplication: %q", f.CHFunc)
			return nil, err
		}

		q = s.db.NewSelect().
			TableExpr("(?) AS d", q)
	}

	q, err = s.agg(q, metric, f)
	if err != nil {
		return nil, err
	}

	if shouldDedup {
		q = q.ColumnExpr("d.time").GroupExpr("d.time").OrderExpr("d.time ASC")

		if selectAll {
			q = q.ColumnExpr("d.attrs_hash").GroupExpr("d.attrs_hash")
		}

		for i := range f.Grouping {
			elem := &f.Grouping[i]
			q = q.Column(elem.Alias).Group(elem.Alias)
		}
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
					b = append(b, "d."...)
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
	case attrkey.SpanTime:
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
		case mql.CHAggNone, mql.CHAggSum:
			q = q.ColumnExpr("sumWithOverflow(d.sum) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.CHFunc)
		}

	case InstrumentGauge:
		switch f.CHFunc {
		case mql.CHAggMin:
			q = q.ColumnExpr("min(d.value) AS value")
			return q, nil
		case mql.CHAggMax:
			q = q.ColumnExpr("max(d.value) AS value")
			return q, nil
		case mql.CHAggSum: // may be okay
			q = q.ColumnExpr("sumWithOverflow(d.value) AS value")
			return q, nil
		case mql.CHAggNone, mql.CHAggAvg:
			q = q.ColumnExpr("avg(d.value) AS value")
			return q, nil
		case mql.CHAggMedian:
			q = q.ColumnExpr("median(d.value) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.CHFunc)
		}

	case InstrumentAdditive:
		switch f.CHFunc {
		case mql.CHAggMin:
			q = q.ColumnExpr("min(d.value) AS value")
			return q, nil
		case mql.CHAggMax:
			q = q.ColumnExpr("max(d.value) AS value")
			return q, nil
		case mql.CHAggNone, mql.CHAggSum:
			q = q.ColumnExpr("sumWithOverflow(d.value) AS value")
			return q, nil
		case mql.CHAggAvg:
			q = q.ColumnExpr("avg(d.value) AS value")
			return q, nil
		case mql.CHAggMedian:
			q = q.ColumnExpr("median(d.value) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.CHFunc)
		}

	case InstrumentSummary:
		switch f.CHFunc {
		case mql.CHAggCount:
			q = q.ColumnExpr("sumWithOverflow(d.count) AS value")
			return q, nil
		case mql.CHAggMin:
			q = q.ColumnExpr("min(d.min) AS value")
			return q, nil
		case mql.CHAggMax:
			q = q.ColumnExpr("max(d.max) AS value")
			return q, nil
		case mql.CHAggSum:
			q = q.ColumnExpr("sumWithOverflow(d.sum) AS value")
			return q, nil
		case mql.CHAggAvg:
			q = q.ColumnExpr("sumWithOverflow(d.sum) / sumWithOverflow(d.count) AS value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.CHFunc)
		}

	case InstrumentHistogram:
		switch f.CHFunc {
		case mql.CHAggMin:
			q = q.ColumnExpr("min(d.min) AS value")
			return q, nil
		case mql.CHAggMax:
			q = q.ColumnExpr("max(d.max) AS value")
			return q, nil
		case mql.CHAggAvg:
			q = q.ColumnExpr("sumWithOverflow(d.sum) / sumWithOverflow(d.count) AS value")
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

		timeCol := m["time"].([]uint32)
		ts.Time = *(*[]unixtime.Nano)(unsafe.Pointer(&timeCol))
		delete(m, "time")

		ts.Value = bunutil.FillUnixNum(
			ts.Value,
			ts.Time,
			math.NaN(),
			f.TimeGTE,
			f.TimeLT,
			f.Interval,
		)
		ts.Time = bunutil.FillUnixTime(
			ts.Time,
			f.TimeGTE,
			f.TimeLT,
			f.Interval,
		)

		if len(f.Grouping) > 0 {
			attrs := make([]string, 0, 2*len(f.Grouping))
			for _, elem := range f.Grouping {
				attrs = append(attrs, elem.Alias, fmt.Sprint(m[elem.Alias]))
				delete(m, elem.Alias)
			}
			ts.Attrs = mql.NewAttrs(attrs...)
		} else if hash, ok := m["attrs_hash"].(uint64); ok {
			ts.Attrs = mql.NewAttrs("hash", strconv.FormatUint(hash, 10))
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
	case attrkey.SpanTime:
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
	if funcName == "" || funcName == mql.CHAggNone {
		return fmt.Errorf("%s instrument requires a func", instrument)
	}
	return fmt.Errorf("%s instrument does not support %s", instrument, funcName)
}

func unsupportedAttrFunc(attrName, funcName string) error {
	if funcName == "" || funcName == mql.CHAggNone {
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

func chGrouping(elem *ast.GroupingElem, attr ch.Safe) (ch.Safe, error) {
	b, err := appendCHGrouping(nil, elem, attr)
	if err != nil {
		return "", err
	}
	return ch.Safe(b), nil
}

func appendCHGrouping(b []byte, elem *ast.GroupingElem, attr ch.Safe) ([]byte, error) {
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

//------------------------------------------------------------------------------

func (s *CHStorage) tracingSubquery(
	metric *Metric, f *mql.TimeseriesFilter,
) (*ch.SelectQuery, error) {
	if f.CHFunc == mql.CHAggNone {
		return nil, errors.New("tracing metric requires an agg func")
	}

	selq, err := s.spanQuery(metric, f)
	if err != nil {
		return nil, err
	}

	q := selq.
		ColumnExpr("toUnixTimestamp(toStartOfInterval(time, INTERVAL ? SECOND)) AS time",
			int64(f.Interval.Seconds())).
		GroupExpr("time").
		OrderExpr("time")
	return q, nil
}

func (s *CHStorage) spanQuery(
	metric *Metric,
	f *mql.TimeseriesFilter,
) (_ *ch.SelectQuery, err error) {
	var query []string

	query, err = s.tracingAgg(query, f, metric)
	if err != nil {
		return nil, err
	}

	if len(f.Filters) > 0 {
		query = append(query, s.tracingFilters(f.Filters))
	}
	for _, filters := range f.Where {
		query = append(query, s.tracingFilters(filters))
	}

	for _, elem := range f.Grouping {
		var b []byte

		b = append(b, "group by "...)
		b = elem.AppendString(b)

		query = append(query, string(b))
	}

	timeFilter := &org.TimeFilter{
		TimeGTE: f.TimeGTE,
		TimeLT:  f.TimeLT,
	}
	typeFilter, err := newTypeFilter(s.ctx, s.conf.ProjectID, timeFilter, metric.Name)
	if err != nil {
		return nil, err
	}

	spanFilter := newSpanFilter(typeFilter, mql.JoinQuery(query))
	spanFilter.SearchTokens = s.conf.Search

	for _, part := range spanFilter.QueryParts {
		if part.Error.Wrapped != nil {
			return nil, part.Error.Wrapped
		}
	}

	selq, _ := tracing.BuildSpanIndexQuery(s.db, spanFilter, f.Interval)
	return selq, nil
}

func (s *CHStorage) tracingAgg(
	q []string, f *mql.TimeseriesFilter, metric *Metric,
) ([]string, error) {
	switch metric.Instrument {
	case InstrumentDeleted:
		return nil, fmt.Errorf("metric %q not found", metric.Name)

	case InstrumentCounter:
		switch f.CHFunc {
		case "", "_", "sum", "per_min", "per_sec":
			q = append(q, "sum(_count) as value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.CHFunc)
		}

	case InstrumentHistogram:
		switch f.CHFunc {
		case mql.CHAggAvg:
			q = append(q, "avg(_duration) as value")
			return q, nil
		case mql.CHAggMin:
			q = append(q, "min(_duration) as value")
			return q, nil
		case mql.CHAggMax:
			q = append(q, "max(_duration) as value")
			return q, nil
		case mql.CHAggP50:
			q = append(q, "p50(_duration) as value")
			return q, nil
		case mql.CHAggP75:
			q = append(q, "p75(_duration) as value")
			return q, nil
		case mql.CHAggP90:
			q = append(q, "p90(_duration) as value")
			return q, nil
		case mql.CHAggP95:
			q = append(q, "p95(_duration) as value")
			return q, nil
		case mql.CHAggP99:
			q = append(q, "p99(_duration) as value")
			return q, nil
		case mql.CHAggCount:
			q = append(q, "sum(_count) as value")
			return q, nil
		default:
			return nil, unsupportedInstrumentFunc(metric.Instrument, f.CHFunc)
		}

	default:
		return nil, fmt.Errorf("unsupported instrument %q", metric.Instrument)
	}
}

func (s *CHStorage) tracingFilters(filters ast.Filters) string {
	var b []byte
	b = append(b, "where "...)
	b = filters.AppendString(b)
	return unsafeconv.String(b)
}

func newTypeFilter(
	ctx context.Context,
	projectID uint32,
	timeFilter *org.TimeFilter,
	metricName string,
) (*tracing.TypeFilter, error) {
	typeFilter := &tracing.TypeFilter{
		ProjectID:  projectID,
		TimeFilter: *timeFilter,
	}

	switch metricName {
	case uptraceTracingSpans:
		typeFilter.System = []string{tracing.SystemSpansAll}
	case uptraceTracingEvents:
		typeFilter.System = []string{tracing.SystemEventsAll}
	case uptraceTracingLogs:
		typeFilter.System = []string{tracing.SystemLogAll}
	default:
		return nil, fmt.Errorf("unsupported tracing metric: %s", metricName)
	}

	return typeFilter, nil
}

func newSpanFilter(typeFilter *tracing.TypeFilter, query string) *tracing.SpanFilter {
	spanFilter := &tracing.SpanFilter{
		TypeFilter: *typeFilter,
		Query:      query,
	}
	spanFilter.QueryParts = tql.ParseQuery(query)
	return spanFilter
}
