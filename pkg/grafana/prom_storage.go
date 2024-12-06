package grafana

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/org"
	"go.uber.org/zap"

	"github.com/prometheus/prometheus/model/histogram"
	promlabels "github.com/prometheus/prometheus/model/labels"
	promstorage "github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/prometheus/prometheus/util/annotations"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/chschema"
)

type PromStorage struct {
	logger    *otelzap.Logger
	ch        *ch.DB
	projectID uint32
}

func NewPromStorage(logger *otelzap.Logger, ch *ch.DB, projectID uint32) *PromStorage {
	return &PromStorage{
		logger:    logger,
		ch:        ch,
		projectID: projectID,
	}
}

var _ promstorage.Queryable = (*PromStorage)(nil)

func (s *PromStorage) Querier(
	timeGTE, timeLE int64,
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
	panic("not implemented")
}

//------------------------------------------------------------------------------

type promQuerier struct {
	*PromStorage
}

var _ promstorage.Querier = (*promQuerier)(nil)

func (pq *promQuerier) Select(
	ctx context.Context,
	sortSeries bool,
	hints *promstorage.SelectHints,
	matchers ...*promlabels.Matcher,
) promstorage.SeriesSet {
	tf := &org.TimeFilter{
		TimeGTE: time.Unix(hints.Start/1000, 0),
		TimeLT:  time.Unix(hints.End/1000, 0),
	}

	step := time.Duration(hints.Step) * time.Millisecond
	step = max(15*time.Second, step)

	var tableName string
	if step >= time.Hour {
		tableName = metrics.TableDatapointHours
		tf.Round(time.Hour)
	} else {
		tableName = metrics.TableDatapointMinutes
		tf.Round(time.Minute)
	}

	chQuery := pq.ch.NewSelect().
		ColumnExpr("d.metric, d.attrs_hash, d.instrument").
		ColumnExpr("toStartOfInterval(d.time, INTERVAL ? second) AS time_start", step.Seconds()).
		ColumnExpr(
			"multiIf("+
				"d.instrument = 'counter', sumWithOverflow(d.sum), "+
				"d.instrument = 'gauge', argMax(d.gauge, time), "+
				"-1) AS value",
		).
		TableExpr("? AS d", ch.Name(tableName)).
		Where("d.project_id = ?", pq.projectID).
		Where("d.time >= ?", tf.TimeGTE).
		Where("d.time <= ?", tf.TimeLT).
		GroupExpr("d.metric, d.attrs_hash, time_start, d.instrument").
		OrderExpr("d.metric, d.attrs_hash, time_start").
		Limit(100_000)

	chQuery.ColumnExpr("any(d.string_keys)").
		ColumnExpr("any(d.string_values)")

	if err := compilePromMatchers(chQuery, matchers); err != nil {
		return &promSeriesSet{err: err}
	}

	rows, err := chQuery.Query(ctx)
	if err != nil {
		return &promSeriesSet{err: err}
	}

	seriesSet := new(promSeriesSet)

	lastSeries := new(promSeries)
	var metric string
	var attrsHash uint64
	var instrument string
	var tm time.Time
	var value float64
	var keys []string
	var values []string
	for rows.Next() {
		if err := rows.Scan(
			&metric,
			&attrsHash,
			&instrument,
			&tm,
			&value,
			&keys,
			&values,
		); err != nil {
			return &promSeriesSet{err: err}
		}

		if lastSeries.metric != metric || lastSeries.attrsHash != attrsHash {
			if len(keys) != len(values) {
				pq.logger.Error("keys and values length does not match",
					zap.Strings("keys", keys),
					zap.Strings("values", values))
				continue
			}

			lastSeries = &promSeries{
				metric:    metric,
				attrsHash: attrsHash,
				labels:    makePromLabels(metric, keys, values),
			}
			seriesSet.slice = append(seriesSet.slice, lastSeries)
		}

		lastSeries.AddSample(value, tm)
	}

	if err := rows.Err(); err != nil {
		return &promSeriesSet{err: err}
	}

	return seriesSet
}

