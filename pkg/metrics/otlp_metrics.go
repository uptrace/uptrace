package metrics

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/otlpconv"
	"github.com/uptrace/uptrace/pkg/tracing/xotel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	collectormetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	cumulativeAggTemp = metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE
	deltaAggTemp      = metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA
)

type MetricsServiceServer struct {
	collectormetricspb.UnimplementedMetricsServiceServer

	*bunapp.App

	mp *MeasureProcessor
}

func NewMetricsServiceServer(app *bunapp.App) *MetricsServiceServer {
	return &MetricsServiceServer{
		App: app,
		mp:  NewMeasureProcessor(app),
	}
}

var _ collectormetricspb.MetricsServiceServer = (*MetricsServiceServer)(nil)

func (s *MetricsServiceServer) ExportHTTP(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn := req.Header.Get("uptrace-dsn")
	if dsn == "" {
		return errors.New("uptrace-dsn header is empty or missing")
	}

	project, err := org.SelectProjectByDSN(ctx, s.App, dsn)
	if err != nil {
		return err
	}

	switch contentType := req.Header.Get("content-type"); contentType {
	case jsonContentType:
		metricsReq := new(collectormetricspb.ExportMetricsServiceRequest)

		if err := jsonpb.Unmarshal(req.Body, metricsReq); err != nil {
			return err
		}

		resp, err := s.export(ctx, metricsReq, project)
		if err != nil {
			return err
		}

		return jsonMarshaler.Marshal(w, resp)
	case protobufContentType:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}

		metricsReq := new(collectormetricspb.ExportMetricsServiceRequest)
		if err := proto.Unmarshal(body, metricsReq); err != nil {
			return err
		}

		resp, err := s.export(ctx, metricsReq, project)
		if err != nil {
			return err
		}

		b, err := proto.Marshal(resp)
		if err != nil {
			return err
		}

		if _, err := w.Write(b); err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("unsupported content type: %q", contentType)
	}
}

var _ collectormetricspb.MetricsServiceServer = (*MetricsServiceServer)(nil)

func (s *MetricsServiceServer) Export(
	ctx context.Context, req *collectormetricspb.ExportMetricsServiceRequest,
) (*collectormetricspb.ExportMetricsServiceResponse, error) {
	if ctx.Err() == context.Canceled {
		return nil, status.Error(codes.Canceled, "Client cancelled, abandoning.")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("metadata is empty")
	}

	dsn := md.Get("uptrace-dsn")
	if len(dsn) == 0 {
		return nil, errors.New("uptrace-dsn header is required")
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attribute.String("dsn", dsn[0]))
	}

	project, err := org.SelectProjectByDSN(ctx, s.App, dsn[0])
	if err != nil {
		return nil, err
	}

	return s.export(ctx, req, project)
}

func (s *MetricsServiceServer) export(
	ctx context.Context,
	req *collectormetricspb.ExportMetricsServiceRequest,
	project *bunapp.Project,
) (*collectormetricspb.ExportMetricsServiceResponse, error) {
	var numMetric int

	for _, rm := range req.ResourceMetrics {
		if len(rm.ScopeMetrics) == 0 {
			for _, ilm := range rm.InstrumentationLibraryMetrics {
				scopeMetrics := metricspb.ScopeMetrics{
					Scope: &commonpb.InstrumentationScope{
						Name:    ilm.InstrumentationLibrary.Name,
						Version: ilm.InstrumentationLibrary.Version,
					},
					Metrics:   ilm.Metrics,
					SchemaUrl: ilm.SchemaUrl,
				}
				rm.ScopeMetrics = append(rm.ScopeMetrics, &scopeMetrics)
			}
			rm.InstrumentationLibraryMetrics = nil
		}

		for _, sm := range rm.ScopeMetrics {
			numMetric += len(sm.Metrics)
		}
	}

	p := otlpProcessor{
		App: s.App,

		mp: s.mp,

		ctx:         ctx,
		project:     project,
		projectAttr: attribute.Int64("project_id", int64(project.ID)),
	}

	for _, rms := range req.ResourceMetrics {
		p.resource = make(xotel.AttrMap, len(rms.Resource.Attributes))
		otlpconv.ForEachAttr(rms.Resource.Attributes, func(key string, value any) {
			p.resource[cleanPromName(key)] = value
		})

		for _, sm := range rms.ScopeMetrics {
			for _, metric := range sm.Metrics {
				if metric == nil {
					continue
				}

				switch data := metric.Data.(type) {
				case *metricspb.Metric_Gauge:
					p.otlpGauge(sm.Scope, metric, data)
				case *metricspb.Metric_Sum:
					p.otlpSum(sm.Scope, metric, data)
				case *metricspb.Metric_Histogram:
					p.otlpHistogram(sm.Scope, metric, data)
				case *metricspb.Metric_ExponentialHistogram:
					p.otlpExpHistogram(sm.Scope, metric, data)
				default:
					p.Zap(p.ctx).Error("unknown metric",
						zap.String("type", fmt.Sprintf("%T", data)))
				}
			}
		}
	}

	return &collectormetricspb.ExportMetricsServiceResponse{}, nil
}

