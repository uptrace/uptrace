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
	"github.com/zyedidia/generic/cache"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"go4.org/syncutil"
	"golang.org/x/exp/slices"
)

type DatapointProcessor struct {
	*bunapp.App

	batchSize int
	dropAttrs map[string]struct{}

	queue chan *Datapoint
	gate  *syncutil.Gate

	c2d    *CumToDeltaConv
	logger *otelzap.Logger

	metricCacheMu sync.RWMutex
	metricCache   *cache.Cache[MetricKey, time.Time]

	dashSyncer *DashSyncer
}

func NewDatapointProcessor(app *bunapp.App) *DatapointProcessor {
	conf := app.Config()
	maxprocs := runtime.GOMAXPROCS(0)
	p := &DatapointProcessor{
		App: app,

		batchSize: conf.Metrics.BatchSize,

		queue: make(chan *Datapoint, conf.Metrics.BufferSize),
		gate:  syncutil.NewGate(maxprocs),

		c2d:    NewCumToDeltaConv(conf.Metrics.CumToDeltaSize),
		logger: app.Logger,

		metricCache: cache.New[MetricKey, time.Time](conf.Metrics.CumToDeltaSize),

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

	queueLen, _ := bunotel.Meter.Int64ObservableGauge("uptrace.metrics.queue_length",
		metric.WithUnit("{datapoints}"),
	)

	if _, err := bunotel.Meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			o.ObserveInt64(queueLen, int64(len(p.queue)))
			return nil
		},
		queueLen,
	); err != nil {
		panic(err)
	}

	return p
}

func (p *DatapointProcessor) AddDatapoint(ctx context.Context, datapoint *Datapoint) {
	select {
	case p.queue <- datapoint:
	default:
		p.logger.Error("datapoint buffer is full (consider increasing metrics.buffer_size)",
			zap.Int("len", len(p.queue)))
		datapointCounter.Add(
			ctx,
			1,
			metric.WithAttributes(
				bunotel.ProjectIDAttr(datapoint.ProjectID),
				attribute.String("type", "dropped"),
			),
		)
	}
}

func (p *DatapointProcessor) processLoop(ctx context.Context) {
	const timeout = 5 * time.Second

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	datapoints := make([]*Datapoint, 0, p.batchSize)

loop:
	for {
		select {
		case datapoint := <-p.queue:
			datapoints = append(datapoints, datapoint)

			if len(datapoints) < p.batchSize {
				break
			}

			p.processDatapoints(ctx, datapoints)
			datapoints = datapoints[:0]

			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(timeout)
		case <-timer.C:
			if len(datapoints) > 0 {
				p.processDatapoints(ctx, datapoints)
				datapoints = datapoints[:0]
			}
			timer.Reset(timeout)
		case <-p.Done():
			break loop
		}
	}

	if len(datapoints) > 0 {
		p.processDatapoints(ctx, datapoints)
	}
}

func (p *DatapointProcessor) processDatapoints(ctx context.Context, src []*Datapoint) {
	ctx, span := bunotel.Tracer.Start(ctx, "process-datapoints")

	p.WaitGroup().Add(1)
	p.gate.Start()

	datapoints := make([]*Datapoint, len(src))
	copy(datapoints, src)

	go func() {
		defer span.End()
		defer p.gate.Done()
		defer p.WaitGroup().Done()

		mctx := newDatapointContext(ctx)
		p._processDatapoints(mctx, datapoints)
	}()
}

func (p *DatapointProcessor) _processDatapoints(ctx *datapointContext, datapoints []*Datapoint) {
	for i := len(datapoints) - 1; i >= 0; i-- {
		datapoint := datapoints[i]

		if !p.processDatapoint(ctx, datapoint) {
			datapoints = append(datapoints[:i], datapoints[i+1:]...)
			continue
		}

		datapointCounter.Add(
			ctx,
			1,
			metric.WithAttributes(
				bunotel.ProjectIDAttr(datapoint.ProjectID),
				attribute.String("type", "inserted"),
			),
		)

		p.upsertMetric(ctx, datapoint)
		datapoint.Time = datapoint.Time.Truncate(time.Minute)
	}

	if len(datapoints) > 0 {
		if err := InsertDatapoints(ctx, p.App, datapoints); err != nil {
			p.Zap(ctx).Error("InsertDatapoints failed", zap.Error(err))
		}
	}

	if len(ctx.metrics) > 0 {
		if err := UpsertMetrics(ctx, p.App, ctx.metrics); err != nil {
			p.Zap(ctx).Error("upsertMetrics failed", zap.Error(err))
		}
	}
}

func (p *DatapointProcessor) processDatapoint(ctx *datapointContext, datapoint *Datapoint) bool {
	p.initDatapoint(ctx, datapoint)

	switch point := datapoint.CumPoint.(type) {
	case nil:
		return true
	case *NumberPoint:
		if !p.convertNumberPoint(ctx, datapoint, point) {
			return false
		}
		datapoint.CumPoint = nil
		return true
	case *HistogramPoint:
		if !p.convertHistogramPoint(ctx, datapoint, point) {
			return false
		}
		datapoint.CumPoint = nil
		return true
	case *ExpHistogramPoint:
		if !p.convertExpHistogramPoint(ctx, datapoint, point) {
			return false
		}
		datapoint.CumPoint = nil
		return true
	default:
		p.Zap(ctx).Error("unknown cum point type",
			zap.String("type", reflect.TypeOf(point).String()))
		return false
	}
}

