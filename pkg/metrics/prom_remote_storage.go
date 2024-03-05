package metrics

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	promlabels "github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
)

type PrometheusHandler struct {
	*bunapp.App

	mp *DatapointProcessor
}

func NewPrometheusHandler(app *bunapp.App, mp *DatapointProcessor) *PrometheusHandler {
	return &PrometheusHandler{
		App: app,

		mp: mp,
	}
}

func (h *PrometheusHandler) Write(
	w http.ResponseWriter, req bunrouter.Request,
) error {
	ctx := req.Context()

	dsn, err := org.DSNFromRequest(req)
	if err != nil {
		return err
	}

	project, err := org.SelectProjectByDSN(ctx, h.App, dsn)
	if err != nil {
		return err
	}

	promReq, err := remote.DecodeWriteRequest(req.Body)
	if err != nil {
		return err
	}

	if err := h.handleTimeseries(ctx, project, promReq.Timeseries); err != nil {
		return err
	}

	w.WriteHeader(http.StatusAccepted)
	return nil
}

func (h *PrometheusHandler) handleTimeseries(
	ctx context.Context, project *org.Project, tss []prompb.TimeSeries,
) error {
	p := otlpProcessor{
		App:     h.App,
		mp:      h.mp,
		project: project,
	}
	defer p.close(ctx)

	for _, ts := range tss {
		metricName, isCumCounter, unit := promMetadata(ts.Labels)
		if metricName == "" {
			continue
		}
		if project.PromCompat {
			isCumCounter = false
		}

		attrs := make(AttrMap, len(ts.Labels))
		for _, l := range ts.Labels {
			if strings.HasPrefix(l.Name, "__") {
				continue
			}
			attrs[l.Name] = l.Value
		}

		for i := range ts.Samples {
			s := &ts.Samples[i]
			unixNano := uint64(s.Timestamp * int64(time.Millisecond))

			if isCumCounter {
				dp := p.newDatapoint(metricName, InstrumentCounter, attrs, unixNano)
				dp.Unit = unit
				dp.CumPoint = &NumberPoint{
					Double: s.Value,
				}
				p.enqueue(ctx, dp)
				continue
			}

			dp := p.newDatapoint(metricName, InstrumentGauge, attrs, unixNano)
			dp.Unit = unit
			dp.Gauge = float64(s.Value)
			dp.OtelLibraryName = "Prometheus"
			p.enqueue(ctx, dp)
		}

		if len(ts.Histograms) > 0 {
			h.Zap(ctx).Error("histograms are unsupported")
		}
	}

	return nil
}

func promMetadata(labels []prompb.Label) (metric string, isCumCounter bool, unit string) {
	const nameStr = promlabels.MetricName
	for _, label := range labels {
		if label.Name == nameStr {
			isCumCounter, unit := _promMetadata(label.Value)
			return label.Value, isCumCounter, unit
		}
	}
	return "", false, ""
}

func _promMetadata(name string) (isCumCounter bool, unit string) {
	for {
		word, after, hasNext := strings.Cut(name, "_")

		switch word {
		case "sum", "count", "total":
			isCumCounter = true
		case "milliseconds", "seconds", "bytes":
			unit = word
		}

		if !hasNext {
			return
		}
		name = after
	}
}

//------------------------------------------------------------------------------

const minGroupingInterval = time.Minute

func (h *PrometheusHandler) Read(
	w http.ResponseWriter, req bunrouter.Request,
) error {
	ctx := req.Context()

	dsn, err := org.DSNFromRequest(req)
	if err != nil {
		return err
	}

	project, err := org.SelectProjectByDSN(ctx, h.App, dsn)
	if err != nil {
		return err
	}

	promReq, err := remote.DecodeReadRequest(req.Request)
	if err != nil {
		return err
	}

	promResp, err := h.handleReadRequest(ctx, promReq, project)
	if err != nil {
		return err
	}

	if err := remote.EncodeReadResponse(promResp, w); err != nil {
		return err
	}
	return nil
}

