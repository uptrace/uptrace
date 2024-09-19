package tracing

import (
	"context"
	"runtime"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"go4.org/syncutil"
	"golang.org/x/exp/slices"
)

type LogProcessor struct {
	*bunapp.App

	batchSize int
	queue     chan *Span
	gate      *syncutil.Gate

	logger *otelzap.Logger
}

func NewLogProcessor(app *bunapp.App) *LogProcessor {
	conf := app.Config()
	maxprocs := runtime.GOMAXPROCS(0)

	p := &LogProcessor{
		App: app,

		batchSize: conf.Logs.BatchSize,
		queue:     make(chan *Span, conf.Logs.BufferSize*2),
		gate:      syncutil.NewGate(maxprocs),

		logger: app.Logger,
	}

	p.logger.Info("starting processing logs...",
		zap.Int("threads", maxprocs),
		zap.Int("batch_size", p.batchSize),
		zap.Int("buffer_size", conf.Logs.BufferSize))

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()
		p.processLoop(app.Context())
	}()

	queueLen, _ := bunotel.Meter.Int64ObservableGauge("uptrace.tracing.queue_length",
		metric.WithUnit("{logs}"),
	)

	if _, err := bunotel.Meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			o.ObserveInt64(queueLen, int64(len(p.queue)))
			return nil
		},
		queueLen,
	); err != nil {
		p.logger.Error("Failed to register queue length metric", zap.Error(err))
		panic(err)
	}

	return p
}

func (p *LogProcessor) AddLog(ctx context.Context, log *Span) {
	select {
	case p.queue <- log:
	default:
		p.logger.Error("Log buffer is full (consider increasing logs.buffer_size)", zap.Int("len", len(p.queue)))
		logCounter.Add(
			ctx,
			1,
			metric.WithAttributes(
				bunotel.ProjectIDAttr(log.ProjectID),
				attribute.String("type", "dropped"),
			),
		)
	}
}

func (p *LogProcessor) processLoop(ctx context.Context) {
	const timeout = 5 * time.Second
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	logs := make([]*Span, 0, p.batchSize)

loop:
	for {
		select {
		case log := <-p.queue:
			logs = append(logs, log)

			if len(logs) < p.batchSize {
				break
			}

			p.processLogs(ctx, logs)
			logs = logs[:0]

			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(timeout)

		case <-timer.C:
			if len(logs) > 0 {
				p.processLogs(ctx, logs)
				logs = logs[:0]
			} else {
				p.logger.Info("Timer expired but no logs to process")
			}
			timer.Reset(timeout)

		case <-p.Done():
			break loop
		}
	}

	if len(logs) > 0 {
		p.processLogs(ctx, logs)
	}
}

func (p *LogProcessor) processLogs(ctx context.Context, src []*Span) {
	ctx, logSpan := bunotel.Tracer.Start(ctx, "process-logs")
	defer logSpan.End()

	p.WaitGroup().Add(1)
	p.gate.Start()

	logs := make([]*Span, len(src))
	copy(logs, src)

	go func() {
		defer p.gate.Done()
		defer p.WaitGroup().Done()

		thread := newLogProcessorThread(p)
		thread._processLogs(ctx, logs)
	}()
}
func (p *logProcessorThread) _processLogs(ctx context.Context, logs []*Span) {
	indexedLogs := make([]LogIndex, 0, len(logs))
	dataLogs := make([]LogData, 0, len(logs))

	seenErrors := make(map[uint64]bool)

	for _, log := range logs {
		p.initLogOrEvent(ctx, log)
		spanCounter.Add(
			ctx,
			1,
			metric.WithAttributes(
				bunotel.ProjectIDAttr(log.ProjectID),
				attribute.String("type", "inserted"),
			),
		)

		indexedLogs = append(indexedLogs, LogIndex{})
		index := &indexedLogs[len(indexedLogs)-1]
		initLogIndex(index, log)

		// if p.sgp != nil {
		// 	if err := p.sgp.ProcessSpan(ctx, index); err != nil {
		// 		p.Zap(ctx).Error("service graph failed", zap.Error(err))
		// 	}
		// }

		if log.EventName != "" {
			dataLogs = append(dataLogs, LogData{})
			initLogData(&dataLogs[len(dataLogs)-1], log)
			continue
		}

		var errorCount int
		var logCount int

		for _, event := range log.Events {
			eventSpan := &Span{
				Attrs: NewAttrMap(),
			}
			initEventFromHostSpan(eventSpan, event, log)
			p.initEvent(ctx, eventSpan)

			spanCounter.Add(
				ctx,
				1,
				metric.WithAttributes(
					bunotel.ProjectIDAttr(log.ProjectID),
					attribute.String("type", "inserted"),
				),
			)

			indexedLogs = append(indexedLogs, LogIndex{})
			initLogIndex(&indexedLogs[len(indexedLogs)-1], eventSpan)

			dataLogs = append(dataLogs, LogData{})
			initLogData(&dataLogs[len(dataLogs)-1], eventSpan)

			if isErrorSystem(eventSpan.System) {
				errorCount++
				if !seenErrors[eventSpan.GroupID] {
					seenErrors[eventSpan.GroupID] = true
					scheduleCreateErrorAlert(ctx, p.App, eventSpan)
				}
			}
			if isLogSystem(eventSpan.System) {
				logCount++
			}
		}

		index.LinkCount = uint8(len(log.Links))
		index.EventCount = uint8(len(log.Events))
		index.EventErrorCount = uint8(errorCount)
		index.EventLogCount = uint8(logCount)
		log.Events = nil

		dataLogs = append(dataLogs, LogData{})
		initLogData(&dataLogs[len(dataLogs)-1], log)
	}

	if _, err := p.CH.NewInsert().
		Model(&dataLogs).
		Exec(ctx); err != nil {
		p.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err),
			zap.String("table", "logs_data"))
	}

	if _, err := p.CH.NewInsert().
		Model(&indexedLogs).
		Exec(ctx); err != nil {
		p.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err),
			zap.String("table", "logs_index"))
	}
}

type logProcessorThread struct {
	*LogProcessor

	projects map[uint32]*org.Project
	digest   *xxhash.Digest
}

func newLogProcessorThread(p *LogProcessor) *logProcessorThread {
	return &logProcessorThread{
		LogProcessor: p,

		projects: make(map[uint32]*org.Project),
		digest:   xxhash.New(),
	}
}

func (p *logProcessorThread) project(ctx context.Context, projectID uint32) (*org.Project, bool) {
	if project, ok := p.projects[projectID]; ok {
		return project, true
	}

	project, err := org.SelectProject(ctx, p.App, projectID)
	if err != nil {
		p.Zap(ctx).Error("SelectProject failed", zap.Error(err))
		return nil, false
	}

	p.projects[projectID] = project
	return project, true
}

func (p *logProcessorThread) forceLogName(ctx context.Context, log *Span) bool {
	if log.EventName != "" {
		return false
	}

	project, ok := p.project(ctx, log.ProjectID)
	if !ok {
		return false
	}

	if libName, _ := log.Attrs[attrkey.OtelLibraryName].(string); libName != "" {
		return slices.Contains(project.ForceSpanName, libName)
	}
	return false
}
