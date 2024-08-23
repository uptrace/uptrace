package redisexample

import (
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/vmihailenco/taskq/redisq/v4"
	"github.com/vmihailenco/taskq/v4"
)

var Redis = redis.NewClient(&redis.Options{
	Addr: ":6379",
})

var (
	QueueFactory = redisq.NewFactory()
	MainQueue    = QueueFactory.RegisterQueue(&taskq.QueueConfig{
		Name:  "api-worker",
		Redis: Redis,
	})
	CountTask = taskq.RegisterTask("counter", &taskq.TaskConfig{
		Handler: func() error {
			IncrLocalCounter()
			time.Sleep(time.Millisecond)
			return nil
		},
	})
)

var counter int32

func GetLocalCounter() int32 {
	return atomic.LoadInt32(&counter)
}

func IncrLocalCounter() {
	atomic.AddInt32(&counter, 1)
}

func LogStats() {
	var prev int32
	for range time.Tick(3 * time.Second) {
		n := GetLocalCounter()
		log.Printf("processed %d tasks (%d/s)", n, (n-prev)/3)
		prev = n
	}
}

func WaitSignal() os.Signal {
	ch := make(chan os.Signal, 2)
	signal.Notify(
		ch,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
			return sig
		}
	}
}
