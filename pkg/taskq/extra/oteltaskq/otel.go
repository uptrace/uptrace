package oteltaskq

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/vmihailenco/taskq/v4"
)

var tracer = otel.Tracer("github.com/vmihailenco/taskq")

type OpenTelemetryHook struct{}

var _ taskq.ConsumerHook = (*OpenTelemetryHook)(nil)

func NewHook() *OpenTelemetryHook {
	return new(OpenTelemetryHook)
}

func (h OpenTelemetryHook) BeforeProcessJob(
	ctx context.Context, evt *taskq.ProcessJobEvent,
) context.Context {
	ctx, _ = tracer.Start(ctx, evt.Job.TaskName)
	return ctx
}

func (h OpenTelemetryHook) AfterProcessJob(ctx context.Context, evt *taskq.ProcessJobEvent) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	if !span.IsRecording() {
		return
	}

	if err := evt.Job.Err; err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	}
}
