package metrics

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"

	collectormetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/pkg/clickhouse/bfloat16"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/otlpconv"
)

const (
	cumulativeAggTemp = metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE
	deltaAggTemp      = metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA
)

type MetricsServiceServerParams struct {
	fx.In

	Logger   *otelzap.Logger
	PG       *bun.DB
	MP       *DatapointProcessor
	Projects *org.ProjectGateway
}

type MetricsServiceServer struct {
	*MetricsServiceServerParams
	collectormetricspb.UnimplementedMetricsServiceServer

	//mp *DatapointProcessor
}

func NewMetricsServiceServer(p MetricsServiceServerParams) *MetricsServiceServer {
	return &MetricsServiceServer{MetricsServiceServerParams: &p}
}

var _ collectormetricspb.MetricsServiceServer = (*MetricsServiceServer)(nil)

func (s *MetricsServiceServer) ExportHTTP(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn, err := org.DSNFromRequest(req)
	if err != nil {
		return err
	}

	project, err := s.Projects.SelectByDSN(ctx, dsn)
	if err != nil {
		return err
	}

	switch contentType := req.Header.Get("content-type"); contentType {
	case jsonContentType:
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}

		metricsReq := new(collectormetricspb.ExportMetricsServiceRequest)
		if err := protojson.Unmarshal(body, metricsReq); err != nil {
			return err
		}

		resp, err := s.process(ctx, metricsReq, project)
		if err != nil {
			return err
		}

		b, err := protojson.Marshal(resp)
		if err != nil {
			return err
		}

		if _, err := w.Write(b); err != nil {
			return err
		}

		return nil
	case xprotobufContentType, protobufContentType:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}

		metricsReq := new(collectormetricspb.ExportMetricsServiceRequest)
		if err := proto.Unmarshal(body, metricsReq); err != nil {
			return err
		}

		resp, err := s.process(ctx, metricsReq, project)
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

	dsn, err := org.DSNFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	project, err := s.Projects.SelectByDSN(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return s.process(ctx, req, project)
}

func (s *MetricsServiceServer) process(
	ctx context.Context,
	req *collectormetricspb.ExportMetricsServiceRequest,
	project *org.Project,
) (*collectormetricspb.ExportMetricsServiceResponse, error) {
	p := otlpProcessor{
		logger:  s.Logger,
		pg:      s.PG,
		mp:      s.MP,
		project: project,
	}
	defer p.close(ctx)

	for _, rms := range req.ResourceMetrics {
		var resource AttrMap
		if rms.Resource != nil {
			resource = make(AttrMap, len(rms.Resource.Attributes))
			otlpconv.ForEachKeyValue(rms.Resource.Attributes, func(key string, value any) {
				resource[key] = fmt.Sprint(value)
			})
		}

		for _, sm := range rms.ScopeMetrics {
			var scopeAttrs AttrMap
			if sm.Scope != nil && len(sm.Scope.Attributes) > 0 {
				scopeAttrs = make(AttrMap, len(resource)+len(sm.Scope.Attributes))
				maps.Copy(scopeAttrs, resource)
				otlpconv.ForEachKeyValue(sm.Scope.Attributes, func(key string, value any) {
					scopeAttrs[key] = fmt.Sprint(value)
				})
			} else {
				scopeAttrs = resource
			}

			if sm.Scope != nil && sm.Scope.Name != "" {
				if strings.Contains(sm.Scope.Name, "otelcol") {
					p.hasCollectorMetrics = true
				} else {
					p.hasAppMetrics = true
				}
			}

			for _, metric := range sm.Metrics {
				if metric == nil {
					continue
				}

				switch data := metric.Data.(type) {
				case *metricspb.Metric_Gauge:
					p.otlpGauge(ctx, sm.Scope, scopeAttrs, metric, data)
				case *metricspb.Metric_Sum:
					p.otlpSum(ctx, sm.Scope, scopeAttrs, metric, data)
				case *metricspb.Metric_Histogram:
					p.otlpHistogram(ctx, sm.Scope, scopeAttrs, metric, data)
				case *metricspb.Metric_ExponentialHistogram:
					p.otlpExpHistogram(ctx, sm.Scope, scopeAttrs, metric, data)
				case *metricspb.Metric_Summary:
					p.otlpSummary(ctx, sm.Scope, scopeAttrs, metric, data)
				default:
					p.logger.Error("unknown metric",
						zap.String("type", fmt.Sprintf("%T", data)))
				}
			}
		}
	}

	return &collectormetricspb.ExportMetricsServiceResponse{}, nil
}

