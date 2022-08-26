package metrics

import (
	"context"
	"math"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.uber.org/zap"
	"go4.org/syncutil"
	"golang.org/x/exp/slices"
)

type MeasureProcessor struct {
	*bunapp.App

	batchSize int
	dropAttrs map[string]struct{}

	ch   chan *Measure
	gate *syncutil.Gate

	c2d    *CumToDeltaConv
	logger *otelzap.Logger

	metricMapMu sync.RWMutex
	metricMap   map[MetricKey]struct{}

	dashSyncer *DashSyncer
}

func NewMeasureProcessor(app *bunapp.App) *MeasureProcessor {
	conf := app.Config()
	p := &MeasureProcessor{
		App: app,

		batchSize: conf.Metrics.BatchSize,

		ch:   make(chan *Measure, conf.Metrics.BufferSize),
		gate: syncutil.NewGate(runtime.GOMAXPROCS(0)),

		c2d:    NewCumToDeltaConv(bunconf.ScaleWithCPU(4000, 32000)),
		logger: app.Logger,

		metricMap:  make(map[MetricKey]struct{}),
		dashSyncer: NewDashSyncer(app),
	}

	if len(conf.Metrics.DropAttrs) > 0 {
		p.dropAttrs = listToSet(conf.Metrics.DropAttrs)
	}

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()

		p.processLoop(app.Context())
	}()

	bufferSize, _ := bunotel.Meter.AsyncInt64().Gauge("uptrace.metrics.buffer_size")

	if err := bunotel.Meter.RegisterCallback(
		[]instrument.Asynchronous{
			bufferSize,
		},
		func(ctx context.Context) {
			bufferSize.Observe(ctx, int64(len(p.ch)))
		},
	); err != nil {
		panic(err)
	}

	return p
}

func (p *MeasureProcessor) AddMeasure(ctx context.Context, measure *Measure) {
	select {
	case p.ch <- measure:
	default:
		p.logger.Error("measure buffer is full (consider increasing metrics.buffer_size)")
		measureCounter.Add(
			ctx,
			1,
			bunotel.ProjectIDAttr(measure.ProjectID),
			attribute.String("type", "dropped"),
		)
	}
}

func (p *MeasureProcessor) processLoop(ctx context.Context) {
	const timeout = 5 * time.Second

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	measures := make([]*Measure, 0, p.batchSize)

loop:
	for {
		select {
		case measure := <-p.ch:
			if !p.processMeasure(ctx, measure) {
				break
			}

			p.upsertMetric(ctx, measure)

			measures = append(measures, measure)
			measureCounter.Add(
				ctx,
				1,
				bunotel.ProjectIDAttr(measure.ProjectID),
				attribute.String("type", "inserted"),
			)

			if len(measures) == p.batchSize {
				p.flushMeasures(ctx, measures)
				measures = make([]*Measure, 0, len(measures))
			}
		case <-timer.C:
			if len(measures) > 0 {
				p.flushMeasures(ctx, measures)
				measures = make([]*Measure, 0, len(measures))
			}
			timer.Reset(timeout)
		case <-p.Done():
			break loop
		}
	}

	if len(measures) > 0 {
		p.flushMeasures(ctx, measures)
	}
}

func (s *MeasureProcessor) processMeasure(ctx context.Context, measure *Measure) bool {
	switch point := measure.CumPoint.(type) {
	case nil:
		return true
	case *NumberPoint:
		if !s.convertNumberPoint(ctx, measure, point) {
			return false
		}
		measure.CumPoint = nil
		return true
	case *HistogramPoint:
		if !s.convertHistogramPoint(ctx, measure, point) {
			return false
		}
		measure.CumPoint = nil
		return true
	case *ExpHistogramPoint:
		if !s.convertExpHistogramPoint(ctx, measure, point) {
			return false
		}
		measure.CumPoint = nil
		return true
	default:
		s.Zap(ctx).Error("unknown cum point type",
			zap.String("type", reflect.TypeOf(point).String()))
		return false
	}
}

func (s *MeasureProcessor) convertNumberPoint(
	ctx context.Context, measure *Measure, point *NumberPoint,
) bool {
	key := MeasureKey{
		Metric:            measure.Metric,
		AttrsHash:         measure.AttrsHash,
		StartTimeUnixNano: measure.StartTimeUnixNano,
	}

	prevPoint, ok := s.c2d.Lookup(key, point, measure.Time).(*NumberPoint)
	if !ok {
		return false
	}

	if delta := point.Int - prevPoint.Int; delta > 0 {
		measure.Sum = float64(delta)
	} else if delta := point.Double - prevPoint.Double; delta > 0 {
		measure.Sum = delta
	}
	return true
}

func (s *MeasureProcessor) convertHistogramPoint(
	ctx context.Context, measure *Measure, point *HistogramPoint,
) bool {
	key := MeasureKey{
		Metric:            measure.Metric,
		AttrsHash:         measure.AttrsHash,
		StartTimeUnixNano: measure.StartTimeUnixNano,
	}

	prevPoint, ok := s.c2d.Lookup(key, point, measure.Time).(*HistogramPoint)
	if !ok {
		return false
	}
	if len(point.BucketCounts) != len(prevPoint.BucketCounts) {
		s.Zap(ctx).Error("number of buckets does not match")
		return false
	}

	measure.Sum = max(0, point.Sum-prevPoint.Sum)

	for i, num := range point.BucketCounts {
		prevNum := prevPoint.BucketCounts[i]
		point.BucketCounts[i] = max(0, num-prevNum)
	}

	measure.Histogram = newBFloat16Histogram(point.Bounds, point.BucketCounts)
	return true
}

