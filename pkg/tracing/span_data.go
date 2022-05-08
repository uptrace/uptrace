package tracing

import (
	"context"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/uuid"
)

type SpanData struct {
	ch.CHModel `ch:"table:spans_data_buffer,alias:s"`

	TraceID  uuid.UUID
	ID       uint64
	ParentID uint64
	Time     time.Time
	Data     []byte
}

func newSpanData(data *SpanData, span *Span) {
	data.TraceID = span.TraceID
	data.ID = span.ID
	data.ParentID = span.ParentID
	data.Time = span.Time
	data.Data = marshalSpan(span)
}

func SelectSpan(ctx context.Context, app *bunapp.App, span *Span) error {
	var data []byte

	q := app.CH().NewSelect().
		Model((*SpanData)(nil)).
		Column("data").
		Where("trace_id = ?", span.TraceID).
		Limit(1)

	if span.ID != 0 {
		q = q.Where("id = ?", span.ID)
	}

	if err := q.Scan(ctx, &data); err != nil {
		return err
	}

	return unmarshalSpan(data, span)
}

// TODO: add project id filtering
func SelectTraceSpans(ctx context.Context, app *bunapp.App, traceID uuid.UUID) ([]*Span, error) {
	var data []SpanData

	if err := app.CH().NewSelect().
		Model(&data).
		Column("data").
		Where("trace_id = ?", traceID).
		Limit(10000).
		Scan(ctx); err != nil {
		return nil, err
	}

	spans := make([]*Span, len(data))

	for i, span := range spans {
		span = new(Span)
		spans[i] = span

		if err := unmarshalSpan(data[i].Data, span); err != nil {
			return nil, err
		}
	}

	return spans, nil
}
