package tracing

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/tracing/attrkey"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.uber.org/zap"
	"go4.org/syncutil"
)

type SpanProcessor struct {
	*bunapp.App

	batchSize int
	ch        chan *Span
	gate      *syncutil.Gate

	notifier *bunapp.Notifier

	logger *otelzap.Logger
}

func NewSpanProcessor(app *bunapp.App) *SpanProcessor {
	conf := app.Config()
	maxprocs := runtime.GOMAXPROCS(0)
	p := &SpanProcessor{
		App: app,

		batchSize: conf.Spans.BatchSize,
		ch:        make(chan *Span, conf.Spans.BufferSize),
		gate:      syncutil.NewGate(maxprocs),

		logger: app.Logger,
	}

	p.logger.Info("starting processing spans...",
		zap.Int("threads", maxprocs),
		zap.Int("batch_size", p.batchSize),
		zap.Int("buffer_size", conf.Spans.BufferSize))

	app.WaitGroup().Add(1)
	go func() {
		defer app.WaitGroup().Done()

		p.processLoop(app.Context())
	}()

	bufferSize, _ := bunotel.Meter.AsyncInt64().Gauge("uptrace.spans.buffer_size")

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

func (p *SpanProcessor) AddSpan(ctx context.Context, span *Span) {
	select {
	case p.ch <- span:
	default:
		p.logger.Error("span buffer is full (consider increasing spans.buffer_size)")
		spanCounter.Add(
			ctx,
			1,
			bunotel.ProjectIDAttr(span.ProjectID),
			attribute.String("type", "dropped"),
		)
	}
}

func (s *SpanProcessor) processLoop(ctx context.Context) {
	const timeout = 5 * time.Second

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	spans := make([]*Span, 0, s.batchSize)

loop:
	for {
		select {
		case span := <-s.ch:
			spans = append(spans, span)

			if len(spans) < s.batchSize {
				break
			}
			s.flushSpans(ctx, spans)
			spans = make([]*Span, 0, len(spans))
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(timeout)
		case <-timer.C:
			if len(spans) > 0 {
				s.flushSpans(ctx, spans)
				spans = make([]*Span, 0, len(spans))
			}
			timer.Reset(timeout)
		case <-s.Done():
			break loop
		}
	}

	if len(spans) > 0 {
		s.flushSpans(ctx, spans)
	}
}

func (s *SpanProcessor) flushSpans(ctx context.Context, spans []*Span) {
	ctx, span := bunotel.Tracer.Start(ctx, "flush-spans")

	s.WaitGroup().Add(1)
	s.gate.Start()

	go func() {
		defer span.End()
		defer s.gate.Done()
		defer s.WaitGroup().Done()

		s._flushSpans(ctx, spans)
	}()
}

func (s *SpanProcessor) _flushSpans(ctx context.Context, spans []*Span) {
	indexedSpans := make([]SpanIndex, 0, len(spans))
	dataSpans := make([]SpanData, 0, len(spans))

	seenErrors := make(map[uint64]bool) // basic deduplication
	var errors []*Span

	spanCtx := newSpanContext(ctx)
	for _, span := range spans {
		initSpan(spanCtx, span)
		spanCounter.Add(
			ctx,
			1,
			bunotel.ProjectIDAttr(span.ProjectID),
			attribute.String("type", "inserted"),
		)

		indexedSpans = append(indexedSpans, SpanIndex{})
		index := &indexedSpans[len(indexedSpans)-1]
		initSpanIndex(index, span)

		dataSpans = append(dataSpans, SpanData{})
		initSpanData(&dataSpans[len(dataSpans)-1], span)

		var errorCount int
		var logCount int

		for _, eventSpan := range span.Events {
			initSpanEvent(spanCtx, eventSpan, span)
			spanCounter.Add(
				ctx,
				1,
				bunotel.ProjectIDAttr(span.ProjectID),
				attribute.String("type", "inserted"),
			)

			indexedSpans = append(indexedSpans, SpanIndex{})
			initSpanIndex(&indexedSpans[len(indexedSpans)-1], eventSpan)

			dataSpans = append(dataSpans, SpanData{})
			initSpanData(&dataSpans[len(dataSpans)-1], eventSpan)

			if isErrorSystem(eventSpan.System) {
				errorCount++
				if !seenErrors[eventSpan.GroupID] {
					seenErrors[eventSpan.GroupID] = true
					errors = append(errors, eventSpan)
				}
			}
			if isLogSystem(eventSpan.System) {
				logCount++
			}
		}

		index.LinkCount = uint8(len(span.Links))
		index.EventCount = uint8(len(span.Events))
		index.EventErrorCount = uint8(errorCount)
		index.EventLogCount = uint8(logCount)
	}

	if _, err := s.CH.NewInsert().Model(&dataSpans).Exec(ctx); err != nil {
		s.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err), zap.String("table", "spans_data"))
	}

	if _, err := s.CH.NewInsert().Model(&indexedSpans).Exec(ctx); err != nil {
		s.Zap(ctx).Error("ch.Insert failed",
			zap.Error(err), zap.String("table", "spans_index"))
	}

	if len(errors) > 0 {
		s.notifyOnErrors(ctx, errors)
	}
}

func (s *SpanProcessor) notifyOnErrors(ctx context.Context, errors []*Span) {
	conf := s.Config()
	if !conf.Alerting.CreateAlertsFromSpans.Enabled {
		return
	}

	alerts := make(models.PostableAlerts, len(errors))

	for i, error := range errors {
		labels := models.LabelSet{
			"alertname":  "New error has occurred",
			"project_id": strconv.FormatUint(uint64(error.ProjectID), 10),
			"system":     error.System,
			"group_id":   strconv.FormatUint(error.GroupID, 10),
		}
		if service := error.Attrs.ServiceName(); service != "" {
			labels["service"] = service
		}
		if sev, _ := error.Attrs[attrkey.LogSeverity].(string); sev != "" {
			labels["severity"] = sev
		}
		for k, v := range conf.Alerting.CreateAlertsFromSpans.Labels {
			labels[k] = v
		}
		traceURL := s.Config().SitePath(fmt.Sprintf("/traces/%s", error.TraceID.String()))

		alerts[i] = &models.PostableAlert{
			Alert: models.Alert{
				Labels:       labels,
				GeneratorURL: strfmt.URI(traceURL),
			},
			Annotations: models.LabelSet{
				"span_name":  error.Name,
				"event_name": error.EventName,
				"trace_id":   error.TraceID.String(),
			},
		}
	}

	s.Notifier.Send(ctx, alerts)
}
