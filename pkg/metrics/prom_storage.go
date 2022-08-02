package metrics

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/exponential/mapping"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/exponential/mapping/exponent"
	"golang.org/x/exp/slices"

	"github.com/prometheus/prometheus/model/exemplar"
	promlabels "github.com/prometheus/prometheus/model/labels"
	promstorage "github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/uptrace/go-clickhouse/ch/chschema"
)

type PromStorage struct {
	*bunapp.App

	ctx       context.Context
	projectID uint32
	logger    *otelzap.Logger
}

func NewPromStorage(
	ctx context.Context, app *bunapp.App, projectID uint32,
) *PromStorage {
	return &PromStorage{
		App:       app,
		ctx:       ctx,
		projectID: projectID,
		logger:    app.Logger(),
	}
}

var _ promstorage.Queryable = (*PromStorage)(nil)

func (s *PromStorage) Querier(
	ctx context.Context, timeGTE, timeLE int64,
) (promstorage.Querier, error) {
	return &promQuerier{
		PromStorage: s,
	}, nil
}

func (s *PromStorage) querier() *promQuerier {
	return &promQuerier{
		PromStorage: s,
	}
}

var _ promstorage.Appendable = (*PromStorage)(nil)

func (s *PromStorage) Appender(ctx context.Context) promstorage.Appender {
	return &promAppender{
		PromStorage: s,
	}
}

//------------------------------------------------------------------------------

type promAppender struct {
	*PromStorage
}

func (a *promAppender) Append(
	ref promstorage.SeriesRef, l promlabels.Labels, t int64, v float64,
) (promstorage.SeriesRef, error) {
	return 0, nil
}

func (a *promAppender) Commit() error {
	return nil
}

func (a *promAppender) Rollback() error {
	return nil
}

func (a *promAppender) AppendExemplar(
	ref promstorage.SeriesRef, l promlabels.Labels, e exemplar.Exemplar,
) (promstorage.SeriesRef, error) {
	return 0, nil
}

//------------------------------------------------------------------------------

type promQuerier struct {
	*PromStorage
}

var _ promstorage.Querier = (*promQuerier)(nil)

