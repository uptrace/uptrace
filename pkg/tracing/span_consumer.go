package tracing

import (
	"context"
)

type SpanConsumer struct {
	*BaseRecorder[*SpanIndex, *SpanData]
}

func (c *SpanConsumer) processItem(ctx context.Context, span *Span) {
	index := new(SpanIndex)
	index.init(span)
	c.indexSlice = append(c.indexSlice, index)

	data := new(SpanData)
	data.init(span)
	c.dataSlice = append(c.dataSlice, data)
}
