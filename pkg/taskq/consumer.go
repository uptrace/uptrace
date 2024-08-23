package taskq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis_rate/v10"

	"github.com/vmihailenco/taskq/v4/backend"
)

const stopTimeout = 30 * time.Second

var ErrAsyncTask = errors.New("taskq: async task")

type Delayer interface {
	Delay() time.Duration
}

type ConsumerStats struct {
	NumWorker  uint32
	NumFetcher uint32

	BufferSize uint32
	Buffered   uint32

	InFlight  uint32
	Processed uint32
	Retries   uint32
	Fails     uint32
}

//------------------------------------------------------------------------------

const (
	stateInit = iota
	stateStarted
	stateStoppingFetchers
	stateStoppingWorkers
)

// Consumer reserves messages from the queue, processes them,
// and then either releases or deletes messages from the queue.
type Consumer struct {
	q   Queue
	opt *QueueConfig

	buffer  chan *Job // never closed
	limiter *limiter

	consecutiveNumErr uint32

	inFlight  uint32
	processed uint32
	fails     uint32
	retries   uint32

	hooks []ConsumerHook

	startStopMu sync.Mutex

	fetchersCtx  context.Context
	fetchersWG   sync.WaitGroup
	stopFetchers func()

	workersCtx  context.Context
	workersWG   sync.WaitGroup
	stopWorkers func()
}

// NewConsumer creates new Consumer for the queue using provided processing options.
func NewConsumer(q Queue) *Consumer {
	opt := q.Options()
	c := &Consumer{
		q:   q,
		opt: opt,

		buffer: make(chan *Job, opt.BufferSize),

		limiter: &limiter{
			bucket:  q.Name(),
			limiter: opt.RateLimiter,
			limit:   opt.RateLimit,
		},
	}
	return c
}

// StartConsumer creates new QueueConsumer and starts it.
func StartConsumer(ctx context.Context, q Queue) *Consumer {
	c := NewConsumer(q)
	if err := c.Start(ctx); err != nil {
		panic(err)
	}
	return c
}

// AddHook adds a hook into message processing.
func (c *Consumer) AddHook(hook ConsumerHook) {
	c.hooks = append(c.hooks, hook)
}

func (c *Consumer) Queue() Queue {
	return c.q
}

func (c *Consumer) Options() *QueueConfig {
	return c.opt
}

func (c *Consumer) Len() int {
	return len(c.buffer)
}

// Stats returns processor stats.
func (c *Consumer) Stats() *ConsumerStats {
	return &ConsumerStats{
		BufferSize: uint32(cap(c.buffer)),
		Buffered:   uint32(len(c.buffer)),

		InFlight:  atomic.LoadUint32(&c.inFlight),
		Processed: atomic.LoadUint32(&c.processed),
		Retries:   atomic.LoadUint32(&c.retries),
		Fails:     atomic.LoadUint32(&c.fails),
	}
}

func (c *Consumer) AddJob(ctx context.Context, job *Job) error {
	_, _ = c.limiter.Reserve(ctx, 1)
	c.buffer <- job
	return nil
}

// Start starts consuming messages in the queue.
func (c *Consumer) Start(ctx context.Context) error {
	c.startStopMu.Lock()
	defer c.startStopMu.Unlock()

	if c.fetchersCtx != nil || c.workersCtx != nil {
		return nil
	}

	c.workersCtx, c.stopWorkers = context.WithCancel(ctx)
	for i := 0; i < c.opt.NumWorker; i++ {
		i := i
		c.workersWG.Add(1)
		go func() {
			defer c.workersWG.Done()
			c.worker(c.workersCtx, i)
		}()
	}

	c.fetchersCtx, c.stopFetchers = context.WithCancel(ctx)
	for i := 0; i < c.opt.NumFetcher; i++ {
		i := i
		c.fetchersWG.Add(1)
		go func() {
			defer c.fetchersWG.Done()
			c.fetcher(c.fetchersCtx, i)
		}()
	}

	return nil
}

// Stop is StopTimeout with 30 seconds timeout.
func (c *Consumer) Stop() error {
	return c.StopTimeout(stopTimeout)
}

