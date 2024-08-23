package taskqtest

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/vmihailenco/taskq/v4"
)

const (
	waitTimeout = time.Second
	testTimeout = 30 * time.Second
)

var (
	ringOnce sync.Once
	ring     *redis.Ring
)

func redisRing() *redis.Ring {
	ringOnce.Do(func() {
		ring = redis.NewRing(&redis.RingOptions{
			Addrs: map[string]string{"0": ":6379"},
		})
	})
	_ = ring.FlushDB(context.TODO()).Err()
	return ring
}

func TestConsumer(t *testing.T, factory taskq.Factory, opt *taskq.QueueConfig) {
	ctx := context.Background()
	opt.WaitTimeout = waitTimeout
	opt.Redis = redisRing()

	q := factory.RegisterQueue(opt)
	defer q.Close()
	purge(t, q)

	ch := make(chan time.Time)
	task := taskq.RegisterTask(nextTaskID(), &taskq.TaskConfig{
		Handler: func(hello, world string) error {
			if hello != "hello" {
				t.Fatalf("got %s, wanted hello", hello)
			}
			if world != "world" {
				t.Fatalf("got %s, wanted world", world)
			}
			ch <- time.Now()
			return nil
		},
	})

	err := q.AddJob(ctx, task.NewJob("hello", "world"))
	if err != nil {
		t.Fatal(err)
	}

	p := q.Consumer()
	if err := p.Start(ctx); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ch:
	case <-time.After(testTimeout):
		t.Fatalf("message was not processed")
	}

	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestUnknownTask(t *testing.T, factory taskq.Factory, opt *taskq.QueueConfig) {
	ctx := context.Background()
	opt.WaitTimeout = waitTimeout
	opt.Redis = redisRing()

	q := factory.RegisterQueue(opt)
	defer q.Close()
	purge(t, q)

	_ = taskq.RegisterTask(nextTaskID(), &taskq.TaskConfig{
		Handler: func() {},
	})

	taskq.SetUnknownTaskConfig(&taskq.TaskConfig{
		RetryLimit: 1,
	})

	msg := taskq.NewJob()
	msg.TaskName = "unknown"
	err := q.AddJob(ctx, msg)
	if err != nil {
		t.Fatal(err)
	}

	p := q.Consumer()
	if err := p.Start(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestFallback(t *testing.T, factory taskq.Factory, opt *taskq.QueueConfig) {
	ctx := context.Background()
	opt.WaitTimeout = waitTimeout
	opt.Redis = redisRing()

	q := factory.RegisterQueue(opt)
	defer q.Close()
	purge(t, q)

	ch := make(chan time.Time)
	task := taskq.RegisterTask(nextTaskID(), &taskq.TaskConfig{
		Handler: func() error {
			return errors.New("fake error")
		},
		FallbackHandler: func(hello, world string) error {
			if hello != "hello" {
				t.Fatalf("got %s, wanted hello", hello)
			}
			if world != "world" {
				t.Fatalf("got %s, wanted world", world)
			}
			ch <- time.Now()
			return nil
		},
		RetryLimit: 1,
	})

	err := q.AddJob(ctx, task.NewJob("hello", "world"))
	if err != nil {
		t.Fatal(err)
	}

	p := q.Consumer()
	p.Start(ctx)

	select {
	case <-ch:
	case <-time.After(testTimeout):
		t.Fatalf("message was not processed")
	}

	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestDelay(t *testing.T, factory taskq.Factory, opt *taskq.QueueConfig) {
	ctx := context.Background()
	opt.WaitTimeout = waitTimeout
	opt.Redis = redisRing()

	q := factory.RegisterQueue(opt)
	defer q.Close()
	purge(t, q)

	handlerCh := make(chan time.Time, 10)
	task := taskq.RegisterTask(nextTaskID(), &taskq.TaskConfig{
		Handler: func() {
			handlerCh <- time.Now()
		},
	})

	start := time.Now()

	msg := task.NewJob()
	msg.Delay = 5 * time.Second
	err := q.AddJob(ctx, msg)
	if err != nil {
		t.Fatal(err)
	}

	p := q.Consumer()
	p.Start(ctx)

	var tm time.Time
	select {
	case tm = <-handlerCh:
	case <-time.After(testTimeout):
		t.Fatalf("message was not processed")
	}

	sub := tm.Sub(start)
	if !durEqual(sub, msg.Delay) {
		t.Fatalf("message was delayed by %s, wanted %s", sub, msg.Delay)
	}

	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRetry(t *testing.T, factory taskq.Factory, opt *taskq.QueueConfig) {
	ctx := context.Background()
	opt.WaitTimeout = waitTimeout
	opt.Redis = redisRing()

	q := factory.RegisterQueue(opt)
	defer q.Close()
	purge(t, q)

	handlerCh := make(chan time.Time, 10)
	task := taskq.RegisterTask(nextTaskID(), &taskq.TaskConfig{
		Handler: func(hello, world string) error {
			if hello != "hello" {
				t.Fatalf("got %q, wanted hello", hello)
			}
			if world != "world" {
				t.Fatalf("got %q, wanted world", world)
			}
			handlerCh <- time.Now()
			return errors.New("fake error")
		},
		FallbackHandler: func(msg *taskq.Job) error {
			handlerCh <- time.Now()
			return nil
		},
		RetryLimit: 3,
		MinBackoff: time.Second,
	})

	err := q.AddJob(ctx, task.NewJob("hello", "world"))
	if err != nil {
		t.Fatal(err)
	}

	p := q.Consumer()
	p.Start(ctx)

	timings := []time.Duration{0, time.Second, 3 * time.Second, 3 * time.Second}
	testTimings(t, handlerCh, timings)

	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNamedJob(t *testing.T, factory taskq.Factory, opt *taskq.QueueConfig) {
	ctx := context.Background()
	opt.WaitTimeout = waitTimeout
	opt.Redis = redisRing()

	q := factory.RegisterQueue(opt)
	defer q.Close()
	purge(t, q)

	ch := make(chan time.Time, 10)
	task := taskq.RegisterTask(nextTaskID(), &taskq.TaskConfig{
		Handler: func(hello string) error {
			if hello != "world" {
				panic("hello != world")
			}
			ch <- time.Now()
			return nil
		},
	})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			msg := task.NewJob("world")
			msg.Name = "the-name"
			err := q.AddJob(ctx, msg)
			if err != nil {
				t.Fatal(err)
			}
		}()
	}
	wg.Wait()

	p := q.Consumer()
	p.Start(ctx)

	select {
	case <-ch:
	case <-time.After(testTimeout):
		t.Fatalf("message was not processed")
	}

	select {
	case <-ch:
		t.Fatalf("message was processed twice")
	default:
	}

	require.NoError(t, p.Stop())
	require.NoError(t, q.Close())
}

func TestCallOnce(t *testing.T, factory taskq.Factory, opt *taskq.QueueConfig) {
	ctx := context.Background()
	opt.WaitTimeout = waitTimeout
	opt.Redis = redisRing()

	q := factory.RegisterQueue(opt)
	defer q.Close()
	purge(t, q)

	ch := make(chan time.Time, 10)
	task := taskq.RegisterTask(nextTaskID(), &taskq.TaskConfig{
		Handler: func() {
			ch <- time.Now()
		},
	})

	go func() {
		for i := 0; i < 3; i++ {
			for j := 0; j < 10; j++ {
				msg := task.NewJob()
				msg.OnceInPeriod(500 * time.Millisecond)

				err := q.AddJob(ctx, msg)
				if err != nil {
					t.Fatal(err)
				}
			}
			time.Sleep(time.Second)
		}
	}()

	p := q.Consumer()
	if err := p.Start(ctx); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		select {
		case <-ch:
		case <-time.After(testTimeout):
			t.Fatalf("message was not processed")
		}
	}

	select {
	case <-ch:
		t.Fatalf("message was processed twice")
	case <-time.After(time.Second):
	}

	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestLen(t *testing.T, factory taskq.Factory, opt *taskq.QueueConfig) {
	const N = 10

	ctx := context.Background()
	opt.WaitTimeout = waitTimeout
	opt.Redis = redisRing()

	q := factory.RegisterQueue(opt)
	defer q.Close()
	purge(t, q)

	task := taskq.RegisterTask(nextTaskID(), &taskq.TaskConfig{
		Handler: func() {},
	})

	for i := 0; i < N; i++ {
		err := q.AddJob(ctx, task.NewJob())
		if err != nil {
			t.Fatal(err)
		}
	}

	eventually(func() error {
		n, err := q.Len(ctx)
		if err != nil {
			return err
		}

		if n != N {
			return fmt.Errorf("got %d messages, wanted %d", n, N)
		}
		return nil
	}, testTimeout)

	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRateLimit(t *testing.T, factory taskq.Factory, opt *taskq.QueueConfig) {
	ctx := context.Background()
	opt.WaitTimeout = waitTimeout
	opt.RateLimit = redis_rate.PerSecond(1)
	opt.Redis = redisRing()

	q := factory.RegisterQueue(opt)
	defer q.Close()
	purge(t, q)

	var count int64
	task := taskq.RegisterTask(nextTaskID(), &taskq.TaskConfig{
		Handler: func() {
			atomic.AddInt64(&count, 1)
		},
	})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := q.AddJob(ctx, task.NewJob())
			if err != nil {
				t.Fatal(err)
			}
		}()
	}
	wg.Wait()

	p := q.Consumer()
	p.Start(ctx)

	time.Sleep(5 * time.Second)

	if n := atomic.LoadInt64(&count); n-5 > 2 {
		t.Fatalf("processed %d messages, wanted 5", n)
	}

	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestErrorDelay(t *testing.T, factory taskq.Factory, opt *taskq.QueueConfig) {
	ctx := context.Background()
	opt.WaitTimeout = waitTimeout
	opt.Redis = redisRing()

	q := factory.RegisterQueue(opt)
	defer q.Close()
	purge(t, q)

	handlerCh := make(chan time.Time, 10)
	task := taskq.RegisterTask(nextTaskID(), &taskq.TaskConfig{
		Handler: func() error {
			handlerCh <- time.Now()
			return RateLimitError("fake error")
		},
		MinBackoff: time.Second,
		RetryLimit: 3,
	})

	err := q.AddJob(ctx, task.NewJob())
	if err != nil {
		t.Fatal(err)
	}

	p := q.Consumer()
	p.Start(ctx)

	timings := []time.Duration{0, 3 * time.Second, 3 * time.Second}
	testTimings(t, handlerCh, timings)

	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidCredentials(t *testing.T, factory taskq.Factory, opt *taskq.QueueConfig) {
	ctx := context.Background()
	opt.WaitTimeout = waitTimeout
	opt.Redis = redisRing()

	q := factory.RegisterQueue(opt)
	defer q.Close()

	ch := make(chan time.Time, 10)
	task := taskq.RegisterTask(nextTaskID(), &taskq.TaskConfig{
		Handler: func(s1, s2 string) {
			if s1 != "hello" {
				t.Fatalf("got %q, wanted hello", s1)
			}
			if s2 != "world" {
				t.Fatalf("got %q, wanted world", s1)
			}
			ch <- time.Now()
		},
	})

	err := q.AddJob(ctx, task.NewJob("hello", "world"))
	if err != nil {
		t.Fatal(err)
	}

	timings := []time.Duration{3 * time.Second}
	testTimings(t, ch, timings)

	err = q.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestBatchConsumer(
	t *testing.T, factory taskq.Factory, opt *taskq.QueueConfig, messageSize int,
) {
	const N = 16

	ctx := context.Background()
	opt.WaitTimeout = waitTimeout
	opt.Redis = redisRing()

	payload := make([]byte, messageSize)
	_, err := rand.Read(payload)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(N)

	opt.WaitTimeout = waitTimeout
	q := factory.RegisterQueue(opt)
	defer q.Close()
	purge(t, q)

	task := taskq.RegisterTask(nextTaskID(), &taskq.TaskConfig{
		Handler: func(s string) {
			defer wg.Done()
			if s != string(payload) {
				t.Fatalf("s != payload")
			}
		},
	})

	for i := 0; i < N; i++ {
		err := q.AddJob(ctx, task.NewJob(payload))
		if err != nil {
			t.Fatal(err)
		}
	}

	p := q.Consumer()
	if err := p.Start(ctx); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatalf("messages were not processed")
	}

	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
}

func durEqual(d1, d2 time.Duration) bool {
	return d1 >= d2 && d2-d1 < 3*time.Second
}

func testTimings(t *testing.T, ch chan time.Time, timings []time.Duration) {
	start := time.Now()
	for i, timing := range timings {
		var tm time.Time
		select {
		case tm = <-ch:
		case <-time.After(testTimeout):
			t.Fatalf("message is not processed after %s", 2*timing)
		}
		since := tm.Sub(start)
		if !durEqual(since, timing) {
			t.Fatalf("#%d: timing is %s, wanted %s", i+1, since, timing)
		}
	}
}

func purge(t *testing.T, q taskq.Queue) {
	err := q.Purge(context.TODO())
	if err == nil {
		return
	}

	task := taskq.RegisterTask("*", &taskq.TaskConfig{
		Handler: func() {},
	})

	consumer := taskq.NewConsumer(q)
	err = consumer.ProcessAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	taskq.Tasks.Unregister(task)
}

func eventually(fn func() error, timeout time.Duration) error {
	errCh := make(chan error)
	done := make(chan struct{})
	exit := make(chan struct{})

	go func() {
		for {
			err := fn()
			if err == nil {
				close(done)
				return
			}

			select {
			case errCh <- err:
			default:
			}

			select {
			case <-exit:
				return
			case <-time.After(timeout / 100):
			}
		}
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		close(exit)
		select {
		case err := <-errCh:
			return err
		default:
			return fmt.Errorf("timeout after %s", timeout)
		}
	}
}

var taskID int

func nextTaskID() string {
	id := strconv.Itoa(taskID)
	taskID++
	return id
}

func unixMs(tm time.Time) int64 {
	return tm.UnixNano() / int64(time.Millisecond)
}

//------------------------------------------------------------------------------

type RateLimitError string

func (e RateLimitError) Error() string {
	return string(e)
}

func (RateLimitError) Delay() time.Duration {
	return 3 * time.Second
}
