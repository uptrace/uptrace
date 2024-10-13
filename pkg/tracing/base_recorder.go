package tracing

import (
	"context"
)

type IndexRecord interface {
	init(*Span)
}

var (
	_ IndexRecord = (*SpanIndex)(nil)
	_ IndexRecord = (*LogIndex)(nil)
)

type DataRecord interface {
	init(*Span)
}

var (
	_ DataRecord = (*SpanData)(nil)
	_ DataRecord = (*LogData)(nil)
)

type processItemFunc func(ctx context.Context, span *Span)

type BaseRecorder[I IndexRecord, D DataRecord] struct {

	indexSlice []I
	dataSlice  []D

	processItem processItemFunc
}

// func NewBaseRecorder[I IndexRecord, D DataRecord](insertSize int, processItem processItemFunc) *BaseRecorder[I, D] {
// 	return &BaseRecorder[I, D]{
// 		indexSlice:  make([]I, 0, insertSize),
// 		dataSlice:   make([]D, 0, insertSize),
// 		processItem: processItem,
// 	}
// }

