package redisq

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/bsm/redislock"

	"github.com/vmihailenco/taskq/v4"
	"github.com/vmihailenco/taskq/v4/backend"
	"github.com/vmihailenco/taskq/v4/backend/jobutil"
	"github.com/vmihailenco/taskq/v4/backend/unsafeconv"
)

const batchSize = 100

type RedisStreamClient interface {
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	TxPipeline() redis.Pipeliner

	XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd
	XDel(ctx context.Context, stream string, ids ...string) *redis.IntCmd
	XLen(ctx context.Context, stream string) *redis.IntCmd
	XRangeN(ctx context.Context, stream, start, stop string, count int64) *redis.XMessageSliceCmd
	XGroupCreateMkStream(ctx context.Context, stream, group, start string) *redis.StatusCmd
	XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd
	XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd
	XPendingExt(ctx context.Context, a *redis.XPendingExtArgs) *redis.XPendingExtCmd
	XTrimMaxLen(ctx context.Context, key string, maxLen int64) *redis.IntCmd
	XGroupDelConsumer(ctx context.Context, stream, group, consumer string) *redis.IntCmd

	ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd
	ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd
	ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	XInfoConsumers(ctx context.Context, key string, group string) *redis.XInfoConsumersCmd
}

type Queue struct {
	opt *taskq.QueueConfig

	consumer *taskq.Consumer

	redis RedisStreamClient
	wg    sync.WaitGroup

	zset                string
	stream              string
	streamGroup         string
	streamConsumer      string
	schedulerLockPrefix string

	_closed uint32
}

var _ taskq.Queue = (*Queue)(nil)

func NewQueue(opt *taskq.QueueConfig) *Queue {
	const redisPrefix = "taskq:"

	if opt.WaitTimeout == 0 {
		opt.WaitTimeout = time.Second
	}
	opt.Init()
	if opt.Redis == nil {
		panic(fmt.Errorf("redisq: Redis client is required"))
	}
	red, ok := opt.Redis.(RedisStreamClient)
	if !ok {
		panic(fmt.Errorf("redisq: Redis client must support streams"))
	}

	q := &Queue{
		opt: opt,

		redis: red,

		zset:                redisPrefix + "{" + opt.Name + "}:zset",
		stream:              redisPrefix + "{" + opt.Name + "}:stream",
		streamGroup:         "taskq",
		streamConsumer:      consumer(),
		schedulerLockPrefix: redisPrefix + opt.Name + ":scheduler-lock:",
	}

	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		q.scheduler("delayed", q.scheduleDelayed)
	}()

	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		q.scheduler("pending", q.schedulePending)
	}()

	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		q.scheduler("clean_zombie_consumers", q.cleanZombieConsumers)
	}()

	return q
}

func consumer() string {
	s, _ := os.Hostname()
	s += ":pid:" + strconv.Itoa(os.Getpid())
	s += ":" + strconv.Itoa(rand.Int())
	return s
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
	if q.consumer == nil {
		q.consumer = taskq.NewConsumer(q)
	}
	return q.consumer
}

func (q *Queue) Len(ctx context.Context) (int, error) {
	n, err := q.redis.XLen(ctx, q.stream).Result()
	return int(n), err
}

// Add adds message to the queue.
func (q *Queue) AddJob(ctx context.Context, msg *taskq.Job) error {
	return q.add(ctx, q.redis, msg)
}

func (q *Queue) add(ctx context.Context, pipe RedisStreamClient, msg *taskq.Job) error {
	if msg.TaskName == "" {
		return backend.ErrTaskNameRequired
	}
	if msg.Name != "" && q.isDuplicate(ctx, msg) {
		msg.Err = taskq.ErrDuplicate
		return nil
	}

	if msg.ID == "" {
		u := uuid.New()
		msg.ID = unsafeconv.String(u[:])
	}

	body, err := msg.MarshalBinary()
	if err != nil {
		return err
	}

	if msg.Delay > 0 {
		tm := time.Now().Add(msg.Delay)
		return pipe.ZAdd(ctx, q.zset, redis.Z{
			Score:  float64(unixMs(tm)),
			Member: body,
		}).Err()
	}

	return pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: q.stream,
		Values: map[string]interface{}{
			"body": body,
		},
	}).Err()
}

