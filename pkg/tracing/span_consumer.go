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
	// fmt.Printf("Base Recorder: %#v\n", c.br)
	// fmt.Println("successful processing span item........................")
	// fmt.Printf("len span indexes: %d len span data: %d\n", len(c.br.indexSlice), len(c.br.dataSlice))
}