type otlpProcessor struct {
	logger *otelzap.Logger
	pg     *bun.DB
	mp     *DatapointProcessor

	project     *org.Project
	metricIDMap map[MetricKey]struct{}

	hasCollectorMetrics bool
	hasAppMetrics       bool
}

func (p *otlpProcessor) close(ctx context.Context) {
	if p.hasCollectorMetrics {
		org.CreateAchievementOnce(ctx, p.logger, p.pg, &org.Achievement{
			ProjectID: p.project.ID,
			Name:      org.AchievInstallCollector,
		})
	}
	if p.hasAppMetrics {
		org.CreateAchievementOnce(ctx, p.logger, p.pg, &org.Achievement{
			ProjectID: p.project.ID,
			Name:      org.AchievConfigureMetrics,
		})
	}
}

func (p *otlpProcessor) otlpGauge(
	ctx context.Context,
	scope *commonpb.InstrumentationScope,
	scopeAttrs AttrMap,
	metric *metricspb.Metric,
	data *metricspb.Metric_Gauge,
) {
	for _, dp := range data.Gauge.DataPoints {
		if dp.Flags&uint32(metricspb.DataPointFlags_DATA_POINT_FLAGS_NO_RECORDED_VALUE_MASK) != 0 {
			continue
		}

		dest := p.otlpNewDatapoint(
			scope, scopeAttrs, metric, InstrumentGauge, dp.Attributes, dp.TimeUnixNano)
		switch num := dp.Value.(type) {
		case nil:
			dest.Gauge = 0
			p.enqueue(ctx, dest)
		case *metricspb.NumberDataPoint_AsInt:
			dest.Gauge = float64(num.AsInt)
			p.enqueue(ctx, dest)
		case *metricspb.NumberDataPoint_AsDouble:
			dest.Gauge = num.AsDouble
			p.enqueue(ctx, dest)
		default:
			p.logger.Error("unknown data point value",
				zap.String("type", fmt.Sprintf("%T", dp.Value)))
		}
	}
}

func (p *otlpProcessor) otlpSum(
	ctx context.Context,
	scope *commonpb.InstrumentationScope,
	scopeAttrs AttrMap,
	metric *metricspb.Metric,
	data *metricspb.Metric_Sum,
) {
	isDelta := data.Sum.AggregationTemporality == deltaAggTemp
	for _, dp := range data.Sum.DataPoints {
		if dp.Flags&uint32(metricspb.DataPointFlags_DATA_POINT_FLAGS_NO_RECORDED_VALUE_MASK) != 0 {
			continue
		}

		dest := p.otlpNewDatapoint(
			scope, scopeAttrs, metric, "", dp.Attributes, dp.TimeUnixNano)

		if !data.Sum.IsMonotonic {
			dest.Instrument = InstrumentAdditive
			dest.Gauge = toFloat64(dp.Value)
			p.enqueue(ctx, dest)
			continue
		}

		dest.Instrument = InstrumentCounter

		if isDelta {
			dest.Sum = toFloat64(dp.Value)
			p.enqueue(ctx, dest)
			continue
		}

		switch value := dp.Value.(type) {
		case *metricspb.NumberDataPoint_AsInt:
			dest.StartTimeUnixNano = dp.StartTimeUnixNano
			dest.CumPoint = &NumberPoint{
				Int: value.AsInt,
			}
			p.enqueue(ctx, dest)
		case *metricspb.NumberDataPoint_AsDouble:
			dest.StartTimeUnixNano = dp.StartTimeUnixNano
			dest.CumPoint = &NumberPoint{
				Double: value.AsDouble,
			}
			p.enqueue(ctx, dest)
		default:
			p.logger.Error("unknown point value type",
				zap.String("type", fmt.Sprintf("%T", dp.Value)))
		}
	}
}

