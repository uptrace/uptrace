package metrics

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/otlpconv"
	"github.com/uptrace/uptrace/pkg/tracing/xotel"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
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
					// ignore
				case *metricspb.Metric_ExponentialHistogram:
					// ignore
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
			dest.StartTimeUnix = uint32(dp.StartTimeUnixNano / uint64(time.Second))
			dest.NumberPoint = &NumberPoint{
				Int: value.AsInt,
			}
			p.enqueue(dest)
		case *metricspb.NumberDataPoint_AsDouble:
			dest.StartTimeUnix = uint32(dp.StartTimeUnixNano / uint64(time.Second))
			dest.NumberPoint = &NumberPoint{
				Double: value.AsDouble,
			}
			p.enqueue(dest)
		default:
			p.Zap(p.ctx).Error("unknown point value type",
				zap.String("type", fmt.Sprintf("%T", dp.Value)))
		}
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

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func cleanPromName(s string) string {
	if isValidPromName(s) {
		return s
	}

	r := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if isAllowedPromNameChar(c) {
			r = append(r, c)
		} else {
			r = append(r, '_')
		}
	}
	return unsafeconv.String(r)
}

func isValidPromName(s string) bool {
	for _, c := range []byte(s) {
		if !isAllowedPromNameChar(c) {
			return false
		}
	}
	return true
}

func isAllowedPromNameChar(c byte) bool {
	return isAlpha(c) || isDigit(c) || c == '_'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
