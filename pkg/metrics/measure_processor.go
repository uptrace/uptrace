package metrics

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"go.uber.org/zap"
	"go4.org/syncutil"
	"golang.org/x/exp/slices"
)

type MeasureProcessor struct {
	*bunapp.App

	batchSize int
	ch        chan *Measure
	gate      *syncutil.Gate

	logger *otelzap.Logger
}

func NewMeasureProcessor(app *bunapp.App) *MeasureProcessor {
	cfg := app.Config()
	p := &MeasureProcessor{
		App: app,

		batchSize: cfg.Metrics.BatchSize,
		ch:        make(chan *Measure, cfg.Metrics.BufferSize),
		gate:      syncutil.NewGate(runtime.GOMAXPROCS(0)),

		logger: app.ZapLogger(),
	}

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()

		p.processLoop(app.Context())
	}()

	return p
}

func (s *MeasureProcessor) AddMeasure(measure *Measure) {
	select {
	case s.ch <- measure:
	default:
		s.logger.Error("measure buffer is full (measure is dropped)")
	}
}

func (s *MeasureProcessor) processLoop(ctx context.Context) {
	const timeout = time.Second

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	measures := make([]*Measure, 0, s.batchSize)

loop:
	for {
		select {
		case measure := <-s.ch:
			measures = append(measures, measure)
		case <-timer.C:
			if len(measures) > 0 {
				s.flushMeasures(ctx, measures)
				measures = make([]*Measure, 0, len(measures))
			}
			timer.Reset(timeout)
		case <-s.Done():
			break loop
		}

		if len(measures) == s.batchSize {
			s.flushMeasures(ctx, measures)
			measures = make([]*Measure, 0, len(measures))
		}
	}

	if len(measures) > 0 {
		s.flushMeasures(ctx, measures)
	}
}

func (s *MeasureProcessor) flushMeasures(ctx context.Context, measures []*Measure) {
	ctx, measure := bunapp.Tracer.Start(ctx, "flush-measures")

	s.WaitGroup().Add(1)
	s.gate.Start()

	go func() {
		defer measure.End()
		defer s.gate.Done()
		defer s.WaitGroup().Done()

		ctx := newMeasureContext(ctx)
		s._flushMeasures(ctx, measures)
	}()
}

func (s *MeasureProcessor) _flushMeasures(ctx *measureContext, measures []*Measure) {
	for _, m := range measures {
		initMeasure(ctx, m)
	}

	if err := InsertMeasures(ctx, s.App, measures); err != nil {
		s.Zap(ctx).Error("InsertMeasures failed", zap.Error(err))
	}
}

func initMeasure(ctx *measureContext, measure *Measure) {
	keys := make([]string, 0, len(measure.Attrs))
	values := make([]string, 0, len(measure.Attrs))

	for key := range measure.Attrs {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	digest := ctx.ResettedDigest()

	for _, key := range keys {
		value := fmt.Sprint(measure.Attrs[key])
		values = append(values, value)

		digest.WriteString(key)
		digest.WriteString(value)
	}

	measure.Time = measure.Time.Truncate(time.Minute)
	measure.AttrsHash = digest.Sum64()
	measure.Keys = keys
	measure.Values = values
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