func (q *Queue) ReserveN(
	ctx context.Context, n int, waitTimeout time.Duration,
) ([]taskq.Job, error) {
	streams, err := q.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
		Streams:  []string{q.stream, ">"},
		Group:    q.streamGroup,
		Consumer: q.streamConsumer,
		Count:    int64(n),
		Block:    waitTimeout,
	}).Result()
	if err != nil {
		if err == redis.Nil { // timeout
			return nil, nil
		}
		if strings.HasPrefix(err.Error(), "NOGROUP") {
			q.createStreamGroup(ctx)
			return q.ReserveN(ctx, n, waitTimeout)
		}
		return nil, err
	}

	stream := &streams[0]
	msgs := make([]taskq.Job, len(stream.Messages))
	for i := range stream.Messages {
		xmsg := &stream.Messages[i]
		msg := &msgs[i]

		if err := unmarshalJob(msg, xmsg); err != nil {
			msg.Err = err
		}
	}

	return msgs, nil
}

func (q *Queue) createStreamGroup(ctx context.Context) {
	_ = q.redis.XGroupCreateMkStream(ctx, q.stream, q.streamGroup, "0").Err()
}

func (q *Queue) Release(ctx context.Context, msg *taskq.Job) error {
	// Make the delete and re-queue operation atomic in case we crash midway
	// and lose a message.
	pipe := q.redis.TxPipeline()
	// When Release a msg, ack it before we delete msg.
	if err := pipe.XAck(ctx, q.stream, q.streamGroup, msg.ID).Err(); err != nil {
		return err
	}

	err := pipe.XDel(ctx, q.stream, msg.ID).Err()
	if err != nil {
		return err
	}

	msg.ReservedCount++
	if err := q.add(ctx, pipe, msg); err != nil {
		return err
	}

	_, err = pipe.Exec(ctx)
	return err
}

// Delete deletes the message from the queue.
func (q *Queue) Delete(ctx context.Context, msg *taskq.Job) error {
	if err := q.redis.XAck(ctx, q.stream, q.streamGroup, msg.ID).Err(); err != nil {
		return err
	}
	return q.redis.XDel(ctx, q.stream, msg.ID).Err()
}

// Purge deletes all messages from the queue.
func (q *Queue) Purge(ctx context.Context) error {
	_ = q.redis.Del(ctx, q.zset).Err()
	_ = q.redis.XTrimMaxLen(ctx, q.stream, 0).Err()
	return nil
}

// Close is like CloseTimeout with 30 seconds timeout.
func (q *Queue) Close() error {
	return q.CloseTimeout(30 * time.Second)
}

// CloseTimeout closes the queue waiting for pending messages to be processed.
func (q *Queue) CloseTimeout(timeout time.Duration) error {
	if !atomic.CompareAndSwapUint32(&q._closed, 0, 1) {
		return nil
	}

	if q.consumer != nil {
		_ = q.consumer.StopTimeout(timeout)
	}

	_ = q.redis.XGroupDelConsumer(
		context.TODO(), q.stream, q.streamGroup, q.streamConsumer).Err()

	return nil
}

func (q *Queue) closed() bool {
	return atomic.LoadUint32(&q._closed) == 1
}

func (q *Queue) scheduler(name string, fn func(ctx context.Context) (int, error)) {
	for {
		if q.closed() {
			break
		}

		ctx := context.TODO()

		var n int
		err := q.withRedisLock(ctx, q.schedulerLockPrefix+name, func(ctx context.Context) error {
			var err error
			n, err = fn(ctx)
			return err
		})
		if err != nil && err != redislock.ErrNotObtained {
			backend.Error(err, "redisq: scheduler failed",
				"name", name)
		}
		if err != nil || n == 0 {
			time.Sleep(q.schedulerBackoff())
		}
	}
}

