package taskq_test

import (
	"context"
	"fmt"
	"time"

	"github.com/vmihailenco/taskq/memqueue/v4"
	"github.com/vmihailenco/taskq/v4"
)

type RateLimitError string

func (e RateLimitError) Error() string {
	return string(e)
}

func (RateLimitError) Delay() time.Duration {
	return 3 * time.Second
}

func Example_customRateLimit() {
	start := time.Now()
	q := memqueue.NewQueue(&taskq.QueueConfig{
		Name: "test",
	})
	task := taskq.RegisterTask("Example_customRateLimit", &taskq.TaskConfig{
		Handler: func() error {
			fmt.Println("retried in", timeSince(start))
			return RateLimitError("calm down")
		},
		RetryLimit: 2,
		MinBackoff: time.Millisecond,
	})

	ctx := context.Background()
	q.AddJob(ctx, task.NewJob())

	// Wait for all messages to be processed.
	_ = q.Close()

	// Output: retried in 0s
	// retried in 3s
}
