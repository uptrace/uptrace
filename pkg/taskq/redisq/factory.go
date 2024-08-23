package redisq

import (
	"context"

	"github.com/vmihailenco/taskq/v4"
	"github.com/vmihailenco/taskq/v4/backend/base"
)

type factory struct {
	base base.Factory
}

var _ taskq.Factory = (*factory)(nil)

func NewFactory() taskq.Factory {
	return &factory{}
}

func (f *factory) RegisterQueue(opt *taskq.QueueConfig) taskq.Queue {
	q := NewQueue(opt)
	if err := f.base.Register(q); err != nil {
		panic(err)
	}
	return q
}

func (f *factory) Range(fn func(taskq.Queue) bool) {
	f.base.Range(fn)
}

func (f *factory) StartConsumers(ctx context.Context) error {
	return f.base.StartConsumers(ctx)
}

func (f *factory) StopConsumers() error {
	return f.base.StopConsumers()
}

func (f *factory) Close() error {
	return f.base.Close()
}
