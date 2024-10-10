package tracing

import (
	"fmt"
	"strings"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/vmihailenco/msgpack/v5"
)

type LogData struct {
	ch.CHModel `ch:"table:logs_data_buffer,insert:logs_data_buffer,alias:s"`

	*BaseData
}

func NewLogData(base *BaseData) *LogData {
	return &LogData{BaseData: base}
}

func (ld *LogData) TableName() string {
	return "logs_data"
}

func (ld *LogData) GetBaseData() *BaseData {
	if ld.BaseData == nil {
		return new(BaseData)
	}
	return ld.BaseData
}

func (ld *LogData) Decode(span *Span) error {
	if err := msgpack.Unmarshal(ld.Data, span); err != nil {
		return fmt.Errorf("msgpack.Unmarshal failed: %w", err)
	}

	span.ProjectID = ld.ProjectID
	span.TraceID = ld.TraceID
	span.ID = ld.ID
	span.ParentID = ld.ParentID
	span.Time = ld.Time

	span.Type = span.System
	if i := strings.IndexByte(span.Type, ':'); i >= 0 {
		span.Type = span.Type[:i]
	}

	return nil
}

func (ld *LogData) FilledSpan() (*Span, error) {
	span := new(Span)
	if err := ld.Decode(span); err != nil {
		return nil, err
	}
	return span, nil
}
