package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/exponential/mapping"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/exponential/mapping/logarithm"
	"golang.org/x/exp/slices"

	promlabels "github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/uptrace/go-clickhouse/ch/chschema"
)

type promQueryable struct {
	ctx       context.Context
	app       *bunapp.App
	db        *ch.DB
	projectID uint32
	logger    *otelzap.Logger
}

func newPromQueryable(ctx context.Context, app *bunapp.App, projectID uint32) *promQueryable {
	return &promQueryable{
		ctx:       ctx,
		app:       app,
		db:        app.CH(),
		projectID: projectID,
		logger:    app.ZapLogger(),
	}
}

var _ storage.Queryable = (*promQueryable)(nil)

func (q *promQueryable) Querier(
	ctx context.Context, timeGTE, timeLE int64,
) (storage.Querier, error) {
	return &promQuerier{
		promQueryable: q,
	}, nil
}

func (q *promQueryable) querier() *promQuerier {
	return &promQuerier{
		promQueryable: q,
	}
}

//------------------------------------------------------------------------------

type promQuerier struct {
	*promQueryable
}

var _ storage.Querier = (*promQuerier)(nil)

func (q *promQuerier) Select(
	sortSeries bool,
	hints *storage.SelectHints,
	matchers ...*promlabels.Matcher,
) storage.SeriesSet {
	chQuery := q.db.NewSelect().
		ColumnExpr("metric, attrs_hash, time").
		ColumnExpr(
			"multiIf("+
				"instrument = 'additive', argMax(value, time), "+
				"instrument = 'counter', sumWithOverflow(sum), "+
				"instrument = 'gauge', avg(value), "+
				"-1) AS value",
		).
		ColumnExpr("if("+
			"instrument = 'histogram', "+
			"argMax(histogram, time), "+
			"defaultValueOfTypeName('AggregateFunction(quantilesBFloat16(0.5, 0.9, 0.99), Float32)')) "+
			"AS histogram").
		ColumnExpr("anyLast(keys) AS keys").
		ColumnExpr("anyLast(values) AS values").
		TableExpr("?", q.app.DistTable("measure_minutes")).
		Where("project_id = ?", q.projectID).
		Where("time >= toDateTime(?)", hints.Start/1000).
		Where("time < toDateTime(?)", hints.End/1000).
		GroupExpr("metric, attrs_hash, time, instrument").
		OrderExpr("metric, attrs_hash, time").
		Limit(1e6)
	chQuery = transpileLabelMatchers(chQuery, matchers)

	rows, err := q.db.QueryContext(q.ctx, chQuery.String())
	if err != nil {
		return &promSeriesSet{err: err}
	}

	seriesSet := new(promSeriesSet)

	series := new(promSeries)
	var promHist *promHistogram
	for rows.Next() {
		var metric string
		var attrsHash uint32
		var tm time.Time
		var value float32
		var hist bfloat16.Map
		var keys []string
		var values []string

		if err := rows.Scan(&metric, &attrsHash, &tm, &value, &hist, &keys, &values); err != nil {
			return &promSeriesSet{err: err}
		}

		if series.metric != metric || series.attrsHash != attrsHash {
			if promHist != nil {
				for _, series := range promHist.buckets {
					if series != nil {
						seriesSet.slice = append(seriesSet.slice, series)
					}
				}
				promHist = nil
			}

			labels := make(promlabels.Labels, 0, 1+len(keys))

			labels = append(labels, promlabels.Label{
				Name:  "__name__",
				Value: metric,
			})

			for i, key := range keys {
				labels = append(labels, promlabels.Label{
					Name:  key,
					Value: fmt.Sprint(values[i]),
				})
			}

			series = &promSeries{
				metric:    metric,
				attrsHash: attrsHash,
				labels:    labels,
			}
			if len(hist) > 0 {
				promHist = newPromHistogram(func(le float64) *promSeries {
					series := series.Clone()
					series.labels = append(series.labels, promlabels.Label{
						Name:  "le",
						Value: fmt.Sprint(le),
					})
					seriesSet.slice = append(seriesSet.slice, series)
					return series
				})
			} else {
				seriesSet.slice = append(seriesSet.slice, series)
			}
		}

		if len(hist) > 0 {
			promHist.Add(hist, tm)
		} else {
			series.AddSample(float64(value), tm)
		}
	}

	spew.Dump(seriesSet)
	if err := rows.Err(); err != nil {
		return &promSeriesSet{err: err}
	}

	return seriesSet
}

