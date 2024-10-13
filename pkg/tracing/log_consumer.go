package tracing

import (
	"context"
)

type LogConsumer struct {
	*BaseRecorder[*LogIndex, *LogData]
}

func (c *LogConsumer) processItem(ctx context.Context, span *Span) {
	index := new(LogIndex)
	index.init(span)
	c.indexSlice = append(c.indexSlice, index)

	data := new(LogData)
	data.init(span)
	c.dataSlice = append(c.dataSlice, data)
}
