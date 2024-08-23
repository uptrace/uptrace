package redisq_test

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/taskq/redisq/v4"
	"github.com/vmihailenco/taskq/taskqtest/v4"
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

func queueName(s string) string {
	version := strings.Split(runtime.Version(), " ")[0]
	version = strings.Replace(version, ".", "", -1)
	return "test-" + s + "-" + version
}

func newFactory() taskq.Factory {
	return redisq.NewFactory()
}

func TestConsumer(t *testing.T) {
	taskqtest.TestConsumer(t, newFactory(), &taskq.QueueConfig{
		Name: queueName("consumer"),
	})
}

func TestUnknownTask(t *testing.T) {
	taskqtest.TestUnknownTask(t, newFactory(), &taskq.QueueConfig{
		Name: queueName("unknown-task"),
	})
}

func TestFallback(t *testing.T) {
	taskqtest.TestFallback(t, newFactory(), &taskq.QueueConfig{
		Name: queueName("fallback"),
	})
}

func TestDelay(t *testing.T) {
	taskqtest.TestDelay(t, newFactory(), &taskq.QueueConfig{
		Name: queueName("delay"),
	})
}

func TestRetry(t *testing.T) {
	taskqtest.TestRetry(t, newFactory(), &taskq.QueueConfig{
		Name: queueName("retry"),
	})
}

func TestNamedJob(t *testing.T) {
	taskqtest.TestNamedJob(t, newFactory(), &taskq.QueueConfig{
		Name: queueName("named-message"),
	})
}

func TestCallOnce(t *testing.T) {
	taskqtest.TestCallOnce(t, newFactory(), &taskq.QueueConfig{
		Name: queueName("call-once"),
	})
}

func TestLen(t *testing.T) {
	taskqtest.TestLen(t, newFactory(), &taskq.QueueConfig{
		Name: queueName("queue-len"),
	})
}

func TestRateLimit(t *testing.T) {
	taskqtest.TestRateLimit(t, newFactory(), &taskq.QueueConfig{
		Name: queueName("rate-limit"),
	})
}

func TestErrorDelay(t *testing.T) {
	taskqtest.TestErrorDelay(t, newFactory(), &taskq.QueueConfig{
		Name: queueName("delayer"),
	})
}

func TestBatchConsumerSmallJob(t *testing.T) {
	taskqtest.TestBatchConsumer(t, newFactory(), &taskq.QueueConfig{
		Name: queueName("batch-consumer-small-message"),
	}, 100)
}

func TestBatchConsumerLarge(t *testing.T) {
	taskqtest.TestBatchConsumer(t, newFactory(), &taskq.QueueConfig{
		Name: queueName("batch-processor-large-message"),
	}, 64000)
}

func TestAckJob(t *testing.T) {
	ctx := context.Background()
	opt := &taskq.QueueConfig{
		Name:               queueName("ack-message"),
		ReservationTimeout: 1 * time.Second,
		WaitTimeout:        waitTimeout,
		Redis:              redisRing(),
	}

	red, ok := opt.Redis.(redisq.RedisStreamClient)
	if !ok {
		log.Fatal(fmt.Errorf("redisq: Redis client must support streams"))
	}

	factory := newFactory()
	q := factory.RegisterQueue(opt)
	defer q.Close()

	err := q.Purge(ctx)
	require.NoError(t, err)

	ch := make(chan time.Time)
	task := taskq.RegisterTask(ulid.Make().String(), &taskq.TaskConfig{
		Handler: func() error {
			ch <- time.Now()
			return nil
		},
	})

	err = q.AddJob(ctx, task.NewJob())
	require.NoError(t, err)

	p := q.Consumer()
	if err := p.Start(ctx); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ch:
	case <-time.After(testTimeout):
		t.Fatalf("message was not processed")
	}

	tm := time.Now().Add(opt.ReservationTimeout)
	end := strconv.FormatInt(tm.UnixNano()/int64(time.Millisecond), 10)
	pending, err := red.XPendingExt(context.Background(), &redis.XPendingExtArgs{
		Stream: "taskq:{" + opt.Name + "}:stream",
		Group:  "taskq",
		Start:  "-",
		End:    end,
		Count:  100,
	}).Result()
	if err != nil {
		t.Fatal(err)
	}

	if len(pending) > 0 {
		t.Fatal("task not acknowledged and still exists in pending list.")
	}

	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
}
