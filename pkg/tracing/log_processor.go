package tracing

import (
	"context"
	"runtime"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

type LogProcessor struct {
	*ProcessorThread[Span, LogIndex]
}

func NewLogProcessor(app *bunapp.App) *LogProcessor {
	conf := app.Config()

	processor := NewProcessor[Span](
		app,
		conf.Spans.BatchSize,
		conf.Spans.BufferSize,
	)
	thread := NewProcessorThread[Span, LogIndex](processor)

	p := &LogProcessor{
		ProcessorThread: thread,
	}

	p.logger.Info("starting processing logs...",
		zap.Int("threads", runtime.GOMAXPROCS(0)),
		zap.Int("batch_size", conf.Spans.BatchSize),
		zap.Int("buffer_size", conf.Spans.BufferSize))

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()
		p.processLoop(app.Context())
	}()

	return p
}

func (p *LogProcessor) AddLog(ctx context.Context, log *Span) {
	select {
	case p.queue <- log:
		p.logger.Debug("log added")
	default:
		p.logger.Error("Log buffer is full (consider increasing logs.buffer_size)", zap.Int("len", len(p.queue)))
		p.processItems(ctx, []*Span{log})
		p.AddItem(ctx, log)
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

func (p *LogProcessor) processItems(ctx context.Context, logs []*Span) {
	p.logger.Info("processItems called", zap.Int("batch_size", len(logs)))
	ctx, span := bunotel.Tracer.Start(ctx, "process-logs")

	p.ProcessorThread.processItems(ctx, logs)

	p.App.WaitGroup().Add(1)
	p.gate.Start()

	p.logger.Info("Processing batch of logs", zap.Int("batch_size", len(logs)))

	go func() {
		p.logger.Info("Log processing goroutine started")
		defer span.End()
		defer p.gate.Done()
		defer p.App.WaitGroup().Done()

		thread := newLogProcessorThread(p)
		thread._processLogs(ctx, logs)
	}()
}

func (p *logProcessorThread) _processLogs(ctx context.Context, logs []*Span) {
	p.logger.Info("Started processing logs", zap.Int("count", len(logs)))
	indexedLogs := make([]LogIndex, 0, len(logs))
	dataLogs := make([]LogData, 0, len(logs))

	seenErrors := make(map[uint64]bool)

	for _, log := range logs {
		p.logger.Debug("Processing log", zap.String("logID", log.ID.String()), zap.Any("attributes", log.Attrs))
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

	if _, err := p.App.CH.NewInsert().
		Model(&dataLogs).
		ModelTableExpr("logs_data_buffer").
		Exec(ctx); err != nil {
		p.App.Logger.Error("CH insert failed",
			zap.Error(err),
			zap.String("table", "logs_index"))
	}

	if _, err := p.App.CH.NewInsert().
		Model(&indexedLogs).
		ModelTableExpr("logs_index_buffer").
		Exec(ctx); err != nil {
		p.App.Logger.Error("CH insert failed",
			zap.Error(err),
			zap.String("table", "logs_index"))
	}

	p.logger.Info("Finished processing logs")
}

type logProcessorThread struct {
	*ProcessorThread[Span, LogProcessor]
}

func newLogProcessorThread(p *LogProcessor) *logProcessorThread {
	return &logProcessorThread{
		ProcessorThread: NewProcessorThread[Span, LogProcessor](p.Processor),
	}
}

func (p *logProcessorThread) forceLogName(ctx context.Context, log *Span) bool {
	return p.forceName(ctx, log, func(s *Span) map[string]interface{} {
		return s.Attrs
	}, func(s *Span) uint32 {
		return s.ProjectID
	}, func(s *Span) string {
		return s.EventName
	})
}
