package memqueue

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vmihailenco/taskq/v4"
	"github.com/vmihailenco/taskq/v4/backend"
	"github.com/vmihailenco/taskq/v4/backend/jobutil"
)

type scheduler struct {
	timerLock sync.Mutex
	timerMap  map[*taskq.Job]*time.Timer
}

func (q *scheduler) Schedule(msg *taskq.Job, fn func()) {
	q.timerLock.Lock()
	defer q.timerLock.Unlock()

	timer := time.AfterFunc(msg.Delay, func() {
		// Remove our entry from the map
		q.timerLock.Lock()
		delete(q.timerMap, msg)
		q.timerLock.Unlock()

		fn()
	})

	if q.timerMap == nil {
		q.timerMap = make(map[*taskq.Job]*time.Timer)
	}
	q.timerMap[msg] = timer
}

func (q *scheduler) Remove(msg *taskq.Job) {
	q.timerLock.Lock()
	defer q.timerLock.Unlock()

	timer, ok := q.timerMap[msg]
	if ok {
		timer.Stop()
		delete(q.timerMap, msg)
	}
}

func (q *scheduler) Purge() int {
	q.timerLock.Lock()
	defer q.timerLock.Unlock()

	// Stop all delayed items
	for _, timer := range q.timerMap {
		timer.Stop()
	}

	n := len(q.timerMap)
	q.timerMap = nil

	return n
}

//------------------------------------------------------------------------------

const (
	stateRunning = 0
	stateClosing = 1
	stateClosed  = 2
)

type Queue struct {
	opt *taskq.QueueConfig

	sync    bool
	noDelay bool

	wg       sync.WaitGroup
	consumer *taskq.Consumer

	scheduler scheduler

	_state int32
}

var _ taskq.Queue = (*Queue)(nil)

func NewQueue(opt *taskq.QueueConfig) *Queue {
	opt.Init()

	q := &Queue{
		opt: opt,
	}

	q.consumer = taskq.NewConsumer(q)
	if err := q.consumer.Start(context.Background()); err != nil {
		panic(err)
	}

	return q
}

func (q *Queue) Name() string {
	return q.opt.Name
}

func (q *Queue) String() string {
	return fmt.Sprintf("queue=%q", q.Name())
}

func (q *Queue) Options() *taskq.QueueConfig {
	return q.opt
}

func (q *Queue) Consumer() taskq.QueueConsumer {
	return q.consumer
}

func (q *Queue) SetSync(sync bool) {
	q.sync = sync
}

func (q *Queue) SetNoDelay(noDelay bool) {
	q.noDelay = noDelay
}

// Close is like CloseTimeout with 30 seconds timeout.
func (q *Queue) Close() error {
	return q.CloseTimeout(30 * time.Second)
}

// CloseTimeout closes the queue waiting for pending messages to be processed.
func (q *Queue) CloseTimeout(timeout time.Duration) error {
	if !atomic.CompareAndSwapInt32(&q._state, stateRunning, stateClosing) {
		return fmt.Errorf("taskq: %s is already closed", q)
	}
	err := q.WaitTimeout(timeout)

	if !atomic.CompareAndSwapInt32(&q._state, stateClosing, stateClosed) {
		panic("not reached")
	}

	_ = q.consumer.StopTimeout(timeout)
	_ = q.Purge(context.Background())

	return err
}

func (q *Queue) WaitTimeout(timeout time.Duration) error {
	done := make(chan struct{}, 1)
	go func() {
		q.wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		return fmt.Errorf("taskq: %s: messages are not processed after %s", q.consumer, timeout)
	}

	return nil
}

func (q *Queue) Len(ctx context.Context) (int, error) {
	return q.consumer.Len(), nil
}

// Add adds message to the queue.
func (q *Queue) AddJob(ctx context.Context, msg *taskq.Job) error {
	if q.closed() {
		return fmt.Errorf("taskq: %s is closed", q)
	}
	if msg.TaskName == "" {
		return backend.ErrTaskNameRequired
	}
	if msg.Name != "" && q.isDuplicate(ctx, msg) {
		msg.Err = taskq.ErrDuplicate
		return nil
	}
	q.wg.Add(1)
	return q.enqueueJob(ctx, msg)
}

func (q *Queue) enqueueJob(ctx context.Context, msg *taskq.Job) error {
	if (q.noDelay || q.sync) && msg.Delay > 0 {
		msg.Delay = 0
	}
	msg.ReservedCount++

	if q.sync {
		return q.consumer.Process(ctx, msg)
	}

	if msg.Delay > 0 {
		q.scheduler.Schedule(msg, func() {
			// If the queue closed while we were waiting, just return
			if q.closed() {
				q.wg.Done()
				return
			}
			msg.Delay = 0
			_ = q.consumer.AddJob(ctx, msg)
		})
		return nil
	}
	return q.consumer.AddJob(ctx, msg)
}

func (q *Queue) ReserveN(ctx context.Context, _ int, _ time.Duration) ([]taskq.Job, error) {
	return nil, backend.ErrNotSupported
}

func (q *Queue) Release(ctx context.Context, msg *taskq.Job) error {
	// Shallow copy.
	clone := *msg
	clone.Err = nil
	return q.enqueueJob(ctx, &clone)
}

func (q *Queue) Delete(ctx context.Context, msg *taskq.Job) error {
	q.scheduler.Remove(msg)
	q.wg.Done()
	return nil
}

func (q *Queue) DeleteBatch(ctx context.Context, msgs []*taskq.Job) error {
	if len(msgs) == 0 {
		return errors.New("taskq: no messages to delete")
	}
	for _, msg := range msgs {
		if err := q.Delete(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

func (q *Queue) Purge(ctx context.Context) error {
	// Purge any messages already in the consumer
	err := q.consumer.Purge(ctx)

	numPurged := q.scheduler.Purge()
	for i := 0; i < numPurged; i++ {
		q.wg.Done()
	}

	return err
}

func (q *Queue) closed() bool {
	return atomic.LoadInt32(&q._state) == stateClosed
}

func (q *Queue) isDuplicate(ctx context.Context, job *taskq.Job) bool {
	return q.opt.Storage.Exists(ctx, jobutil.FullJobName(q, job))
}