func (pq *promQuerier) Select(
	sortSeries bool,
	hints *promstorage.SelectHints,
	matchers ...*promlabels.Matcher,
) promstorage.SeriesSet {
	chQuery := pq.CH.NewSelect().
		ColumnExpr("metric, attrs_hash, time, instrument").
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
		TableExpr("?", pq.DistTable("measure_minutes")).
		Where("time >= toDateTime(?)", hints.Start/1000).
		Where("time < toDateTime(?)", hints.End/1000).
		GroupExpr("metric, attrs_hash, time, instrument").
		OrderExpr("metric, attrs_hash, time").
		Limit(1e6)

	if pq.projectID != 0 {
		chQuery = chQuery.Where("project_id = ?", pq.projectID)
	}

	chQuery = transpileLabelMatchers(chQuery, matchers)
	// TODO: use hints.Grouping

	rows, err := pq.CH.QueryContext(pq.ctx, chQuery.String())
	if err != nil {
		return &promSeriesSet{err: err}
	}

	seriesSet := new(promSeriesSet)

	series := new(promSeries)
	var promHist *promHistogram
	for rows.Next() {
		var metric string
		var attrsHash uint64
		var tm time.Time
		var instrument string
		var value float32
		var hist bfloat16.Map
		var keys []string
		var values []string

		if err := rows.Scan(
			&metric,
			&attrsHash,
			&tm,
			&instrument,
			&value,
			&hist,
			&keys,
			&values,
		); err != nil {
			return &promSeriesSet{err: err}
		}

		if series.metric != metric || series.attrsHash != attrsHash {
			labels := make(promlabels.Labels, 0, 1+len(keys))

			labels = append(labels, promlabels.Label{
				Name:  promlabels.MetricName,
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
			if instrument == HistogramInstrument {
				promHist = newPromHistogram(func(le float64) *promSeries {
					series := series.Clone()
					series.labels = append(series.labels, promlabels.Label{
						Name:  "le",
						Value: strconv.FormatFloat(roundPretty(le), 'f', -1, 64),
					})
					seriesSet.slice = append(seriesSet.slice, series)
					return series
				})
			} else {
				seriesSet.slice = append(seriesSet.slice, series)
			}
		}

		if instrument == HistogramInstrument {
			promHist.Add(hist, tm)
		} else {
			series.AddSample(float64(value), tm)
		}
	}

	if err := rows.Err(); err != nil {
		return &promSeriesSet{err: err}
	}

	return seriesSet
}

func (pq *promQuerier) Series(
	hints *promstorage.SelectHints,
	matchers ...*promlabels.Matcher,
) ([][][]string, error) {
	chQuery := pq.CH.NewSelect().
		DistinctOn("metric, attrs_hash").
		ColumnExpr("metric, keys, values").
		TableExpr("?", pq.DistTable("measure_minutes")).
		Where("time >= toDateTime(?)", hints.Start/1000).
		Where("time < toDateTime(?)", hints.End/1000).
		OrderExpr("metric, attrs_hash").
		Limit(10000)

	if pq.projectID != 0 {
		chQuery = chQuery.Where("project_id = ?", pq.projectID)
	}

	chQuery = transpileLabelMatchers(chQuery, matchers)
	// TODO: use hints.Grouping

	rows, err := pq.CH.QueryContext(pq.ctx, chQuery.String())
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

		labels := make([][]string, 0, 1+len(keys))
		labels = append(labels, []string{promlabels.MetricName, metric})

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
) ([]string, promstorage.Warnings, error) {
	return nil, nil, nil
}

func (q *promQuerier) LabelNames(
	matchers ...*promlabels.Matcher,
) ([]string, promstorage.Warnings, error) {
	return nil, nil, nil
}

func (q *promQuerier) Close() error {
	return nil
}

func transpileLabelMatchers(q *ch.SelectQuery, matchers []*promlabels.Matcher) *ch.SelectQuery {
	for _, m := range matchers {
		if m.Name == promlabels.BucketLabel {
			return q.Err(errors.New("Uptrace does not support 'le' Prometheus matcher"))
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
		default:
			return q.Err(fmt.Errorf("unknown Prometheus matcher type: %q", m.Type))
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

var _ promstorage.SeriesSet = (*promSeriesSet)(nil)

func (e *promSeriesSet) Err() error {
	return e.err
}

func (e *promSeriesSet) Next() bool {
	e.index++
	return e.index <= len(e.slice)
}

func (e *promSeriesSet) At() promstorage.Series {
	if e.index >= 1 && e.index <= len(e.slice) {
		return e.slice[e.index-1]
	}
	return nil
}

func (e *promSeriesSet) Warnings() promstorage.Warnings {
	return nil
}

//------------------------------------------------------------------------------

type promSeries struct {
	metric    string
	attrsHash uint64
	labels    promlabels.Labels
	samples   []promSample
}

func (s *promSeries) Clone() *promSeries {
	clone := *s
	clone.labels = slices.Clone(clone.labels)
	clone.samples = slices.Clone(clone.samples)
	return &clone
}

func (s *promSeries) UpdateLabel(name, value string) {
	for i := range s.labels {
		label := &s.labels[i]
		if label.Name == name {
			label.Value = value
			break
		}
	}
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
	newSeries   func(le float64) *promSeries
	seriesSlice []*promSeries
	infSeries   *promSeries

	mapping  mapping.Mapping
	buckets  []uint64
	minIndex int32
}

func newPromHistogram(newSeries func(le float64) *promSeries) *promHistogram {
	mapping, err := exponent.NewMapping(-1)
	if err != nil {
		panic(err)
	}

	h := &promHistogram{
		newSeries:   newSeries,
		seriesSlice: make([]*promSeries, 10),
		infSeries:   newSeries(math.Inf(+1)),
		mapping:     mapping,
	}
	h.initBuckets()

	return h
}

func (h *promHistogram) initBuckets() {
	maxIndex := h.mapping.MapToIndex(exponent.MaxValue)
	h.minIndex = h.mapping.MapToIndex(exponent.MinValue)
	h.buckets = make([]uint64, int(maxIndex-h.minIndex+1))
}

func (h *promHistogram) bucketIndex(value float64) int {
	return int(h.mapping.MapToIndex(value) - h.minIndex)
}

func (h *promHistogram) lowerBoundary(index int) float64 {
	number, err := h.mapping.LowerBoundary(int32(index) + h.minIndex)
	if err != nil {
		panic(err)
	}
	return number
}

func (h *promHistogram) Series(index int) *promSeries {
	if index >= len(h.seriesSlice) {
		h.seriesSlice = slices.Grow(h.seriesSlice, index+1)[:index+1]
	}

	series := h.seriesSlice[index]
	if series == nil {
		lower := h.lowerBoundary(index + 1)
		series = h.newSeries(lower)
		h.seriesSlice[index] = series
	}

	return series
}

func (h *promHistogram) Add(mp bfloat16.Map, tm time.Time) {
	for i := range h.buckets {
		h.buckets[i] = 0
	}

	var sum uint64

	for value, count := range mp {
		sum += count

		index := int(h.bucketIndex(value.Float64()))
		if index >= len(h.buckets) {
			h.buckets = slices.Grow(h.buckets, index+1)[:index+1]
		}
		h.buckets[index] += count
	}

	for index, count := range h.buckets {
		if count > 0 {
			h.Series(index).AddSample(float64(count), tm)
		}
	}

	h.infSeries.AddSample(float64(sum), tm)
}

func (h *promHistogram) ForEachSeries(fn func(series *promSeries)) {
	for _, series := range h.seriesSlice {
		if series != nil {
			fn(series)
		}
	}
	fn(h.infSeries)
}

//------------------------------------------------------------------------------

func buildLabelMap(
	hints *promstorage.SelectHints,
	matchers []*promlabels.Matcher,
) map[string]bool {
	labelMap := make(map[string]bool, len(hints.Grouping)+len(matchers))

	for _, key := range hints.Grouping {
		labelMap[key] = true
	}

	for _, matcher := range matchers {
		labelMap[matcher.Name] = true
	}

	return labelMap
}

func isEmptyMatcherSet(matchers []*promlabels.Matcher) bool {
	for _, lm := range matchers {
		if lm != nil && !lm.Matches("") {
			return false
		}
	}
	return true
}

func roundPretty(n float64) float64 {
	if n == 0 || math.IsNaN(n) || math.IsInf(n, 0) {
		return n
	}

	switch abs := math.Abs(n); {
	case abs < 0.001:
		return round(n, 4)
	case abs < 0.01:
		return round(n, 3)
	case abs < 0.1:
		return round(n, 2)
	case abs < 100:
		return round(n, 1)
	default:
		return round(n, 0)
	}
}

func round(f float64, mantissa int) float64 {
	pow := math.Pow(10, float64(mantissa))
	return math.Round(f*pow) / pow
}

func chColumn(key string) ch.Safe {
	return ch.Safe(appendCHColumn(nil, key))
}

func appendCHColumn(b []byte, key string) []byte {
	switch key {
	case promlabels.MetricName:
		return chschema.AppendIdent(b, "metric")
	case "__project_id__":
		return chschema.AppendIdent(b, "project_id")
	default:
		return chschema.AppendQuery(b, "values[indexOf(keys, ?)]", key)
	}
}
