package tracing

import (
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/idgen"
)

type LogData struct {
	ch.CHModel `ch:"table:logs_data_buffer,insert:logs_data_buffer,alias:s"`

	ProjectID uint32
	TraceID   idgen.TraceID
	ID        idgen.SpanID
	ParentID  idgen.SpanID
	Time      time.Time `ch:"type:DateTime64(6)"`
	Data      []byte
}

func (data *LogData) init(span *Span) {
	data.ProjectID = span.ProjectID
	data.TraceID = span.TraceID
	data.ID = span.ID
	data.ParentID = span.ParentID
	data.Time = span.Time
	data.Data = marshalSpanData(span)
}
