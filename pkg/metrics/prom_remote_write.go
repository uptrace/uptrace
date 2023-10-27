package metrics

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
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
		metricName, err := promMetricName(ts.Labels)
		if err != nil {
			return err
		}

		attrs := make(AttrMap, len(ts.Labels))
		for _, l := range ts.Labels {
			if strings.HasPrefix(l.Name, "__") {
				continue
			}
			attrs[l.Name] = l.Value
		}

		isCumCounter, unit := promMetadata(metricName)
		for i := range ts.Samples {
			s := &ts.Samples[i]
			unixNano := uint64(s.Timestamp * int64(time.Millisecond))

			if isCumCounter {
				dp := p.nextDatapoint(metricName, InstrumentCounter, attrs, unixNano)
				dp.Unit = unit
				dp.CumPoint = &NumberPoint{
					Double: s.Value,
				}
				p.enqueue(ctx, dp)
				continue
			}

			dp := p.nextDatapoint(metricName, InstrumentGauge, attrs, unixNano)
			dp.Unit = unit
			dp.Gauge = float64(s.Value)
			p.enqueue(ctx, dp)
		}

		if len(ts.Histograms) > 0 {
			h.Zap(ctx).Error("histograms are unsupported")
		}
	}

	return nil
}

func promMetricName(labels []prompb.Label) (string, error) {
	const nameStr = "__name__"
	for _, label := range labels {
		if label.Name == nameStr {
			return attrkey.Clean(label.Value), nil
		}
	}
	return "", errors.New("prometheus: __name__ label not found")
}

func promMetadata(name string) (isCumCounter bool, unit string) {
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
