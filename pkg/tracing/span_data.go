package tracing

import (
	"fmt"
	"strings"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/vmihailenco/msgpack/v5"
)

type SpanData struct {
	ch.CHModel `ch:"table:spans_data_buffer,insert:spans_data_buffer,alias:s"`

	*BaseData
}

func NewSpanData(base *BaseData) *SpanData {
	return &SpanData{BaseData: base}
}

func (sd *SpanData) TableName() string {
	return "spans_data"
}

func (sd *SpanData) GetBaseData() *BaseData {
	if sd.BaseData == nil {
		return new(BaseData)
	}
	return sd.BaseData
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

func (sd *SpanData) FilledSpan() (*Span, error) {
	span := new(Span)
	if err := sd.Decode(span); err != nil {
		return nil, err
	}
	return span, nil
}
