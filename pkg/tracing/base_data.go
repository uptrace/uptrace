package tracing

import (
	"context"
	"database/sql"
	"time"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/vmihailenco/msgpack/v5"
)

type dater interface {
	*SpanData | *LogData
	GetBaseData() *BaseData
	Decode(span *Span) error
	FilledSpan() (*Span, error)
}

type BaseDater[T dater] struct {
	Type T
}

func NewBaseDater[T dater](dater T) *BaseDater[T] {
	return &BaseDater[T]{Type: dater}
}

type BaseData struct {
	Type      string `ch:",lc"`
	ProjectID uint32
	TraceID   idgen.TraceID
	ID        idgen.SpanID
	ParentID  idgen.SpanID
	Time      time.Time `ch:"type:DateTime64(6)"`
	Data      []byte
}

func (b *BaseDater[T]) initData(span *Span) {
	data := b.Type.GetBaseData()
	data.Type = span.Type
	data.ProjectID = span.ProjectID
	data.TraceID = span.TraceID
	data.ID = span.ID
	data.ParentID = span.ParentID
	data.Time = span.Time
	data.Data = marshalSpanData(span)
}

func SelectSpan[T dater](
	ctx context.Context,
	app *bunapp.App,
	projectID uint32,
	traceID idgen.TraceID,
	spanID idgen.SpanID,
) (*Span, error) {
	var spans []T

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
		base := sd.GetBaseData()
		if base.ID == spanID {
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

func SelectTraceSpans[T dater](
	ctx context.Context, app *bunapp.App, traceID idgen.TraceID,
) ([]*Span, bool, error) {
	const limit = 10000

	var data []T

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
