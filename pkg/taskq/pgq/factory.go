package pgq

import (
	"github.com/uptrace/bun"
	"github.com/vmihailenco/taskq/v4"
	"github.com/vmihailenco/taskq/v4/backend/base"
)

type factory struct {
	base.Factory

	db *bun.DB
}

var _ taskq.Factory = (*factory)(nil)

func NewFactory(db *bun.DB) taskq.Factory {
	return &factory{
		db: db,
	}
}

func (f *factory) RegisterQueue(opt *taskq.QueueConfig) taskq.Queue {
	q := NewQueue(f.db, opt)
	if err := f.Factory.Register(q); err != nil {
		panic(err)
	}
	return q
}
