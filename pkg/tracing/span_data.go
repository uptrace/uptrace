package tracing

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/vmihailenco/msgpack/v5"
)

type SpanData struct {
	ch.CHModel `ch:"table:spans_data_buffer,insert:spans_data_buffer,alias:s"`

	BaseData
}

func (sd *SpanData) FilledSpan() (*Span, error) {
	span := new(Span)
	if err := sd.Decode(span); err != nil {
		return nil, err
	}
	return span, nil
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

	span.Type = span.System
	if i := strings.IndexByte(span.Type, ':'); i >= 0 {
		span.Type = span.Type[:i]
	}

	return nil
}

func initSpanData(data *SpanData, span *Span) {
	data.InitFromSpan(span)
}

func SelectSpan(
	ctx context.Context,
	app *bunapp.App,
	projectID uint32,
	traceID idgen.TraceID,
	spanID idgen.SpanID,
) (*Span, error) {
	var spans []*SpanData

	selq := app.CH.NewSelect().
		ColumnExpr("project_id, trace_id, id, parent_id, time, data").
		Model(&spans).
		Where("project_id = ?", projectID).
		Where("trace_id = ?", traceID)

	if spanID != 0 {
		selq.Where("id = ? OR parent_id = ?", spanID, spanID)
	} else {
		selq.Where("id = 0").Limit(1)
	}

	if err := selq.Scan(ctx); err != nil {
		return nil, err
	}

	var found *Span

	for i := len(spans) - 1; i >= 0; i-- {
		sd := spans[i]
		if sd.ID == spanID {
			span, err := sd.FilledSpan()
			if err != nil {
				return nil, err
			}
			found = span
			spans = append(spans[:i], spans[i+1:]...)
			break
		}
	}

	if found == nil {
		return nil, sql.ErrNoRows
	}

	for _, sd := range spans {
		span, err := sd.FilledSpan()
		if err != nil {
			return nil, err
		}
		if span.IsEvent() {
			found.Events = append(found.Events, span.Event())
		}
	}

	return found, nil
}

func SelectTraceSpans(
	ctx context.Context, app *bunapp.App, traceID idgen.TraceID,
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