func (s *MeasureProcessor) convertExpHistogramPoint(
	ctx context.Context, measure *Measure, point *ExpHistogramPoint,
) bool {
	key := MeasureKey{
		Metric:            measure.Metric,
		AttrsHash:         measure.AttrsHash,
		StartTimeUnixNano: measure.StartTimeUnixNano,
	}

	prevPoint, ok := s.c2d.Lookup(key, point, measure.Time).(*ExpHistogramPoint)
	if !ok {
		return false
	}

	if point.Scale != prevPoint.Scale {
		s.Zap(ctx).Error("scale does not match")
		return false
	}
	if point.Positive.Offset != prevPoint.Positive.Offset {
		s.Zap(ctx).Error("positive offset does not match")
		return false
	}
	if len(point.Positive.BucketCounts) != len(prevPoint.Positive.BucketCounts) {
		s.Zap(ctx).Error("positive number of buckets does not match")
		return false
	}
	if point.Negative.Offset != prevPoint.Negative.Offset {
		s.Zap(ctx).Error("negative offset does not match")
		return false
	}
	if len(point.Negative.BucketCounts) != len(prevPoint.Negative.BucketCounts) {
		s.Zap(ctx).Error("negative number of buckets does not match")
		return false
	}

	measure.Sum = max(0, point.Sum-prevPoint.Sum)

	point.ZeroCount -= prevPoint.ZeroCount
	for i, count := range point.Positive.BucketCounts {
		point.Positive.BucketCounts[i] = count - prevPoint.Positive.BucketCounts[i]
	}
	for i, count := range point.Negative.BucketCounts {
		point.Negative.BucketCounts[i] = count - prevPoint.Negative.BucketCounts[i]
	}

	hist := make(bfloat16.Map)
	measure.Histogram = hist

	if point.ZeroCount > 0 {
		hist[bfloat16.From(0)] += point.ZeroCount
	}
	base := math.Pow(2, math.Pow(2, float64(point.Scale)))
	populateBFloat16Hist(hist, base, int(point.Positive.Offset), point.Positive.BucketCounts, +1)
	populateBFloat16Hist(hist, base, int(point.Negative.Offset), point.Negative.BucketCounts, -1)

	return true
}

func (s *MeasureProcessor) flushMeasures(ctx context.Context, measures []*Measure) {
	ctx, span := bunotel.Tracer.Start(ctx, "flush-measures")

	s.WaitGroup().Add(1)
	s.gate.Start()

	go func() {
		defer span.End()
		defer s.gate.Done()
		defer s.WaitGroup().Done()

		ctx := newMeasureContext(ctx)
		s._flushMeasures(ctx, measures)
	}()
}

func (p *MeasureProcessor) _flushMeasures(ctx *measureContext, measures []*Measure) {
	for _, m := range measures {
		p.initMeasure(ctx, m)
	}

	if err := InsertMeasures(ctx, p.App, measures); err != nil {
		p.Zap(ctx).Error("InsertMeasures failed", zap.Error(err))
	}
}

func (p *MeasureProcessor) initMeasure(ctx *measureContext, measure *Measure) {
	keys := make([]string, 0, len(measure.Attrs))
	values := make([]string, 0, len(measure.Attrs))

	for key := range measure.Attrs {
		if _, ok := p.dropAttrs[key]; ok {
			delete(measure.Attrs, key)
			continue
		}
		keys = append(keys, key)
	}
	slices.Sort(keys)

	digest := ctx.ResettedDigest()

	for _, key := range keys {
		value := measure.Attrs[key]
		values = append(values, value)

		digest.WriteString(key)
		digest.WriteString(value)
	}

	measure.Time = measure.Time.Truncate(time.Minute)
	measure.AttrsHash = digest.Sum64()
	measure.AttrKeys = keys
	measure.AttrValues = values
}

type MetricKey struct {
	ProjectID uint32
	Metric    string
}

func (p *MeasureProcessor) upsertMetric(ctx context.Context, measure *Measure) {
	key := MetricKey{
		ProjectID: measure.ProjectID,
		Metric:    measure.Metric,
	}

	p.metricMapMu.RLock()
	_, ok := p.metricMap[key]
	p.metricMapMu.RUnlock()
	if ok {
		return
	}

	p.metricMapMu.Lock()
	defer p.metricMapMu.Unlock()

	if _, ok := p.metricMap[key]; ok {
		return
	}
	p.metricMap[key] = struct{}{}

	metric := &Metric{
		ProjectID:   measure.ProjectID,
		Name:        measure.Metric,
		Description: measure.Description,
		Unit:        measure.Unit,
		Instrument:  measure.Instrument,
	}
	inserted, err := UpsertMetric(ctx, p.App, metric)
	if err != nil {
		p.Zap(ctx).Error("CreateMetric failed", zap.Error(err))
		return
	}
	if inserted {
		p.dashSyncer.Sync(ctx, metric.ProjectID)
	}
}

//------------------------------------------------------------------------------

type measureContext struct {
	context.Context

	digest *xxhash.Digest
}

func newMeasureContext(ctx context.Context) *measureContext {
	return &measureContext{
		Context: ctx,
		digest:  xxhash.New(),
	}
}

func (ctx *measureContext) ResettedDigest() *xxhash.Digest {
	ctx.digest.Reset()
	return ctx.digest
}