type otlpProcessor struct {
	*bunapp.App

	mp *MeasureProcessor

	ctx         context.Context
	project     *bunapp.Project
	projectAttr attribute.KeyValue
	resource    xotel.AttrMap
}

func (p *otlpProcessor) otlpGauge(
	scope *commonpb.InstrumentationScope,
	metric *metricspb.Metric,
	data *metricspb.Metric_Gauge,
) {
	for _, dp := range data.Gauge.DataPoints {
		if dp.Flags&uint32(metricspb.DataPointFlags_FLAG_NO_RECORDED_VALUE) != 0 {
			continue
		}

		dest := p.nextMeasure(scope, metric, GaugeInstrument, dp.Attributes, dp.TimeUnixNano)
		switch num := dp.Value.(type) {
		case *metricspb.NumberDataPoint_AsInt:
			dest.Value = float32(num.AsInt)
			p.enqueue(dest)
		case *metricspb.NumberDataPoint_AsDouble:
			dest.Value = float32(num.AsDouble)
			p.enqueue(dest)
		default:
			p.Zap(p.ctx).Error("unknown data point value",
				zap.String("type", fmt.Sprintf("%T", dp.Value)))
		}
	}
}

func (p *otlpProcessor) otlpSum(
	scope *commonpb.InstrumentationScope,
	metric *metricspb.Metric,
	data *metricspb.Metric_Sum,
) {
	isDelta := data.Sum.AggregationTemporality == deltaAggTemp
	for _, dp := range data.Sum.DataPoints {
		if dp.Flags&uint32(metricspb.DataPointFlags_FLAG_NO_RECORDED_VALUE) != 0 {
			continue
		}

		dest := p.nextMeasure(scope, metric, "", dp.Attributes, dp.TimeUnixNano)
		if !data.Sum.IsMonotonic {
			// Agg temporality does not matter.
			dest.Instrument = AdditiveInstrument
			dest.Value = float32(toFloat64(dp.Value))
			p.enqueue(dest)
			continue
		}

		dest.Instrument = CounterInstrument

		if isDelta {
			dest.Sum = float32(toFloat64(dp.Value))
			p.enqueue(dest)
			continue
		}

		switch value := dp.Value.(type) {
		case *metricspb.NumberDataPoint_AsInt:
			dest.StartTimeUnix = dp.StartTimeUnixNano
			dest.CumPoint = &NumberPoint{
				Int: value.AsInt,
			}
			p.enqueue(dest)
		case *metricspb.NumberDataPoint_AsDouble:
			dest.StartTimeUnix = dp.StartTimeUnixNano
			dest.CumPoint = &NumberPoint{
				Double: value.AsDouble,
			}
			p.enqueue(dest)
		default:
			p.Zap(p.ctx).Error("unknown point value type",
				zap.String("type", fmt.Sprintf("%T", dp.Value)))
		}
	}
}

func (p *otlpProcessor) otlpHistogram(
	scope *commonpb.InstrumentationScope,
	metric *metricspb.Metric,
	data *metricspb.Metric_Histogram,
) {
	isDelta := data.Histogram.AggregationTemporality == deltaAggTemp
	for _, dp := range data.Histogram.DataPoints {
		if dp.Flags&uint32(metricspb.DataPointFlags_FLAG_NO_RECORDED_VALUE) != 0 {
			continue
		}

		dest := p.nextMeasure(scope, metric, HistogramInstrument, dp.Attributes, dp.TimeUnixNano)
		if isDelta {
			dest.Histogram = newBFloat16Histogram(dp.ExplicitBounds, dp.BucketCounts)
		} else {
			dest.StartTimeUnix = dp.StartTimeUnixNano
			dest.CumPoint = &HistogramPoint{
				Sum:          dp.GetSum(),
				Count:        dp.Count,
				Bounds:       dp.ExplicitBounds,
				BucketCounts: dp.BucketCounts,
			}
		}
		p.enqueue(dest)
	}
}