func (h *PrometheusHandler) handleReadRequest(
	ctx context.Context, promReq *prompb.ReadRequest, project *org.Project,
) (*prompb.ReadResponse, error) {
	promResp := &prompb.ReadResponse{
		Results: make([]*prompb.QueryResult, len(promReq.Queries)),
	}

	for i, query := range promReq.Queries {
		result, err := h.handleQuery(ctx, query)
		if err != nil {
			return nil, err
		}
		promResp.Results[i] = result
	}

	return promResp, nil
}

type PromDatapoint struct {
	Metric       string
	Value        float64
	StringKeys   []string
	StringValues []string
	Time         time.Time
}

func (h *PrometheusHandler) handleQuery(
	ctx context.Context, query *prompb.Query,
) (*prompb.QueryResult, error) {
	groupingInterval := promGroupingInterval(query.Hints)
	q := h.CH.NewSelect().
		ColumnExpr("m.metric").
		ColumnExpr("m.string_keys, m.string_values").
		ColumnExpr(
			"multiIf("+
				"instrument = 'additive', argMax(value, time), "+
				"instrument = 'counter', sumWithOverflow(sum), "+
				"instrument = 'gauge', avg(value), "+
				"-1) AS value",
		).
		ColumnExpr("toStartOfInterval(m.time, INTERVAL ? SECOND)", groupingInterval.Seconds())

	q.Where("m.time >= toDateTime(?)", query.StartTimestampMs/1000)
	if query.EndTimestampMs > 0 {
		q.Where("m.time < toDateTime(?)", query.EndTimestampMs/1000)
	}

	if err := compilePromMatchers(q, query.Matchers); err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx)
	if err != nil {
		return nil, err
	}

	result := new(prompb.QueryResult)

	var dp PromDatapoint
	var lastTs *prompb.TimeSeries
	var lastDp PromDatapoint
	for rows.Next() {
		if err := rows.Scan(
			&dp.Metric,
			&dp.StringKeys,
			&dp.StringValues,
			&dp.Value,
			&dp.Time,
		); err != nil {
			return nil, err
		}

		if lastTs == nil ||
			lastDp.Metric != dp.Metric ||
			!slices.Equal(lastDp.StringKeys, dp.StringKeys) ||
			!slices.Equal(lastDp.StringValues, dp.StringValues) {
			lastDp = dp

			lastTs = &prompb.TimeSeries{
				Labels: makePromLabels(dp.Metric, dp.StringKeys, dp.StringValues),
			}
			result.Timeseries = append(result.Timeseries, lastTs)
		}

		lastTs.Samples = append(lastTs.Samples, prompb.Sample{
			Value:     dp.Value,
			Timestamp: dp.Time.UnixMilli(),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func promGroupingInterval(hints *prompb.ReadHints) time.Duration {
	return time.Minute
}

func compilePromMatchers(q *ch.SelectQuery, matchers []*prompb.LabelMatcher) error {
	for _, matcher := range matchers {
		if err := compilePromMatcher(q, matcher); err != nil {
			return err
		}
	}
	return nil
}

func compilePromMatcher(q *ch.SelectQuery, matcher *prompb.LabelMatcher) error {
	var chExpr ch.Safe
	if matcher.Name == promlabels.MetricName {
		chExpr = ch.Safe("metric")
	} else {
		chExpr = CHExpr(matcher.Name)
	}

	switch matcher.Type {
	case prompb.LabelMatcher_EQ:
		q.Where("? = ?", chExpr, matcher.Value)
	case prompb.LabelMatcher_NEQ:
		q.Where("? != ?", chExpr, matcher.Value)
	case prompb.LabelMatcher_RE:
		q.Where("match(?, ?)", chExpr, "^"+matcher.Value+"$")
	case prompb.LabelMatcher_NRE:
		q.Where("NOT match(?, ?)", chExpr, "^"+matcher.Value+"$")
	default:
		return fmt.Errorf("unsupported prom matcher type: %s", matcher.Type)
	}

	return nil
}

func makePromLabels(metric string, keys, values []string) []prompb.Label {
	labels := make([]prompb.Label, 0, len(keys)+1)
	labels = append(labels, prompb.Label{
		Name:  promlabels.MetricName,
		Value: metric,
	})
	for i, key := range keys {
		labels = append(labels, prompb.Label{
			Name:  key,
			Value: values[i],
		})
	}
	return labels
}
