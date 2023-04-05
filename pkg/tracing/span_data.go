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
	ch.CHModel `ch:"table:spans_data_buffer,alias:s"`

	Type      string `ch:",lc"`
	ProjectID uint32
	TraceID   uuid.UUID
	ID        uint64
	ParentID  uint64
	Time      time.Time
	Data      []byte
}

func (sd *SpanData) Decode(span *Span) error {
	span.ProjectID = sd.ProjectID
	span.TraceID = sd.TraceID
	span.ID = sd.ID
	span.ParentID = sd.ParentID

	if err := msgpack.Unmarshal(sd.Data, span); err != nil {
		return fmt.Errorf("msgpack.Unmarshal failed: %w", err)
	}

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

	q := app.CH.NewSelect().
		ColumnExpr("project_id, trace_id, id, parent_id, data").
		Model(&data).
		ModelTableExpr("?", app.DistTable("spans_data_buffer")).
		Where("trace_id = ?", span.TraceID).
		Limit(1)

	if span.ProjectID != 0 {
		q = q.Where("project_id = ?", span.ProjectID)
	}
	if span.ID != 0 {
		q = q.Where("id = ?", span.ID)
	}

	if err := q.Scan(ctx); err != nil {
		return err
	}

	return data.Decode(span)
}

func SelectTraceSpans(ctx context.Context, app *bunapp.App, traceID uuid.UUID) ([]*Span, error) {
	var data []SpanData

	if err := app.CH.NewSelect().
		ColumnExpr("project_id, trace_id, id, parent_id").
		Model(&data).
		ModelTableExpr("?", app.DistTable("spans_data_buffer")).
		Column("data").
		Where("trace_id = ?", traceID).
		OrderExpr("time ASC").
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

func marshalSpanData(span *Span) []byte {
	b, err := msgpack.Marshal(span)
	if err != nil {
		panic(err)
	}
	return b
}