func (q *promQuerier) Series(
	hints *storage.SelectHints,
	matchers ...*promlabels.Matcher,
) ([][][]string, error) {
	chQuery := q.db.NewSelect().
		DistinctOn("metric, attrs_hash").
		ColumnExpr("metric, keys, values").
		TableExpr("?", q.app.DistTable("measure_minutes")).
		Where("project_id = ?", q.projectID).
		Where("time >= toDateTime(?)", hints.Start/1000).
		Where("time < toDateTime(?)", hints.End/1000).
		OrderExpr("metric, attrs_hash").
		Limit(10000)
	chQuery = transpileLabelMatchers(chQuery, matchers)

	rows, err := q.db.QueryContext(q.ctx, chQuery.String())
	if err != nil {
		return nil, err
	}

	seriesLabels := make([][][]string, 0, 100)

	for rows.Next() {
		var metric string
		var keys []string
		var values []string

		if err := rows.Scan(&metric, &keys, &values); err != nil {
			return nil, err
		}

		labels := make([][]string, 0, 2*(1+len(keys)))
		labels = append(labels, []string{"__name__", metric})

		for i, key := range keys {
			labels = append(labels, []string{key, fmt.Sprint(values[i])})
		}

		seriesLabels = append(seriesLabels, labels)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return seriesLabels, nil
}

func (c *promQuerier) LabelValues(
	name string, matchers ...*promlabels.Matcher,
) ([]string, storage.Warnings, error) {
	return nil, nil, nil
}

func (q *promQuerier) LabelNames(
	matchers ...*promlabels.Matcher,
) ([]string, storage.Warnings, error) {
	return nil, nil, nil
}

func (q *promQuerier) Close() error {
	return nil
}

func transpileLabelMatchers(q *ch.SelectQuery, matchers []*promlabels.Matcher) *ch.SelectQuery {
	for _, m := range matchers {
		if m.Name == "le" {
			spew.Dump(m)
			continue
		}

		colName := chColumn(m.Name)

		switch m.Type {
		case promlabels.MatchEqual:
			q = q.Where("? = ?", colName, m.Value)
		case promlabels.MatchNotEqual:
			q = q.Where("? != ?", colName, m.Value)
		case promlabels.MatchRegexp:
			q = q.Where("match(?, ?)", colName, m.Value)
		case promlabels.MatchNotRegexp:
			q = q.Where("NOT match(%s, ?)", colName, m.Value)
		}
	}
	return q
}

//------------------------------------------------------------------------------

type promSeriesSet struct {
	slice []*promSeries
	index int
	err   error
}

var _ storage.SeriesSet = (*promSeriesSet)(nil)

func (e *promSeriesSet) Err() error {
	return e.err
}

func (e *promSeriesSet) Next() bool {
	e.index++
	return e.index <= len(e.slice)
}

func (e *promSeriesSet) At() storage.Series {
	if e.index >= 1 && e.index <= len(e.slice) {
		return e.slice[e.index-1]
	}
	return nil
}

func (e *promSeriesSet) Warnings() storage.Warnings {
	return nil
}

//------------------------------------------------------------------------------

type promSeries struct {
	metric    string
	attrsHash uint32
	labels    promlabels.Labels
	samples   []promSample
}

func (s *promSeries) Clone() *promSeries {
	clone := *s
	clone.labels = slices.Clone(clone.labels)
	clone.samples = slices.Clone(clone.samples)
	return &clone
}

func (s *promSeries) Labels() promlabels.Labels {
	return s.labels
}

func (s *promSeries) Iterator() chunkenc.Iterator {
	return &seriesIter{
		samples: s.samples,
	}
}

func (s *promSeries) AddSample(value float64, tm time.Time) {
	s.samples = append(s.samples, newPromSample(value, tm))
}

type promSample struct {
	value     float64
	timestamp int64
}

func newPromSample(value float64, tm time.Time) promSample {
	return promSample{
		value:     value,
		timestamp: tm.UnixNano() / int64(time.Millisecond),
	}
}

//------------------------------------------------------------------------------

type seriesIter struct {
	samples []promSample
	index   int
}

func (s *seriesIter) Next() bool {
	s.index++
	return s.index <= len(s.samples)
}

func (s *seriesIter) Seek(timestamp int64) bool {
	if len(s.samples) == 0 {
		return false
	}

	if timestamp <= s.samples[0].timestamp {
		s.index = 1
		return true
	}

	target := promSample{timestamp: timestamp}
	index, ok := slices.BinarySearchFunc(s.samples, target, func(a, b promSample) int {
		return int(a.timestamp - b.timestamp)
	})
	s.index = index + 1
	return ok
}

func (s *seriesIter) At() (int64, float64) {
	sample := s.samples[s.index-1]
	return sample.timestamp, sample.value
}

func (s *seriesIter) Err() error {
	return nil
}

//------------------------------------------------------------------------------

type promHistogram struct {
	newSeries func(le float64) *promSeries
	buckets   []*promSeries
	mapping   mapping.Mapping
}

func newPromHistogram(newSeries func(le float64) *promSeries) *promHistogram {
	mapping, err := logarithm.NewMapping(3)
	if err != nil {
		panic(err)
	}

	return &promHistogram{
		newSeries: newSeries,
		buckets:   make([]*promSeries, 10),
		mapping:   mapping,
	}
}

func (h *promHistogram) Bucket(value float64) *promSeries {
	index := int(h.mapping.MapToIndex(value))
	if index >= len(h.buckets) {
		h.buckets = slices.Grow(h.buckets, index+1)[:index+1]
	}

	series := h.buckets[index]
	if series == nil {
		series = h.newSeries(value)
		h.buckets[index] = series
	}

	return series
}

func (h *promHistogram) Add(hist bfloat16.Map, tm time.Time) {
	var sum uint64
	for value, count := range hist {
		sum += count
		h.Bucket(value.Float64()).AddSample(float64(sum), tm)
	}
}

//------------------------------------------------------------------------------

func chColumn(key string) ch.Safe {
	return ch.Safe(appendCHColumn(nil, key))
}

func appendCHColumn(b []byte, key string) []byte {
	if key == "__name__" {
		return chschema.AppendIdent(b, "metric")
	}
	return chschema.AppendQuery(b, "values[indexOf(keys, ?)]", key)
}
