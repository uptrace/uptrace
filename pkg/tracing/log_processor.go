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
		conf.Logs.BatchSize,
		conf.Logs.BufferSize,
	)
	thread := NewProcessorThread[Span, LogIndex](processor)

	p := &LogProcessor{
		ProcessorThread: thread,
	}

	p.logger.Info("starting processing logs...",
		zap.Int("threads", runtime.GOMAXPROCS(0)),
		zap.Int("batch_size", conf.Logs.BatchSize),
		zap.Int("buffer_size", conf.Logs.BufferSize))

	app.WaitGroup().Add(1)

	p.logger.Info("Before launching goroutine")

	go func() {
		p.logger.Info("Starting process loop goroutine")

		defer func() {
			p.logger.Info("Goroutine finished, calling WaitGroup Done")
			app.WaitGroup().Done()
		}()

		p.logger.Info("Calling processLoop")
		p.processLoop(app.Context())
		p.logger.Info("processLoop exited")
	}()

	p.logger.Info("After launching goroutine")

	return p
}

func (p *LogProcessor) AddLog(ctx context.Context, log *Span) {
	p.logger.Info("AddLog called", zap.Any("log", log))

	select {
	case p.queue <- log:
		p.logger.Info("Log added to queue", zap.Int("currentQueueSize", len(p.queue)), zap.Int("queueCapacity", cap(p.queue)))
		p.logger.Debug("Log successfully added to the queue")
	default:
		p.logger.Error("Log buffer is full (consider increasing logs.buffer_size)", zap.Int("currentQueueSize", len(p.queue)), zap.Int("queueCapacity", cap(p.queue)))
		p.logger.Info("Calling processItems due to full buffer")

		go p.processItems(ctx, []*Span{log})
		p.logger.Info("Processing log directly since queue is full")

		p.logger.Info("Calling AddItem after processItems")
		p.AddItem(ctx, log)
		p.logger.Info("AddItem call completed after processing the log directly")

		logCounter.Add(
			ctx,
			1,
			metric.WithAttributes(
				bunotel.ProjectIDAttr(log.ProjectID),
				attribute.String("type", "dropped"),
			),
		)
		p.logger.Info("Incremented dropped logs counter")
	}
}

func (p *LogProcessor) processItems(ctx context.Context, logs []*Span) {
	p.logger.Info("processItems called", zap.Int("batch_size", len(logs)))

	ctx, log := bunotel.Tracer.Start(ctx, "process-logs")
	p.logger.Info("Starting processing of logs", zap.Int("batch_size", len(logs)))

	p.ProcessorThread.processItems(ctx, logs)

	p.logger.Info("Adding to WaitGroup")
	p.App.WaitGroup().Add(1)
	p.logger.Info("Starting goroutine for processLoop")
	p.gate.Start()

	p.logger.Info("Attempting to start log processing goroutine", zap.Int("batch_size", len(logs)))

	select {
	case <-ctx.Done():
		p.logger.Error("Context canceled before starting goroutine", zap.Error(ctx.Err()))
		return
	default:
		p.logger.Info("Context is active, starting goroutine")
	}

	go func() {
		p.logger.Info("Log processing goroutine started")
		defer log.End()
		defer p.gate.Done()
		defer p.App.WaitGroup().Done()

		p.logger.Info("Creating new LogProcessorThread")

		thread := newLogProcessorThread(p)

		p.logger.Info("Calling _processLogs", zap.Int("logs_count", len(logs)))
		thread._processLogs(ctx, logs)

		p.logger.Info("Finished processing logs in goroutine", zap.Int("logs_count", len(logs)))
	}()

	p.logger.Info("Goroutine for log processing launched")
}

func (p *logProcessorThread) _processLogs(ctx context.Context, logs []*Span) {
	p.logger.Info("Started processing logs", zap.Int("count", len(logs)))
	indexedLogs := make([]LogIndex, 0, len(logs))
	dataLogs := make([]LogData, 0, len(logs))

	seenErrors := make(map[uint64]bool)

	for _, log := range logs {
		p.logger.Debug("Processing log", zap.String("logID", log.ID.String()), zap.Any("attributes", log.Attrs))
		p.initLogOrEvent(ctx, log)
		logCounter.Add(
			ctx,
			1,
			metric.WithAttributes(
				bunotel.ProjectIDAttr(log.ProjectID),
				attribute.String("type", "inserted"),
			),
		)

		p.logger.Debug("Initializing log index", zap.String("logID", log.ID.String()))
		indexedLogs = append(indexedLogs, LogIndex{})
		p.logger.Debug("Log added to indexedLogs", zap.String("logID", log.ID.String()))
		index := &indexedLogs[len(indexedLogs)-1]
		initLogIndex(index, log)

		if log.EventName != "" {
			p.logger.Debug("Initializing log data", zap.String("logID", log.ID.String()))
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

			p.logger.Debug("Processing event", zap.String("eventID", eventSpan.ID.String()), zap.Any("attributes", eventSpan.Attrs))

			logCounter.Add(
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
					p.logger.Debug("Scheduling error alert", zap.String("eventID", eventSpan.ID.String()))
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

	p.logger.Info("Inserting logs data into logs_data_buffer", zap.Int("dataLogsCount", len(dataLogs)))

	if _, err := p.App.CH.NewInsert().
		Model(&dataLogs).
		ModelTableExpr("logs_data_buffer").
		Exec(ctx); err != nil {
		p.App.Logger.Error("CH insert failed",
			zap.Error(err),
			zap.String("table", "logs_index"))
	}

	p.logger.Info("Inserting logs data into logs_index_buffer", zap.Int("dataLogsCount", len(dataLogs)))

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