func (q *Queue) schedulerBackoff() time.Duration {
	n := rand.Intn(500)
	if q.opt.SchedulerBackoffTime > 0 {
		return q.opt.SchedulerBackoffTime + time.Duration(n)*time.Millisecond
	}
	return time.Duration(n+1000) * time.Millisecond
}

func (q *Queue) scheduleDelayed(ctx context.Context) (int, error) {
	tm := time.Now()
	max := strconv.FormatInt(unixMs(tm), 10)
	bodies, err := q.redis.ZRangeByScore(ctx, q.zset, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   max,
		Count: batchSize,
	}).Result()
	if err != nil {
		return 0, err
	}

	pipe := q.redis.TxPipeline()
	for _, body := range bodies {
		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: q.stream,
			Values: map[string]interface{}{
				"body": body,
			},
		})
		pipe.ZRem(ctx, q.zset, body)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return len(bodies), nil
}

func (q *Queue) cleanZombieConsumers(ctx context.Context) (int, error) {
	consumers, err := q.redis.XInfoConsumers(ctx, q.stream, q.streamGroup).Result()
	if err != nil {
		return 0, err
	}

	for _, consumer := range consumers {
		// skip the current stream consumer
		if consumer.Name == q.streamConsumer {
			continue
		}

		if time.Duration(consumer.Idle)*time.Millisecond > q.opt.ConsumerIdleTimeout {
			_ = q.redis.XGroupDelConsumer(ctx, q.stream, q.streamGroup, consumer.Name).Err()
		}
	}
	return 0, nil
}

// schedulePending schedules pending messages that are older than the `ReservationTimeout`.
// ReservationTimeout is the time after which a message is considered to be not processed and need to be re-enqueue.
func (q *Queue) schedulePending(ctx context.Context) (int, error) {
	tm := time.Now().Add(-q.opt.ReservationTimeout)
	end := strconv.FormatInt(unixMs(tm), 10)

	pending, err := q.redis.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: q.stream,
		Group:  q.streamGroup,
		Start:  "-",
		End:    end,
		Count:  batchSize,
	}).Result()
	if err != nil {
		if strings.HasPrefix(err.Error(), "NOGROUP") {
			q.createStreamGroup(ctx)

			return 0, nil
		}
		return 0, err
	}

	for i := range pending {
		xmsgInfo := &pending[i]
		id := xmsgInfo.ID

		xmsgs, err := q.redis.XRangeN(ctx, q.stream, id, id, 1).Result()
		if err != nil {
			return 0, err
		}

		if len(xmsgs) != 1 {
			err := fmt.Errorf("redisq: can't find pending message id=%q in stream=%q",
				id, q.stream)
			return 0, err
		}

		xmsg := &xmsgs[0]
		msg := new(taskq.Job)
		if err := unmarshalJob(msg, xmsg); err != nil {
			return 0, err
		}

		if err := q.Release(ctx, msg); err != nil {
			return 0, err
		}
	}

	return len(pending), nil
}

func (q *Queue) isDuplicate(ctx context.Context, msg *taskq.Job) bool {
	exists := q.opt.Storage.Exists(ctx, jobutil.FullJobName(q, msg))
	return exists
}

func (q *Queue) withRedisLock(
	ctx context.Context, name string, fn func(ctx context.Context) error,
) error {
	lock, err := redislock.Obtain(ctx, q.opt.Redis, name, time.Minute, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err := lock.Release(ctx); err != nil {
			backend.Error(err, "lock.Release failed")
		}
	}()

	return fn(ctx)
}

func unixMs(tm time.Time) int64 {
	return tm.UnixNano() / int64(time.Millisecond)
}

func unmarshalJob(msg *taskq.Job, xmsg *redis.XMessage) error {
	body := xmsg.Values["body"].(string)
	if err := msg.UnmarshalBinary(unsafeconv.Bytes(body)); err != nil {
		return err
	}

	msg.ID = xmsg.ID
	if msg.ReservedCount == 0 {
		msg.ReservedCount = 1
	}

	return nil
}
