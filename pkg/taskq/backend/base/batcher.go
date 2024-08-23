package base

import (
	"context"
	"sync"
	"time"

	"github.com/vmihailenco/taskq/v4"
)

type BatcherOptions struct {
	Handler     func([]*taskq.Job) error
	ShouldBatch func([]*taskq.Job, *taskq.Job) bool

	Timeout time.Duration
}

func (opt *BatcherOptions) init() {
	if opt.Timeout == 0 {
		opt.Timeout = 3 * time.Second
	}
}

// Batcher collects messages for later batch processing.
type Batcher struct {
	consumer taskq.QueueConsumer
	opt      *BatcherOptions

	timer *time.Timer

	mu     sync.Mutex
	batch  []*taskq.Job
	closed bool
}

func NewBatcher(consumer taskq.QueueConsumer, opt *BatcherOptions) *Batcher {
	opt.init()
	b := Batcher{
		consumer: consumer,
		opt:      opt,
	}
	b.timer = time.AfterFunc(time.Minute, b.onTimeout)
	b.timer.Stop()
	return &b
}

func (b *Batcher) flush() {
	if len(b.batch) > 0 {
		b.process(b.batch)
		b.batch = nil
	}
}

func (b *Batcher) Add(msg *taskq.Job) error {
	var batch []*taskq.Job

	b.mu.Lock()

	if b.closed {
		if len(b.batch) > 0 {
			panic("not reached")
		}
		batch = []*taskq.Job{msg}
	} else {
		if len(b.batch) == 0 {
			b.stopTimer()
			b.timer.Reset(b.opt.Timeout)
		}

		if b.opt.ShouldBatch(b.batch, msg) {
			b.batch = append(b.batch, msg)
		} else {
			batch = b.batch
			b.batch = []*taskq.Job{msg}
		}
	}

	b.mu.Unlock()

	if len(batch) > 0 {
		b.process(batch)
	}

	return taskq.ErrAsyncTask
}

func (b *Batcher) stopTimer() {
	if !b.timer.Stop() {
		select {
		case <-b.timer.C:
		default:
		}
	}
}

func (b *Batcher) process(batch []*taskq.Job) {
	err := b.opt.Handler(batch)
	for _, msg := range batch {
		if msg.Err == nil {
			msg.Err = err
		}
		b.consumer.Put(context.TODO(), msg)
	}
}

func (b *Batcher) onTimeout() {
	b.mu.Lock()
	b.flush()
	b.mu.Unlock()
}

func (b *Batcher) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}
	b.closed = true

	b.stopTimer()
	b.flush()

	return nil
}
