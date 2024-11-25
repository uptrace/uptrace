package tracing

import (
	"time"

	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/vmihailenco/msgpack/v5"
)

type BaseData struct {
	Type      string `ch:",lc"`
	ProjectID uint32
	TraceID   idgen.TraceID
	ID        idgen.SpanID
	ParentID  idgen.SpanID
	Time      time.Time `ch:"type:DateTime64(6)"`
	Data      []byte
}

func (data *BaseData) InitFromSpan(span *Span) {
	data.Type = span.Type
	data.ProjectID = span.ProjectID
	data.TraceID = span.TraceID
	data.ID = span.ID
	data.ParentID = span.ParentID
	data.Time = span.Time
	data.Data = marshalSpanData(span)
}

func marshalSpanData(span *Span) []byte {
	b, err := msgpack.Marshal(span)
	if err != nil {
		panic(err)
	}
	return b
}