func (p *otlpProcessor) otlpHistogram(
	ctx context.Context,
	scope *commonpb.InstrumentationScope,
	scopeAttrs AttrMap,
	metric *metricspb.Metric,
	data *metricspb.Metric_Histogram,
) {
	isDelta := data.Histogram.AggregationTemporality == deltaAggTemp
	for _, dp := range data.Histogram.DataPoints {
		if dp.Flags&uint32(metricspb.DataPointFlags_DATA_POINT_FLAGS_NO_RECORDED_VALUE_MASK) != 0 {
			continue
		}

		dest := p.otlpNewDatapoint(
			scope, scopeAttrs, metric, InstrumentHistogram, dp.Attributes, dp.TimeUnixNano)
		if isDelta {
			dest.Sum = dp.GetSum()
			dest.Count = dp.Count

			if dest.Count > 0 {
				avg := dest.Sum / float64(dest.Count)
				dest.Histogram, dest.Min, dest.Max = newBFloat16Histogram(
					dp.ExplicitBounds, dp.BucketCounts, avg)

				if dp.Min != nil {
					dest.Min = dp.GetMin()
				}
				if dp.Max != nil {
					dest.Max = dp.GetMax()
				}
			}
		} else {
			dest.StartTimeUnixNano = dp.StartTimeUnixNano
			dest.CumPoint = &HistogramPoint{
				Sum:          dp.GetSum(),
				Count:        dp.Count,
				Bounds:       dp.ExplicitBounds,
				BucketCounts: dp.BucketCounts,
			}
		}
		p.enqueue(ctx, dest)
	}
}

func (p *otlpProcessor) otlpExpHistogram(
	ctx context.Context,
	scope *commonpb.InstrumentationScope,
	scopeAttrs AttrMap,
	metric *metricspb.Metric,
	data *metricspb.Metric_ExponentialHistogram,
) {
	isDelta := data.ExponentialHistogram.AggregationTemporality == deltaAggTemp
	for _, dp := range data.ExponentialHistogram.DataPoints {
		if dp.Flags&uint32(metricspb.DataPointFlags_DATA_POINT_FLAGS_NO_RECORDED_VALUE_MASK) != 0 {
			continue
		}

		size := 1 + len(dp.Positive.BucketCounts) + len(dp.Negative.BucketCounts)
		hist := make(map[bfloat16.T]uint64, size)

		if dp.ZeroCount > 0 {
			hist[bfloat16.From(0)] += dp.ZeroCount
		}
		base := math.Pow(2, math.Pow(2, float64(dp.Scale)))
		buildBFloat16Hist(hist, base, int(dp.Positive.Offset), dp.Positive.BucketCounts, +1)
		buildBFloat16Hist(hist, base, int(dp.Negative.Offset), dp.Negative.BucketCounts, -1)

		dest := p.otlpNewDatapoint(
			scope, scopeAttrs, metric, InstrumentHistogram, dp.Attributes, dp.TimeUnixNano)

		if isDelta {
			dest.Sum = dp.GetSum()
			dest.Count = dp.Count
			dest.Histogram = hist

			if dp.Min != nil && dp.Max != nil {
				dest.Min = dp.GetMin()
				dest.Max = dp.GetMax()
			} else {
				avg := dest.Sum / float64(dest.Count)
				min, max := avg, avg
				for key := range hist {
					mean := float64(key.Float32())
					if mean < min {
						min = mean
					}
					if mean > max {
						max = mean
					}
				}
				dest.Min = min
				dest.Max = max
			}
		} else {
			dest.StartTimeUnixNano = dp.StartTimeUnixNano
			dest.CumPoint = &ExpHistogramPoint{
				Sum:       dp.GetSum(),
				Count:     dp.Count,
				Histogram: hist,
			}
		}

		p.enqueue(ctx, dest)
	}
}