func (pq *promQuerier) Series(
	ctx context.Context,
	hints *promstorage.SelectHints,
	matchers ...*promlabels.Matcher,
) ([][][]string, error) {
	tf := &org.TimeFilter{
		TimeGTE: time.Unix(hints.Start/1000, 0),
		TimeLT:  time.Unix(hints.End/1000, 0),
	}

	tableName := metrics.DatapointTableForWhere(tf)
	chQuery := pq.ch.NewSelect().
		DistinctOn("d.metric, d.attrs_hash").
		ColumnExpr("d.metric").
		TableExpr("? AS d", ch.Name(tableName)).
		Where("d.project_id = ?", pq.projectID).
		Where("d.time >= ?", tf.TimeGTE).
		Where("d.time <= ?", tf.TimeLT).
		OrderExpr("d.metric, d.attrs_hash").
		Limit(10000)

	chQuery.ColumnExpr("d.string_keys").
		ColumnExpr("d.string_values")

	if err := compilePromMatchers(chQuery, matchers); err != nil {
		return nil, err
	}

	rows, err := chQuery.Query(ctx)
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

func (q *promQuerier) LabelNames(
	ctx context.Context, matchers ...*promlabels.Matcher,
) ([]string, annotations.Annotations, error) {
	return nil, nil, nil
}

func (c *promQuerier) LabelValues(
	ctx context.Context, name string, matchers ...*promlabels.Matcher,
) ([]string, annotations.Annotations, error) {
	return nil, nil, nil
}

func (q *promQuerier) Close() error {
	return nil
}

func compilePromMatchers(q *ch.SelectQuery, matchers []*promlabels.Matcher) error {
	for _, m := range matchers {
		if m.Name == promlabels.BucketLabel {
			return errors.New("Uptrace does not support 'le' Prometheus matcher")
		}
		if m.Value == "" {
			continue
		}

		chExpr := chExpr(m.Name)

		switch m.Type {
		case promlabels.MatchEqual:
			q.Where("? = ?", chExpr, m.Value)
		case promlabels.MatchNotEqual:
			q.Where("? != ?", chExpr, m.Value)
		case promlabels.MatchRegexp:
			q.Where("match(?, ?)", chExpr, m.Value)
		case promlabels.MatchNotRegexp:
			q.Where("NOT match(?, ?)", chExpr, m.Value)
		default:
			return fmt.Errorf("unsupported Prometheus matcher type: %q", m.Type)
		}
	}
	return nil
}

func chExpr(key string) ch.Safe {
	return ch.Safe(appendCHExpr(nil, key))
}

func appendCHExpr(b []byte, key string) []byte {
	switch key {
	case promlabels.MetricName:
		return chschema.AppendQuery(b, "d.metric")
	case "__project_id__":
		return chschema.AppendQuery(b, "d.project_id")
	default:
		return chschema.AppendQuery(b, "d.string_values[indexOf(d.string_keys, ?)]", key)
	}
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

func (e *promSeriesSet) Warnings() annotations.Annotations {
	return nil
}

//------------------------------------------------------------------------------

type promSeries struct {
	metric    string
	attrsHash uint64
	labels    promlabels.Labels
	samples   []promSample
}

type promSample struct {
	value     float64
	timestamp int64
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

func (s *promSeries) Iterator(chunkenc.Iterator) chunkenc.Iterator {
	return &promSeriesIter{
		samples: s.samples,
	}
}

func (s *promSeries) AddSample(value float64, tm time.Time) {
	s.samples = append(s.samples, promSample{
		value:     value,
		timestamp: tm.UnixMilli(),
	})
}

//------------------------------------------------------------------------------

type promSeriesIter struct {
	samples []promSample
	index   int
}

var _ chunkenc.Iterator = (*promSeriesIter)(nil)

func (s *promSeriesIter) Next() chunkenc.ValueType {
	s.index++
	if s.index <= len(s.samples) {
		return chunkenc.ValFloat
	}
	return chunkenc.ValNone
}

func (s *promSeriesIter) Seek(timestamp int64) chunkenc.ValueType {
	if len(s.samples) == 0 {
		return chunkenc.ValNone
	}

	if s.samples[0].timestamp >= timestamp {
		s.index = 1
		return chunkenc.ValFloat
	}

	target := promSample{timestamp: timestamp}
	index, _ := slices.BinarySearchFunc(s.samples, target, func(a, b promSample) int {
		return cmp.Compare(a.timestamp, b.timestamp)
	})
	if index < len(s.samples) {
		s.index = index + 1
		return chunkenc.ValFloat
	}
	s.index = 0
	return chunkenc.ValNone
}

func (s *promSeriesIter) AtT() int64 {
	t, _ := s.At()
	return t
}

func (s *promSeriesIter) At() (int64, float64) {
	sample := s.samples[s.index-1]
	return sample.timestamp, sample.value
}

func (s *promSeriesIter) AtFloatHistogram() (int64, *histogram.FloatHistogram) {
	return math.MinInt64, nil
}

func (s *promSeriesIter) AtHistogram() (int64, *histogram.Histogram) {
	return math.MinInt64, nil
}

func (s *promSeriesIter) Err() error {
	return nil
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

func makePromLabels(metric string, keys, values []string) []promlabels.Label {
	labels := make([]promlabels.Label, 0, len(keys)+1)
	labels = append(labels, promlabels.Label{
		Name:  promlabels.MetricName,
		Value: metric,
	})
	for i, key := range keys {
		labels = append(labels, promlabels.Label{
			Name:  key,
			Value: values[i],
		})
	}
	return labels
}
