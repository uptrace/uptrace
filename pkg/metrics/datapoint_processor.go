package metrics

import (
	"context"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/vmihailenco/taskq/v4"
	"github.com/zyedidia/generic/cache"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go4.org/syncutil"
	"golang.org/x/exp/slices"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/bfloat16"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
)

type DatapointProcessorParams struct {
	fx.In

	Logger    *otelzap.Logger
	Conf      *bunconf.Config
	PG        *bun.DB
	CH        *ch.DB
	MainQueue taskq.Queue
}

type DatapointProcessor struct {
	*DatapointProcessorParams

	batchSize int
	dropAttrs map[string]struct{}

	queue chan *Datapoint
	gate  *syncutil.Gate

	c2d *CumToDeltaConv

	metricCacheMu sync.RWMutex
	metricCache   *cache.Cache[MetricKey, time.Time]

	dashSyncer *DashSyncer

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewDatapointProcessor(p DatapointProcessorParams) *DatapointProcessor {
	maxprocs := runtime.GOMAXPROCS(0)

	conf := p.Conf.Metrics
	dp := &DatapointProcessor{
		DatapointProcessorParams: &p,

		batchSize: conf.BatchSize,

		queue: make(chan *Datapoint, conf.BufferSize),
		gate:  syncutil.NewGate(maxprocs),

		c2d: NewCumToDeltaConv(conf.CumToDeltaSize),

		metricCache: cache.New[MetricKey, time.Time](conf.CumToDeltaSize),

		dashSyncer: NewDashSyncer(p.Logger, p.PG),
	}

	if len(p.Conf.Metrics.DropAttrs) > 0 {
		dp.dropAttrs = listToSet(p.Conf.Metrics.DropAttrs)
	}

	p.Logger.Info("starting processing metrics...",
		zap.Int("threads", maxprocs),
		zap.Int("batch_size", dp.batchSize),
		zap.Int("buffer_size", conf.BufferSize),
		zap.Int("cum_to_delta_size", conf.CumToDeltaSize))

	queueLen, _ := bunotel.Meter.Int64ObservableGauge("uptrace.metrics.queue_length",
		metric.WithUnit("{datapoints}"),
	)

	if _, err := bunotel.Meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			o.ObserveInt64(queueLen, int64(len(dp.queue)))
			return nil
		},
		queueLen,
	); err != nil {
		panic(err)
	}

	return dp
}

func (p *DatapointProcessor) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	p.processLoop(ctx)
}

func (p *DatapointProcessor) Stop() {
	if p.cancel == nil {
		p.Logger.Error("no cancel function registered for BaseConsumer")
		return
	}

	p.cancel()
	p.wg.Wait()
}

func (p *DatapointProcessor) AddDatapoint(ctx context.Context, datapoint *Datapoint) {
	select {
	case p.queue <- datapoint:
	default:
		p.Logger.Error("datapoint buffer is full (consider increasing metrics.buffer_size)",
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
		case <-ctx.Done():
			break loop
		}
	}

	if len(datapoints) > 0 {
		p.processDatapoints(ctx, datapoints)
	}
}

func (p *DatapointProcessor) processDatapoints(ctx context.Context, src []*Datapoint) {
	ctx, span := bunotel.Tracer.Start(ctx, "process-datapoints")

	p.wg.Add(1)
	p.gate.Start()

	datapoints := make([]*Datapoint, len(src))
	copy(datapoints, src)

	go func() {
		defer span.End()
		defer p.gate.Done()
		defer p.wg.Done()

		mctx := newDatapointContext(ctx)
		p._processDatapoints(mctx, datapoints)
	}()
}