func (p *DatapointProcessor) initDatapoint(ctx *datapointContext, datapoint *Datapoint) {
	keys := make([]string, 0, len(datapoint.Attrs))
	values := make([]string, 0, len(datapoint.Attrs))

	for key := range datapoint.Attrs {
		if _, ok := p.dropAttrs[key]; ok {
			delete(datapoint.Attrs, key)
			continue
		}
		keys = append(keys, key)
	}
	slices.Sort(keys)

	digest := ctx.ResettedDigest()

	for _, key := range keys {
		value := datapoint.Attrs[key]
		values = append(values, value)

		digest.WriteString(key)
		digest.WriteString(value)
	}

	datapoint.AttrsHash = digest.Sum64()
	datapoint.StringKeys = keys
	datapoint.StringValues = values
}

func (p *DatapointProcessor) convertNumberPoint(
	ctx context.Context, datapoint *Datapoint, point *NumberPoint,
) bool {
	key := DatapointKey{
		ProjectID:         datapoint.ProjectID,
		Metric:            datapoint.Metric,
		AttrsHash:         datapoint.AttrsHash,
		StartTimeUnixNano: datapoint.StartTimeUnixNano,
	}

	prevPoint, ok := p.c2d.SwapPoint(key, point, datapoint.Time).(*NumberPoint)
	if !ok {
		return false
	}

	if delta := point.Int - prevPoint.Int; delta > 0 {
		datapoint.Sum = float64(delta)
	} else if delta := point.Double - prevPoint.Double; delta > 0 {
		datapoint.Sum = delta
	}
	return true
}

func (p *DatapointProcessor) convertHistogramPoint(
	ctx context.Context, datapoint *Datapoint, point *HistogramPoint,
) bool {
	key := DatapointKey{
		ProjectID:         datapoint.ProjectID,
		Metric:            datapoint.Metric,
		AttrsHash:         datapoint.AttrsHash,
		StartTimeUnixNano: datapoint.StartTimeUnixNano,
	}

	prevPoint, ok := p.c2d.SwapPoint(key, point, datapoint.Time).(*HistogramPoint)
	if !ok {
		return false
	}
	if len(point.BucketCounts) != len(prevPoint.BucketCounts) {
		p.Zap(ctx).Error("number of buckets does not match")
		return false
	}

	if point.Count < prevPoint.Count {
		return false
	}

	datapoint.Sum = point.Sum - prevPoint.Sum
	datapoint.Count = point.Count - prevPoint.Count

	counts := makeDeltaCounts(point.BucketCounts, prevPoint.BucketCounts)
	avg := datapoint.Sum / float64(datapoint.Count)
	datapoint.Histogram, datapoint.Min, datapoint.Max = newBFloat16Histogram(point.Bounds, counts, avg)

	return true
}

func makeDeltaCounts(counts, prevCounts []uint64) []uint64 {
	for i, count := range counts {
		prevCount := prevCounts[i]
		if count > prevCount {
			prevCounts[i] = count - prevCount
		} else {
			prevCounts[i] = 0
		}
	}
	return prevCounts
}

func (p *DatapointProcessor) convertExpHistogramPoint(
	ctx context.Context, datapoint *Datapoint, point *ExpHistogramPoint,
) bool {
	key := DatapointKey{
		ProjectID:         datapoint.ProjectID,
		Metric:            datapoint.Metric,
		AttrsHash:         datapoint.AttrsHash,
		StartTimeUnixNano: datapoint.StartTimeUnixNano,
	}

	prevPoint, ok := p.c2d.SwapPoint(key, point, datapoint.Time).(*ExpHistogramPoint)
	if !ok {
		return false
	}

	if point.Count < prevPoint.Count {
		return false
	}

	datapoint.Sum = point.Sum - prevPoint.Sum
	datapoint.Count = point.Count - prevPoint.Count

	var hist map[bfloat16.T]uint64
	datapoint.Histogram = hist

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

type MetricKey struct {
	ProjectID uint32
	Metric    string
}

func (p *DatapointProcessor) upsertMetric(ctx *datapointContext, datapoint *Datapoint) {
	key := MetricKey{
		ProjectID: datapoint.ProjectID,
		Metric:    datapoint.Metric,
	}

	p.metricCacheMu.RLock()
	cachedAt, found := p.metricCache.Get(key)
	p.metricCacheMu.RUnlock()

	if found && time.Since(cachedAt) < 15*time.Minute {
		return
	}

	p.metricCacheMu.Lock()
	defer p.metricCacheMu.Unlock()

	if cachedAt, found := p.metricCache.Get(key); found && time.Since(cachedAt) < 15*time.Minute {
		return
	}
	p.metricCache.Put(key, time.Now())

	ctx.metrics = append(ctx.metrics, Metric{
		ProjectID:   datapoint.ProjectID,
		Name:        datapoint.Metric,
		Description: datapoint.Description,
		Unit:        datapoint.Unit,
		Instrument:  datapoint.Instrument,
		AttrKeys:    datapoint.StringKeys,
	})
}

//------------------------------------------------------------------------------

type datapointContext struct {
	context.Context

	digest  *xxhash.Digest
	metrics []Metric
}

func newDatapointContext(ctx context.Context) *datapointContext {
	return &datapointContext{
		Context: ctx,
		digest:  xxhash.New(),
	}
}

func (c *datapointContext) ResettedDigest() *xxhash.Digest {
	c.digest.Reset()
	return c.digest
}