// StopTimeout waits workers for timeout duration to finish processing current
// messages and stops workers.
func (c *Consumer) StopTimeout(timeout time.Duration) error {
	c.startStopMu.Lock()
	defer c.startStopMu.Unlock()

	if c.fetchersCtx == nil || c.workersCtx == nil {
		return nil
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	done := make(chan struct{}, 1)

	go func() {
		c.fetchersWG.Wait()
		done <- struct{}{}
	}()

	var firstErr error

	c.stopFetchers()
	select {
	case <-done:
	case <-timer.C:
		if firstErr == nil {
			firstErr = fmt.Errorf("taskq: %s: fetchers are not stopped after %s", c, timeout)
		}
	}

	go func() {
		c.workersWG.Wait()
		done <- struct{}{}
	}()

	c.stopWorkers()
	select {
	case <-done:
	case <-timer.C:
		if firstErr == nil {
			firstErr = fmt.Errorf("taskq: %s: workers are not stopped after %s", c, timeout)
		}
	}

	c.fetchersCtx = nil
	c.workersCtx = nil

	return firstErr
}

func (c *Consumer) paused() time.Duration {
	if c.opt.PauseErrorsThreshold == 0 ||
		atomic.LoadUint32(&c.consecutiveNumErr) < uint32(c.opt.PauseErrorsThreshold) {
		return 0
	}
	return time.Minute
}

// ProcessAll starts workers to process messages in the queue and then stops
// them when all messages are processed.
func (c *Consumer) ProcessAll(ctx context.Context) error {
	if err := c.Start(ctx); err != nil {
		return err
	}

	var prev *ConsumerStats
	var noWork int
	for {
		st := c.Stats()
		if prev != nil &&
			st.Buffered == 0 &&
			st.InFlight == 0 &&
			st.Processed == prev.Processed {
			noWork++
			if noWork == 2 {
				break
			}
		} else {
			noWork = 0
		}
		prev = st
		time.Sleep(time.Second)
	}

	return c.Stop()
}

// ProcessOne processes at most one message in the queue.
func (c *Consumer) ProcessOne(ctx context.Context) error {
	job, err := c.reserveOne(ctx)
	if err != nil {
		return err
	}
	return c.Process(ctx, job)
}

func (c *Consumer) reserveOne(ctx context.Context) (*Job, error) {
	select {
	case job := <-c.buffer:
		return job, nil
	default:
	}

	jobs, err := c.q.ReserveN(ctx, 1, c.opt.WaitTimeout)
	if err != nil && err != backend.ErrNotSupported {
		return nil, err
	}

	if len(jobs) == 0 {
		return nil, errors.New("taskq: queue is empty")
	}
	if len(jobs) != 1 {
		return nil, fmt.Errorf("taskq: queue returned %d messages", len(jobs))
	}

	return &jobs[0], nil
}

func (c *Consumer) fetcher(ctx context.Context, fetcherID int) {
	for {
		if pauseTime := c.paused(); pauseTime > 0 {
			backend.Warn("consumer is automatically paused", "pause_time", pauseTime)
			time.Sleep(pauseTime)
			c.resetPause()
			continue
		}

		switch err := c.reserveJobs(ctx); err {
		case nil:
			// nothing
		case backend.ErrNotSupported, context.Canceled:
			return
		case context.DeadlineExceeded:
			backend.Error(err, "reserveJobs failed (exiting)")
		default:
			backoff := time.Second
			backend.Error(err, "reserveJobs failed (will retry)", "backoff", backoff)
			sleep(ctx, backoff)
		}
	}
}

func (c *Consumer) reserveJobs(ctx context.Context) error {
	size, err := c.limiter.Reserve(ctx, c.opt.ReservationSize)
	if err != nil {
		return err
	}

	jobs, err := c.q.ReserveN(ctx, size, c.opt.WaitTimeout)
	if err != nil {
		return err
	}

	if d := size - len(jobs); d > 0 {
		c.limiter.Cancel(d)
	}

	for i := range jobs {
		job := &jobs[i]
		select {
		case c.buffer <- job:
		case <-ctx.Done():
			for i := range jobs[i:] {
				_ = c.q.Release(ctx, &jobs[i])
			}
			return context.Canceled
		}
	}

	return nil
}

func (c *Consumer) worker(ctx context.Context, workerID int) {
	for {
		job := c.waitJob(ctx)
		if job == nil {
			return
		}
		_ = c.Process(backend.UndoContext(ctx), job)
	}
}

func (c *Consumer) waitJob(ctx context.Context) *Job {
	select {
	case job := <-c.buffer:
		return job
	default:
	}

	select {
	case job := <-c.buffer:
		return job
	case <-ctx.Done():
		return nil
	}
}

// Process is low-level API to process message bypassing the internal queue.
func (c *Consumer) Process(ctx context.Context, job *Job) error {
	atomic.AddUint32(&c.inFlight, 1)

	if job.Delay > 0 {
		if err := c.q.AddJob(ctx, job); err != nil {
			return err
		}
		return nil
	}

	if job.Err != nil {
		job.Delay = -1
		c.Put(ctx, job)
		return job.Err
	}

	ctx, evt := c.beforeProcessJob(ctx, job)
	job.evt = evt

	jobErr := c.opt.Handler.HandleJob(ctx, job)
	if jobErr == ErrAsyncTask {
		return ErrAsyncTask
	}

	job.Err = jobErr
	c.Put(ctx, job)

	return job.Err
}

func (c *Consumer) Put(ctx context.Context, job *Job) {
	c.afterProcessJob(ctx, job)

	if job.Err == nil {
		c.resetPause()
		atomic.AddUint32(&c.processed, 1)
		c.delete(ctx, job)
		return
	}

	atomic.AddUint32(&c.consecutiveNumErr, 1)
	if job.Delay <= 0 {
		atomic.AddUint32(&c.fails, 1)
		c.delete(ctx, job)
		return
	}

	atomic.AddUint32(&c.retries, 1)
	c.release(ctx, job)
}

func (c *Consumer) release(ctx context.Context, job *Job) {
	if job.Err != nil {
		backend.Error(job.Err, "job failed (will retry)",
			"task_name", job.TaskName,
			"reserved_count", job.ReservedCount,
			"delay", job.Delay)
	}

	if err := c.q.Release(ctx, job); err != nil {
		backend.Error(err, "Release failed",
			"task_name", job.TaskName)
	}
	atomic.AddUint32(&c.inFlight, ^uint32(0))
}

func (c *Consumer) delete(ctx context.Context, job *Job) {
	if job.Err != nil {
		backend.Error(job.Err, "job failed (dropping)",
			"task_name", job.TaskName,
			"reserved_count", job.ReservedCount)

		if err := c.opt.Handler.HandleJob(ctx, job); err != nil {
			backend.Error(err, "fallback handler failed (dropping)",
				"task_name", job.TaskName)
		}
	}

	if err := c.q.Delete(ctx, job); err != nil {
		backend.Error(err, "Delete failed",
			"task_name", job.TaskName)
	}
	atomic.AddUint32(&c.inFlight, ^uint32(0))
}

// Purge discards messages from the internal queue.
func (c *Consumer) Purge(ctx context.Context) error {
	for {
		select {
		case job := <-c.buffer:
			c.delete(ctx, job)
		default:
			return nil
		}
	}
}

type ProcessJobEvent struct {
	Job       *Job
	StartTime time.Time

	Stash map[interface{}]interface{}
}

type ConsumerHook interface {
	BeforeProcessJob(context.Context, *ProcessJobEvent) context.Context
	AfterProcessJob(context.Context, *ProcessJobEvent)
}

func (c *Consumer) beforeProcessJob(
	ctx context.Context, job *Job,
) (context.Context, *ProcessJobEvent) {
	if len(c.hooks) == 0 {
		return ctx, nil
	}

	evt := &ProcessJobEvent{
		Job:       job,
		StartTime: time.Now(),
	}

	for _, hook := range c.hooks {
		ctx = hook.BeforeProcessJob(ctx, evt)
	}

	return ctx, evt
}

func (c *Consumer) afterProcessJob(ctx context.Context, job *Job) {
	if job.evt == nil {
		return
	}

	for i := len(c.hooks) - 1; i >= 0; i-- {
		c.hooks[i].AfterProcessJob(ctx, job.evt)
	}
}

func (c *Consumer) resetPause() {
	atomic.StoreUint32(&c.consecutiveNumErr, 0)
}

func (c *Consumer) String() string {
	inFlight := atomic.LoadUint32(&c.inFlight)
	processed := atomic.LoadUint32(&c.processed)
	retries := atomic.LoadUint32(&c.retries)
	fails := atomic.LoadUint32(&c.fails)

	return fmt.Sprintf(
		"%s %d/%d/%d %d/%d/%d",
		c.q.Name(),
		inFlight, len(c.buffer), cap(c.buffer),
		processed, retries, fails)
}

//------------------------------------------------------------------------------

type limiter struct {
	bucket  string
	limiter *redis_rate.Limiter
	limit   redis_rate.Limit

	allowedCount uint32 // atomic
	cancelled    uint32 // atomic
}

func (l *limiter) Reserve(ctx context.Context, max int) (int, error) {
	if l.limiter == nil || l.limit.IsZero() {
		return max, nil
	}

	for {
		cancelled := atomic.LoadUint32(&l.cancelled)
		if cancelled == 0 {
			break
		}

		if cancelled >= uint32(max) {
			if atomic.CompareAndSwapUint32(&l.cancelled, cancelled, uint32(max)-1) {
				return max, nil
			}
			continue
		}

		if atomic.CompareAndSwapUint32(&l.cancelled, cancelled, uint32(cancelled)-1) {
			return int(cancelled), nil
		}
	}

	for {
		res, err := l.limiter.AllowAtMost(ctx, l.bucket, l.limit, max)
		if err != nil {
			if err == context.Canceled {
				return 0, err
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if res.Allowed > 0 {
			atomic.AddUint32(&l.allowedCount, 1)
			return res.Allowed, nil
		}

		atomic.StoreUint32(&l.allowedCount, 0)
		sleep(ctx, res.RetryAfter)
	}
}

func (l *limiter) Cancel(n int) {
	if l.limiter == nil {
		return
	}
	atomic.AddUint32(&l.cancelled, uint32(n))
}

func (l *limiter) Limited() bool {
	return l.limiter != nil && atomic.LoadUint32(&l.allowedCount) < 3
}

//------------------------------------------------------------------------------

func exponentialBackoff(min, max time.Duration, retry int) time.Duration {
	var d time.Duration
	if retry > 0 {
		d = min << uint(retry-1)
	}
	if d < min {
		return min
	}
	if d > max {
		return max
	}
	return d
}

func sleep(ctx context.Context, d time.Duration) error {
	done := ctx.Done()
	if done == nil {
		time.Sleep(d)
		return nil
	}

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-t.C:
		return nil
	case <-done:
		return ctx.Err()
	}
}
