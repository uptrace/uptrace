package tracing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

type SpanData struct {
	ch.CHModel `ch:"table:spans_data_buffer,insert:spans_data_buffer,alias:s"`

	Type      string `ch:",lc"`
	ProjectID uint32
	TraceID   uuid.UUID
	ID        uint64
	ParentID  uint64
	Time      time.Time `ch:"type:DateTime64(6)"`
	Data      []byte
}

func (sd *SpanData) Decode(span *Span) error {
	if err := msgpack.Unmarshal(sd.Data, span); err != nil {
		return fmt.Errorf("msgpack.Unmarshal failed: %w", err)
	}

	span.ProjectID = sd.ProjectID
	span.TraceID = sd.TraceID
	span.ID = sd.ID
	span.ParentID = sd.ParentID
	span.Time = sd.Time
	span.DurationSelf = span.Duration

	span.Type = span.System
	if i := strings.IndexByte(span.Type, ':'); i >= 0 {
		span.Type = span.Type[:i]
	}

	return nil
}

func initSpanData(data *SpanData, span *Span) {
	data.Type = span.Type
	data.ProjectID = span.ProjectID
	data.TraceID = span.TraceID
	data.ID = span.ID
	data.ParentID = span.ParentID
	data.Time = span.Time
	data.Data = marshalSpanData(span)
}

func SelectSpan(ctx context.Context, app *bunapp.App, span *Span) error {
	var data SpanData

	baseq := app.CH.NewSelect().
		ColumnExpr("project_id, trace_id, id, parent_id, time, data").
		Model(&data).
		Where("trace_id = ?", span.TraceID)

	if span.ProjectID != 0 {
		baseq = baseq.Where("project_id = ?", span.ProjectID)
	}

	q := baseq.Clone().Limit(1)
	if span.ID != 0 {
		q = q.Where("id = ?", span.ID)
	}
	if err := q.Scan(ctx); err != nil {
		return err
	}

	if err := data.Decode(span); err != nil {
		return err
	}

	if span.IsEvent() {
		return nil
	}

	var events []*SpanData

	if err := baseq.Clone().
		Where("type IN ?", ch.In(LogAndEventTypes)).
		Where("parent_id = ?", span.ID).
		OrderExpr("time ASC").
		Limit(100).
		Scan(ctx, &events); err != nil {
		return err
	}

	for _, eventData := range events {
		event := new(Span)
		if err := eventData.Decode(event); err != nil {
			return err
		}
		span.AddEvent(event.Event())
	}

	return nil
}

func SelectTraceSpans(
	ctx context.Context, app *bunapp.App, traceID uuid.UUID,
) ([]*Span, bool, error) {
	const limit = 10000

	var data []SpanData

	if err := app.CH.NewSelect().
		DistinctOn("id").
		ColumnExpr("project_id, trace_id, id, parent_id, time, data").
		Model(&data).
		Column("data").
		Where("trace_id = ?", traceID).
		OrderExpr("time ASC").
		Limit(limit).
		Scan(ctx); err != nil {
		return nil, false, err
	}

	spans := make([]*Span, len(data))

	for i := range spans {
		span := new(Span)
		spans[i] = span
		if err := data[i].Decode(span); err != nil {
			return nil, false, err
		}
	}

	return spans, len(spans) == limit, nil
}

func marshalSpanData(span *Span) []byte {
	b, err := msgpack.Marshal(span)
	if err != nil {
		panic(err)
	}
	return b
}