func (p *otlpProcessor) otlpExpHistogram(
	scope *commonpb.InstrumentationScope,
	metric *metricspb.Metric,
	data *metricspb.Metric_ExponentialHistogram,
) {
	isDelta := data.ExponentialHistogram.AggregationTemporality == deltaAggTemp
	for _, dp := range data.ExponentialHistogram.DataPoints {
		if dp.Flags&uint32(metricspb.DataPointFlags_FLAG_NO_RECORDED_VALUE) != 0 {
			continue
		}

		dest := p.nextMeasure(scope, metric, HistogramInstrument, dp.Attributes, dp.TimeUnixNano)
		if isDelta {
			hist := make(bfloat16.Map)
			dest.Histogram = hist

			if dp.ZeroCount > 0 {
				hist[bfloat16.From(0)] += dp.ZeroCount
			}
			base := math.Pow(2, math.Pow(2, float64(dp.Scale)))
			populateBFloat16Hist(hist, base, int(dp.Positive.Offset), dp.Positive.BucketCounts, +1)
			populateBFloat16Hist(hist, base, int(dp.Negative.Offset), dp.Negative.BucketCounts, -1)
		} else {
			dest.StartTimeUnix = dp.StartTimeUnixNano
			dest.CumPoint = &ExpHistogramPoint{
				Sum:       dp.GetSum(),
				Count:     dp.Count,
				Scale:     dp.Scale,
				ZeroCount: dp.ZeroCount,
				Positive: ExpHistogramBuckets{
					Offset:       dp.Positive.Offset,
					BucketCounts: dp.Positive.BucketCounts,
				},
				Negative: ExpHistogramBuckets{
					Offset:       dp.Negative.Offset,
					BucketCounts: dp.Negative.BucketCounts,
				},
			}
		}
		p.enqueue(dest)
	}
}

func (p *otlpProcessor) nextMeasure(
	scope *commonpb.InstrumentationScope,
	metric *metricspb.Metric,
	instrument string,
	labels []*commonpb.KeyValue,
	unixNano uint64,
) *Measure {
	attrs := make(xotel.AttrMap, len(p.resource)+len(labels))
	attrs.Merge(p.resource)
	otlpconv.ForEachAttr(labels, func(key string, value any) {
		attrs[cleanPromName(key)] = value
	})

	out := new(Measure)

	out.ProjectID = p.project.ID
	// enqueue will check whether metric name is empty.
	out.Metric = cleanPromName(metric.Name)
	out.Description = metric.Description
	out.Unit = metric.Unit
	out.Instrument = instrument
	out.Attrs = attrs
	out.Time = time.Unix(0, int64(unixNano))

	// out.Attrs["otel_library_name"] = scope.Name
	// out.Attrs["otel_library_version"] = scope.Version

	return out
}

func (p *otlpProcessor) enqueue(measure *Measure) {
	if measure.ProjectID == 0 {
		p.Zap(p.ctx).Error("project id is empty")
		return
	}
	if measure.Metric == "" {
		p.Zap(p.ctx).Error("metric name is empty")
		return
	}
	if measure.Instrument == "" {
		p.Zap(p.ctx).Error("instrument is empty")
		return
	}
	if measure.Time.IsZero() {
		p.Zap(p.ctx).Error("time is empty")
		return
	}

	p.mp.AddMeasure(measure)
}

//------------------------------------------------------------------------------

func toFloat64(value any) float64 {
	switch num := value.(type) {
	case *metricspb.NumberDataPoint_AsInt:
		return float64(num.AsInt)
	case *metricspb.NumberDataPoint_AsDouble:
		return num.AsDouble
	default:
		return 0
	}
}

//------------------------------------------------------------------------------

type quickBFloat16Histogram struct {
	m bfloat16.Map
}

func (h *quickBFloat16Histogram) Add(mean float64, count uint64) {
	h.m[bfloat16.From(mean)] += count
}

func newBFloat16Histogram(
	bounds []float64, counts []uint64,
) map[bfloat16.T]uint64 {
	h := quickBFloat16Histogram{
		m: make(map[bfloat16.T]uint64, len(counts)),
	}

	if c0 := counts[0]; c0 > 0 {
		h.Add(bounds[0], c0)
	}
	counts = counts[1:]

	prev := bounds[0]
	for i, count := range counts[:len(counts)-1] {
		upper := bounds[i+1]
		if count > 0 {
			h.Add((upper+prev)/2, count)
		}
		prev = upper
	}

	if lastCount := counts[len(counts)-1]; lastCount > 0 {
		max := math.Nextafter(bounds[len(bounds)-1], math.MaxFloat64)
		h.Add(max, lastCount)
	}

	if len(h.m) > 0 {
		return h.m
	}
	return nil
}

func populateBFloat16Hist(
	hist bfloat16.Map, base float64, offset int, counts []uint64, sign float64,
) {
	lower := math.Pow(base, float64(offset))
	for i, count := range counts {
		if count == 0 {
			continue
		}
		upper := math.Pow(base, float64(offset+i+1))
		mean := (lower + upper) / 2
		hist[bfloat16.From(sign*mean)] += count
		lower = upper
	}
}