func buildBFloat16Hist(
	hist map[bfloat16.T]uint64, base float64, offset int, counts []uint64, sign float64,
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

func (p *otlpProcessor) otlpSummary(
	ctx context.Context,
	scope *commonpb.InstrumentationScope,
	scopeAttrs AttrMap,
	metric *metricspb.Metric,
	data *metricspb.Metric_Summary,
) {
	for _, dp := range data.Summary.DataPoints {
		if dp.Flags&uint32(metricspb.DataPointFlags_DATA_POINT_FLAGS_NO_RECORDED_VALUE_MASK) != 0 {
			continue
		}

		avg := dp.Sum / float64(dp.Count)
		min, max := avg, avg
		for _, qv := range dp.QuantileValues {
			if qv.Value < min {
				min = qv.Value
			}
			if qv.Value > max {
				max = qv.Value
			}
		}

		dest := p.otlpNewDatapoint(
			scope, scopeAttrs, metric, InstrumentSummary, dp.Attributes, dp.TimeUnixNano)
		dest.Min = min
		dest.Max = max
		dest.Sum = dp.Sum
		dest.Count = dp.Count

		p.enqueue(ctx, dest)
	}
}

func (p *otlpProcessor) otlpNewDatapoint(
	scope *commonpb.InstrumentationScope,
	scopeAttrs AttrMap,
	metric *metricspb.Metric,
	instrument Instrument,
	labels []*commonpb.KeyValue,
	unixNano uint64,
) *Datapoint {
	attrs := make(AttrMap, len(scopeAttrs)+len(labels))
	maps.Copy(attrs, scopeAttrs)
	otlpconv.ForEachKeyValue(labels, func(key string, value any) {
		attrs[key] = fmt.Sprint(value)
	})

	metricName := attrkey.Clean(metric.Name)
	dest := p.newDatapoint(metricName, instrument, attrs, unixNano)

	dest.Description = metric.Description
	dest.Unit = bunconv.NormUnit(metric.Unit)

	if scope != nil {
		dest.OtelLibraryName = scope.Name
		dest.OtelLibraryVersion = scope.Version
	}

	return dest
}

func (p *otlpProcessor) newDatapoint(
	metricName string,
	instrument Instrument,
	attrs AttrMap,
	unixNano uint64,
) *Datapoint {
	dest := new(Datapoint)
	dest.ProjectID = p.project.ID
	dest.Metric = metricName
	dest.Instrument = instrument
	dest.Time = time.Unix(0, int64(unixNano))
	dest.Attrs = attrs
	return dest
}

func (p *otlpProcessor) enqueue(ctx context.Context, datapoint *Datapoint) {
	if datapoint.ProjectID == 0 {
		p.logger.Error("project id is empty")
		return
	}
	if datapoint.Metric == "" {
		p.logger.Error("metric name is empty")
		return
	}
	if datapoint.Instrument == "" {
		p.logger.Error("instrument is empty")
		return
	}
	if datapoint.Time.IsZero() {
		p.logger.Error("time is empty")
		return
	}

	p.mp.AddDatapoint(ctx, datapoint)
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
	m   map[bfloat16.T]uint64
	min float64
	max float64
}

func (h *quickBFloat16Histogram) add(mean float64, count uint64) {
	h.m[bfloat16.From(mean)] += count
	if mean < h.min {
		h.min = mean
	}
	if mean > h.max {
		h.max = mean
	}
}

func newBFloat16Histogram(
	bounds []float64, counts []uint64, avg float64,
) (map[bfloat16.T]uint64, float64, float64) {
	if len(bounds) == 0 {
		return nil, 0, 0
	}
	if len(counts)-1 != len(bounds) {
		return nil, 0, 0
	}

	h := quickBFloat16Histogram{
		m:   make(map[bfloat16.T]uint64, len(counts)+1),
		min: avg,
		max: avg,
	}

	if firstCount := counts[0]; firstCount > 0 {
		mean := bounds[0]
		h.add(mean, firstCount)
	}
	counts = counts[1:]

	prev := bounds[0]
	for i, count := range counts[:len(counts)-1] {
		upper := bounds[i+1]
		if count > 0 {
			mean := (upper + prev) / 2
			h.add(mean, count)
		}
		prev = upper
	}

	if lastCount := counts[len(counts)-1]; lastCount > 0 {
		mean := math.Nextafter(bounds[len(bounds)-1], math.MaxFloat64)
		h.add(mean, lastCount)
	}

	if len(h.m) == 0 {
		h.m = nil
	}
	return h.m, h.min, h.max
}
