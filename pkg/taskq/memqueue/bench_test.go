package memqueue_test

import (
	"context"
	"sync"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/taskq/memqueue/v4"
	"github.com/vmihailenco/taskq/v4"
)

func BenchmarkCallAsync(b *testing.B) {
	taskq.Tasks.Reset()
	ctx := context.Background()

	q := memqueue.NewQueue(&taskq.QueueConfig{
		Name:    "test",
		Storage: taskq.NewLocalStorage(),
	})
	defer q.Close()

	task := taskq.RegisterTask("test", &taskq.TaskConfig{
		Handler: func() {},
	})

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = q.AddJob(ctx, task.NewJob())
		}
	})
}

func BenchmarkNamedJob(b *testing.B) {
	taskq.Tasks.Reset()
	ctx := context.Background()

	q := memqueue.NewQueue(&taskq.QueueConfig{
		Name:    "test",
		Storage: taskq.NewLocalStorage(),
	})
	defer q.Close()

	task := taskq.RegisterTask("test", &taskq.TaskConfig{
		Handler: func() {},
	})

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			msg := task.NewJob()
			msg.Name = "myname"
			q.AddJob(ctx, msg)
		}
	})
}

func BenchmarkConsumerMemq(b *testing.B) {
	benchmarkConsumer(b, memqueue.NewFactory())
}

var (
	once sync.Once
	q    taskq.Queue
	task *taskq.Task
	wg   sync.WaitGroup
)

func benchmarkConsumer(b *testing.B, factory taskq.Factory) {
	ctx := context.Background()

	once.Do(func() {
		q = factory.RegisterQueue(&taskq.QueueConfig{
			Name:  "bench",
			Redis: redisRing(),
		})

		task = taskq.RegisterTask("bench", &taskq.TaskConfig{
			Handler: func() {
				wg.Done()
			},
		})

		_ = q.Consumer().Start(ctx)
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			wg.Add(1)
			_ = q.AddJob(ctx, task.NewJob())
		}
		wg.Wait()
	}
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
