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


