package tracing

import (
	"context"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

type SpanData struct {
	ch.CHModel `ch:"table:spans_data_buffer,alias:s"`

	TraceID   uuid.UUID
	ID        uint64
	ParentID  uint64
	ProjectID uint32
	Time      time.Time
	Data      []byte
}

func (sd *SpanData) Decode(span *Span) error {
	if err := msgpack.Unmarshal(sd.Data, span); err != nil {
		return err
	}

	span.TraceID = sd.TraceID
	span.ID = sd.ID
	span.ParentID = sd.ParentID
	span.ProjectID = sd.ProjectID

	return nil
}

func initSpanData(data *SpanData, span *Span) {
	data.TraceID = span.TraceID
	data.ID = span.ID
	data.ParentID = span.ParentID
	data.ProjectID = span.ProjectID
	data.Time = span.Time
	data.Data = marshalSpanData(span)
}

func SelectSpan(ctx context.Context, app *bunapp.App, span *Span) error {
	var data SpanData

	q := app.CH.NewSelect().
		ColumnExpr("*").
		Model(&data).
		ModelTableExpr("?", app.DistTable("spans_data_buffer")).
		Where("trace_id = ?", span.TraceID).
		Limit(1)

	if span.ID != 0 {
		q = q.Where("id = ?", span.ID)
	}

	if err := q.Scan(ctx); err != nil {
		return err
	}

	return data.Decode(span)
}

// TODO: add project id filtering
func SelectTraceSpans(ctx context.Context, app *bunapp.App, traceID uuid.UUID) ([]*Span, error) {
	var data []SpanData

	if err := app.CH.NewSelect().
		Model(&data).
		ModelTableExpr("?", app.DistTable("spans_data_buffer")).
		Column("data").
		Where("trace_id = ?", traceID).
		Limit(10000).
		Scan(ctx); err != nil {
		return nil, err
	}

	spans := make([]*Span, len(data))

	for i := range spans {
		span := new(Span)
		spans[i] = span
		if err := data[i].Decode(span); err != nil {
			return nil, err
		}
	}

	return spans, nil
}

type SpanDataMsgpack struct {
	*Span

	TraceID   struct{} `msgpack:"-"`
	ID        struct{} `msgpack:"-"`
	ParentID  struct{} `msgpack:"-"`
	ProjectID struct{} `msgpack:"-"`
}

func marshalSpanData(span *Span) []byte {
	b, err := msgpack.Marshal(SpanDataMsgpack{Span: span})
	if err != nil {
		panic(err)
	}
	return b
}
