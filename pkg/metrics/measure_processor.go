package metrics

import (
	"context"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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
	maxprocs := runtime.GOMAXPROCS(0)
	p := &MeasureProcessor{
		App: app,

		batchSize: conf.Metrics.BatchSize,

		ch:   make(chan *Measure, conf.Metrics.BufferSize),
		gate: syncutil.NewGate(maxprocs),

		c2d:    NewCumToDeltaConv(conf.Metrics.CumToDeltaSize),
		logger: app.Logger,

		metricMap:  make(map[MetricKey]struct{}),
		dashSyncer: NewDashSyncer(app),
	}

	if len(conf.Metrics.DropAttrs) > 0 {
		p.dropAttrs = listToSet(conf.Metrics.DropAttrs)
	}

	p.logger.Info("starting processing metrics...",
		zap.Int("threads", maxprocs),
		zap.Int("batch_size", p.batchSize),
		zap.Int("buffer_size", conf.Metrics.BufferSize),
		zap.Int("cum_to_delta_size", conf.Metrics.CumToDeltaSize))

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()

		p.processLoop(app.Context())
	}()

	bufferSize, _ := bunotel.Meter.Int64ObservableGauge("uptrace.measures.buffer_size")

	if _, err := bunotel.Meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			o.ObserveInt64(bufferSize, int64(len(p.ch)))
			return nil
		},
		bufferSize,
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

	mctx := newMeasureContext(ctx)
	measures := make([]*Measure, 0, p.batchSize)

loop:
	for {
		select {
		case measure := <-p.ch:
			if !p.processMeasure(mctx, measure) {
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

			if len(measures) < p.batchSize {
				break
			}

			p.flushMeasures(ctx, measures)
			measures = measures[:0]

			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(timeout)
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

func (p *MeasureProcessor) processMeasure(ctx *measureContext, measure *Measure) bool {
	p.initMeasure(ctx, measure)

	switch point := measure.CumPoint.(type) {
	case nil:
		return true
	case *NumberPoint:
		if !p.convertNumberPoint(ctx, measure, point) {
			return false
		}
		measure.CumPoint = nil
		return true
	case *HistogramPoint:
		if !p.convertHistogramPoint(ctx, measure, point) {
			return false
		}
		measure.CumPoint = nil
		return true
	case *ExpHistogramPoint:
		if !p.convertExpHistogramPoint(ctx, measure, point) {
			return false
		}
		measure.CumPoint = nil
		return true
	default:
		p.Zap(ctx).Error("unknown cum point type",
			zap.String("type", reflect.TypeOf(point).String()))
		return false
	}
}

func (p *MeasureProcessor) convertNumberPoint(
	ctx context.Context, measure *Measure, point *NumberPoint,
) bool {
	key := MeasureKey{
		ProjectID:         measure.ProjectID,
		Metric:            measure.Metric,
		AttrsHash:         measure.AttrsHash,
		StartTimeUnixNano: measure.StartTimeUnixNano,
	}

	prevPoint, ok := p.c2d.SwapPoint(key, point, measure.Time).(*NumberPoint)
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

func (p *MeasureProcessor) convertHistogramPoint(
	ctx context.Context, measure *Measure, point *HistogramPoint,
) bool {
	key := MeasureKey{
		ProjectID:         measure.ProjectID,
		Metric:            measure.Metric,
		AttrsHash:         measure.AttrsHash,
		StartTimeUnixNano: measure.StartTimeUnixNano,
	}

	prevPoint, ok := p.c2d.SwapPoint(key, point, measure.Time).(*HistogramPoint)
	if !ok {
		return false
	}
	if len(point.BucketCounts) != len(prevPoint.BucketCounts) {
		p.Zap(ctx).Error("number of buckets does not match")
		return false
	}

	measure.Count = point.Count - prevPoint.Count
	if measure.Count <= 0 {
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

func (p *MeasureProcessor) convertExpHistogramPoint(
	ctx context.Context, measure *Measure, point *ExpHistogramPoint,
) bool {
	key := MeasureKey{
		ProjectID:         measure.ProjectID,
		Metric:            measure.Metric,
		AttrsHash:         measure.AttrsHash,
		StartTimeUnixNano: measure.StartTimeUnixNano,
	}

	prevPoint, ok := p.c2d.SwapPoint(key, point, measure.Time).(*ExpHistogramPoint)
	if !ok {
		return false
	}

	if point.Count < prevPoint.Count {
		return false
	}

	measure.Count = point.Count - prevPoint.Count
	measure.Sum = point.Sum - prevPoint.Sum

	var hist map[bfloat16.T]uint64
	measure.Histogram = hist

	for mean, count := range point.Histogram {
		prevCount := prevPoint.Histogram[mean]
		count -= prevCount
		if count > 0 {
			if hist == nil {
				hist = make(map[bfloat16.T]uint64, len(point.Histogram))
			}
			hist[mean] = count
		}
	}

	return true
}

func (p *MeasureProcessor) flushMeasures(ctx context.Context, src []*Measure) {
	ctx, span := bunotel.Tracer.Start(ctx, "flush-measures")

	p.WaitGroup().Add(1)
	p.gate.Start()

	measures := make([]*Measure, len(src))
	copy(measures, src)

	go func() {
		defer span.End()
		defer p.gate.Done()
		defer p.WaitGroup().Done()

		for _, measure := range measures {
			measure.Time = measure.Time.Truncate(time.Minute)
		}

		if err := InsertMeasures(ctx, p.App, measures); err != nil {
			p.Zap(ctx).Error("InsertMeasures failed", zap.Error(err))
		}
	}()
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

	measure.AttrsHash = digest.Sum64()
	measure.StringKeys = keys
	measure.StringValues = values
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
		p.dashSyncer.CreateDashboards(ctx, metric.ProjectID)
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
