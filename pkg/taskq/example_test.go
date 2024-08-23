package taskq_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"

	"github.com/vmihailenco/taskq/memqueue/v4"
	"github.com/vmihailenco/taskq/v4"
)

func Example_retryOnError() {
	start := time.Now()
	q := memqueue.NewQueue(&taskq.QueueConfig{
		Name: "test",
	})
	task := taskq.RegisterTask("Example_retryOnError", &taskq.TaskConfig{
		Handler: func() error {
			fmt.Println("retried in", timeSince(start))
			return errors.New("fake error")
		},
		RetryLimit: 3,
		MinBackoff: time.Second,
	})

	ctx := context.Background()
	q.AddJob(ctx, task.NewJob())

	// Wait for all messages to be processed.
	_ = q.Close()

	// Output: retried in 0s
	// retried in 1s
	// retried in 3s
}

func Example_messageDelay() {
	start := time.Now()
	q := memqueue.NewQueue(&taskq.QueueConfig{
		Name: "test",
	})
	task := taskq.RegisterTask("Example_messageDelay", &taskq.TaskConfig{
		Handler: func() {
			fmt.Println("processed with delay", timeSince(start))
		},
	})

	ctx := context.Background()
	msg := task.NewJob()
	msg.Delay = time.Second
	_ = q.AddJob(ctx, msg)

	// Wait for all messages to be processed.
	_ = q.Close()

	// Output: processed with delay 1s
}

func Example_rateLimit() {
	start := time.Now()
	q := memqueue.NewQueue(&taskq.QueueConfig{
		Name:      "test",
		Redis:     redisRing(),
		RateLimit: redis_rate.PerSecond(1),
	})
	task := taskq.RegisterTask("Example_rateLimit", &taskq.TaskConfig{
		Handler: func() {},
	})

	const n = 5

	ctx := context.Background()
	for i := 0; i < n; i++ {
		_ = q.AddJob(ctx, task.NewJob())
	}

	// Wait for all messages to be processed.
	_ = q.Close()

	fmt.Printf("%d msg/s", timeSinceCeil(start)/time.Second/n)
	// Output: 1 msg/s
}

func Example_once() {
	q := memqueue.NewQueue(&taskq.QueueConfig{
		Name:      "test",
		Redis:     redisRing(),
		RateLimit: redis_rate.PerSecond(1),
	})
	task := taskq.RegisterTask("Example_once", &taskq.TaskConfig{
		Handler: func(name string) {
			fmt.Println("hello", name)
		},
	})

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		msg := task.NewJob("world")
		// Call once in a second.
		msg.OnceInPeriod(time.Second)

		_ = q.AddJob(ctx, msg)
	}

	// Wait for all messages to be processed.
	_ = q.Close()

	// Output: hello world
}

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

func timeSince(start time.Time) time.Duration {
	secs := float64(time.Since(start)) / float64(time.Second)
	return time.Duration(math.Floor(secs)) * time.Second
}

func timeSinceCeil(start time.Time) time.Duration {
	secs := float64(time.Since(start)) / float64(time.Second)
	return time.Duration(math.Ceil(secs)) * time.Second
}
