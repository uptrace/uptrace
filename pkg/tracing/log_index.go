package tracing

import (
	"github.com/uptrace/go-clickhouse/ch"
)

type LogIndex struct {
	ch.CHModel `ch:"table:logs_index,insert:logs_index_buffer,alias:s"`

	*BaseIndex
}

func NewLogIndex(base *BaseIndex) *LogIndex {
	return &LogIndex{BaseIndex: base}
}

func (li *LogIndex) TableName() string {
	return "logs_index"
}

func (li *LogIndex) GetBaseIndex() *BaseIndex {
	if li.BaseIndex == nil {
		return new(BaseIndex)
	}
	return li.BaseIndex
}
