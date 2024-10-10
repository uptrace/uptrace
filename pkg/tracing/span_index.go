package tracing

import (

	"github.com/uptrace/go-clickhouse/ch"
)

type SpanIndex struct {
	ch.CHModel `ch:"table:spans_index,insert:spans_index_buffer,alias:s"`

	*BaseIndex
}

func NewSpanIndex(base *BaseIndex) *SpanIndex {
	return &SpanIndex{BaseIndex: base}
}

func (si *SpanIndex) TableName() string {
	return "spans_index"
}

func (si *SpanIndex) GetBaseIndex() *BaseIndex {
	if si.BaseIndex == nil {
		return new(BaseIndex)
	}
	return si.BaseIndex
}