func (p *DatapointProcessor) _processDatapoints(ctx *datapointContext, datapoints []*Datapoint) {
	for i := len(datapoints) - 1; i >= 0; i-- {
		dp := datapoints[i]
		p.initDatapoint(ctx, dp)

		if !p.cumToDelta(ctx, dp) {
			datapoints = append(datapoints[:i], datapoints[i+1:]...)
			datapointCounter.Add(
				ctx,
				1,
				metric.WithAttributes(
					bunotel.ProjectIDAttr(dp.ProjectID),
					attribute.String("type", "dropped"),
				),
			)
			continue
		}

		project := ctx.Project(p.Logger, p.PG, dp.ProjectID)
		if project == nil {
			datapoints = append(datapoints[:i], datapoints[i+1:]...)
			datapointCounter.Add(
				ctx,
				1,
				metric.WithAttributes(
					bunotel.ProjectIDAttr(dp.ProjectID),
					attribute.String("type", "dropped"),
				),
			)
			continue
		}

		datapointCounter.Add(
			ctx,
			1,
			metric.WithAttributes(
				bunotel.ProjectIDAttr(dp.ProjectID),
				attribute.String("type", "inserted"),
			),
		)

		p.upsertMetric(ctx, dp)

		if project.PromCompat {
			dp.Time = dp.Time.Truncate(15 * time.Second)
		} else {
			dp.Time = dp.Time.Truncate(time.Minute)
		}
	}

	if len(datapoints) > 0 {
		if err := InsertDatapoints(ctx, p.CH, datapoints); err != nil {
			p.Logger.Error("InsertDatapoints failed", zap.Error(err))
		}
	}

	if len(ctx.metrics) > 0 {
		if err := UpsertMetrics(ctx, p.Logger, p.PG, p.MainQueue, ctx.metrics); err != nil {
			p.Logger.Error("upsertMetrics failed", zap.Error(err))
		}
	}
}

func (p *DatapointProcessor) cumToDelta(ctx *datapointContext, datapoint *Datapoint) bool {
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
		p.Logger.Error("unknown cum point type",
			zap.String("type", reflect.TypeOf(point).String()))
		return false
	}
}

func (p *DatapointProcessor) initDatapoint(ctx *datapointContext, datapoint *Datapoint) {
	normAttrs(datapoint.Attrs)

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
		p.Logger.Error("number of buckets does not match")
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

	projects map[uint32]*org.Project
	digest   *xxhash.Digest
	metrics  []Metric
}

func newDatapointContext(ctx context.Context) *datapointContext {
	return &datapointContext{
		Context:  ctx,
		projects: make(map[uint32]*org.Project),
		digest:   xxhash.New(),
	}
}

func (c *datapointContext) Project(logger *otelzap.Logger, pg *bun.DB, projectID uint32) *org.Project {
	if p, ok := c.projects[projectID]; ok {
		return p
	}

	project, err := org.SelectProject(c.Context, pg, projectID)
	if err != nil {
		logger.Error("SelectProject failed", zap.Error(err))
		return nil
	}

	c.projects[projectID] = project
	return project
}

func (c *datapointContext) ResettedDigest() *xxhash.Digest {
	c.digest.Reset()
	return c.digest
}

//------------------------------------------------------------------------------

type AttrName struct {
	Canonical string
	Alts      []string
}

var attrNames = []AttrName{
	{
		Canonical: attrkey.DeploymentEnvironment,
		Alts:      []string{"deployment_environment", "environment", "env"},
	},
	{Canonical: attrkey.ServiceName, Alts: []string{"service", "component"}},
	{Canonical: attrkey.URLScheme, Alts: []string{"http_scheme"}},
	{Canonical: attrkey.URLFull, Alts: []string{"http_url"}},
	{Canonical: attrkey.URLPath, Alts: []string{"http_target"}},
	{Canonical: attrkey.HTTPRequestMethod, Alts: []string{"http_method"}},
	{Canonical: attrkey.HTTPResponseStatusCode, Alts: []string{"http_status_code"}},
	{Canonical: attrkey.HTTPResponseStatusClass, Alts: []string{"http_status_class"}},
}

func normAttrs(attrs AttrMap) {
	for _, name := range attrNames {
		if _, ok := attrs[name.Canonical]; ok {
			continue
		}

		for _, key := range name.Alts {
			if val, ok := attrs[key]; ok {
				delete(attrs, key)
				attrs[name.Canonical] = val
				break
			}
		}
	}
}
