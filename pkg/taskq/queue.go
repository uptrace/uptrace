package taskq

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/go-redis/redis_rate/v10"
)

type QueueConfig struct {
	// Queue name.
	Name string

	NumWorker int
	// Global limit of concurrently running workers across all servers.
	// Overrides NumWorker.
	WorkerLimit int
	// Maximum number of goroutines fetching messages.
	// Default is 8 * number of CPUs.
	NumFetcher int

	// Number of messages reserved by a fetcher in the queue in one request.
	// Default is 10 messages.
	ReservationSize int
	// Time after which the reserved message is returned to the queue.
	// Default is 5 minutes.
	ReservationTimeout time.Duration
	// Time that a long polling receive call waits for a message to become
	// available before returning an empty response.
	// Default is 10 seconds.
	WaitTimeout time.Duration
	// Size of the buffer where reserved messages are stored.
	// Default is the same as ReservationSize.
	BufferSize int

	// Number of consecutive failures after which queue processing is paused.
	// Default is 100 failures.
	PauseErrorsThreshold int

	// Processing rate limit.
	RateLimit redis_rate.Limit

	// Optional rate limiter. The default is to use Redis.
	RateLimiter *redis_rate.Limiter

	// Redis client that is used for storing metadata.
	Redis Redis

	// Optional storage interface. The default is to use Redis.
	Storage Storage

	// Optional message handler. The default is the global Tasks registry.
	Handler Handler

	inited bool

	// ConsumerIdleTimeout Time after which the consumer need to be deleted.
	// Default is 6 hour
	ConsumerIdleTimeout time.Duration

	// SchedulerBackoffTime is the time of backoff for the scheduler(
	// Scheduler was designed to clean zombie Consumer and requeue pending msgs, and so on.
	// Default is randomly between 1~1.5s
	// We can change it to a bigger value so that it won't slowdown the redis when using redis queue.
	// It will be between SchedulerBackoffTime and SchedulerBackoffTime+250ms.
	SchedulerBackoffTime time.Duration
}

func (opt *QueueConfig) Init() {
	if opt.inited {
		return
	}
	opt.inited = true

	if opt.Name == "" {
		panic("QueueConfig.Name is required")
	}

	if opt.WorkerLimit > 0 {
		opt.NumWorker = opt.WorkerLimit
	} else {
		if opt.NumWorker == 0 {
			opt.NumWorker = 2 * runtime.GOMAXPROCS(0)
		}
	}
	if opt.NumFetcher == 0 {
		opt.NumFetcher = 1
	}

	switch opt.PauseErrorsThreshold {
	case -1:
		opt.PauseErrorsThreshold = 0
	case 0:
		opt.PauseErrorsThreshold = 100
	}

	if opt.ReservationSize == 0 {
		opt.ReservationSize = 10
	}
	if opt.ReservationTimeout == 0 {
		opt.ReservationTimeout = 5 * time.Minute
	}
	if opt.BufferSize == 0 {
		opt.BufferSize = opt.ReservationSize
	}
	if opt.WaitTimeout == 0 {
		opt.WaitTimeout = 10 * time.Second
	}

	if opt.ConsumerIdleTimeout == 0 {
		opt.ConsumerIdleTimeout = 6 * time.Hour
	}

	if opt.Storage == nil {
		opt.Storage = newRedisStorage(opt.Redis)
	}

	if !opt.RateLimit.IsZero() && opt.RateLimiter == nil && opt.Redis != nil {
		opt.RateLimiter = redis_rate.NewLimiter(opt.Redis)
	}

	if opt.Handler == nil {
		opt.Handler = &Tasks
	}
}

//------------------------------------------------------------------------------

type Queue interface {
	fmt.Stringer
	Name() string
	Options() *QueueConfig
	Consumer() QueueConsumer

	Len(ctx context.Context) (int, error)
	AddJob(ctx context.Context, msg *Job) error
	ReserveN(ctx context.Context, n int, waitTimeout time.Duration) ([]Job, error)
	Release(ctx context.Context, msg *Job) error
	Delete(ctx context.Context, msg *Job) error
	Purge(ctx context.Context) error
	Close() error
	CloseTimeout(timeout time.Duration) error
}

// QueueConsumer reserves messages from the queue, processes them,
// and then either releases or deletes messages from the queue.
type QueueConsumer interface {
	// AddHook adds a hook into message processing.
	AddHook(hook ConsumerHook)
	Queue() Queue
	Options() *QueueConfig
	Len() int
	// Stats returns processor stats.
	Stats() *ConsumerStats
	AddJob(ctx context.Context, job *Job) error
	// Start starts consuming messages in the queue.
	Start(ctx context.Context) error
	// Stop is StopTimeout with 30 seconds timeout.
	Stop() error
	// StopTimeout waits workers for timeout duration to finish processing current
	// messages and stops workers.
	StopTimeout(timeout time.Duration) error
	// ProcessAll starts workers to process messages in the queue and then stops
	// them when all messages are processed.
	ProcessAll(ctx context.Context) error
	// ProcessOne processes at most one message in the queue.
	ProcessOne(ctx context.Context) error
	// Process is low-level API to process message bypassing the internal queue.
	Process(ctx context.Context, msg *Job) error
	Put(ctx context.Context, msg *Job)
	// Purge discards messages from the internal queue.
	Purge(ctx context.Context) error
	String() string
}
