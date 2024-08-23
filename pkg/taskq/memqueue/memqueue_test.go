package memqueue_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"

	"github.com/vmihailenco/taskq/memqueue/v4"
	"github.com/vmihailenco/taskq/v4"
)

func TestMemqueue(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "memqueue")
}

var _ = BeforeSuite(func() {
	taskq.SetLogger(log.New(ioutil.Discard, "", 0))
})

var _ = BeforeEach(func() {
	taskq.Tasks.Reset()
})

var _ = Describe("message with args", func() {
	ctx := context.Background()
	ch := make(chan bool, 10)

	BeforeEach(func() {
		q := memqueue.NewQueue(&taskq.QueueConfig{
			Name:    "test",
			Storage: taskq.NewLocalStorage(),
		})
		task := taskq.RegisterTask("test", &taskq.TaskConfig{
			Handler: func(s string, i int) {
				Expect(s).To(Equal("string"))
				Expect(i).To(Equal(42))
				ch <- true
			},
		})
		err := q.AddJob(ctx, task.NewJob("string", 42))
		Expect(err).NotTo(HaveOccurred())

		err = q.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("handler is called with args", func() {
		Expect(ch).To(Receive())
		Expect(ch).NotTo(Receive())
	})
})

var _ = Describe("context.Context", func() {
	ctx := context.Background()
	ch := make(chan bool, 10)

	BeforeEach(func() {
		q := memqueue.NewQueue(&taskq.QueueConfig{
			Name:    "test",
			Storage: taskq.NewLocalStorage(),
		})
		task := taskq.RegisterTask("test", &taskq.TaskConfig{
			Handler: func(c context.Context, s string, i int) {
				Expect(s).To(Equal("string"))
				Expect(i).To(Equal(42))
				ch <- true
			},
		})
		err := q.AddJob(ctx, task.NewJob("string", 42))
		Expect(err).NotTo(HaveOccurred())

		err = q.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("handler is called with args", func() {
		Expect(ch).To(Receive())
		Expect(ch).NotTo(Receive())
	})
})

var _ = Describe("message with invalid number of args", func() {
	ctx := context.Background()
	ch := make(chan bool, 10)

	BeforeEach(func() {
		q := memqueue.NewQueue(&taskq.QueueConfig{
			Name:    "test",
			Storage: taskq.NewLocalStorage(),
		})
		task := taskq.RegisterTask("test", &taskq.TaskConfig{
			Handler: func(s string) {
				ch <- true
			},
			RetryLimit: 1,
		})
		q.Consumer().Stop()

		err := q.AddJob(ctx, task.NewJob())
		Expect(err).NotTo(HaveOccurred())

		err = q.Consumer().ProcessOne(ctx)
		Expect(err).To(MatchError("taskq: got 0 args, wanted 1"))

		err = q.Consumer().ProcessAll(ctx)
		Expect(err).NotTo(HaveOccurred())

		err = q.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("handler is not called", func() {
		Expect(ch).NotTo(Receive())
	})
})

var _ = Describe("HandlerFunc", func() {
	ctx := context.Background()
	ch := make(chan bool, 10)

	BeforeEach(func() {
		q := memqueue.NewQueue(&taskq.QueueConfig{
			Name:    "test",
			Storage: taskq.NewLocalStorage(),
		})
		task := taskq.RegisterTask("test", &taskq.TaskConfig{
			Handler: func(msg *taskq.Job) error {
				Expect(msg.Args).To(Equal([]interface{}{"string", 42}))
				ch <- true
				return nil
			},
		})

		err := q.AddJob(ctx, task.NewJob("string", 42))
		Expect(err).NotTo(HaveOccurred())

		err = q.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("is called with Job", func() {
		Expect(ch).To(Receive())
		Expect(ch).NotTo(Receive())
	})
})

var _ = Describe("message retry timing", func() {
	ctx := context.Background()
	var q *memqueue.Queue
	var task *taskq.Task
	backoff := 100 * time.Millisecond
	var count int
	var ch chan time.Time

	BeforeEach(func() {
		count = 0
		ch = make(chan time.Time, 10)
		q = memqueue.NewQueue(&taskq.QueueConfig{
			Name:    "test",
			Storage: taskq.NewLocalStorage(),
		})
		task = taskq.RegisterTask("test", &taskq.TaskConfig{
			Handler: func() error {
				ch <- time.Now()
				count++
				return fmt.Errorf("fake error #%d", count)
			},
			RetryLimit: 3,
			MinBackoff: backoff,
		})
	})

	Context("without delay", func() {
		var now time.Time

		BeforeEach(func() {
			now = time.Now()
			_ = q.AddJob(ctx, task.NewJob())

			err := q.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("is retried in time", func() {
			Expect(ch).To(Receive(BeTemporally("~", now, backoff/10)))
			Expect(ch).To(Receive(BeTemporally("~", now.Add(backoff), backoff/10)))
			Expect(ch).To(Receive(BeTemporally("~", now.Add(3*backoff), backoff/10)))
			Expect(ch).NotTo(Receive())
		})
	})

	Context("message with delay", func() {
		var now time.Time

		BeforeEach(func() {
			msg := task.NewJob()
			msg.Delay = 5 * backoff
			now = time.Now().Add(msg.Delay)

			q.AddJob(ctx, msg)

			err := q.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("is retried in time", func() {
			Expect(ch).To(Receive(BeTemporally("~", now, backoff/10)))
			Expect(ch).To(Receive(BeTemporally("~", now.Add(backoff), backoff/10)))
			Expect(ch).To(Receive(BeTemporally("~", now.Add(3*backoff), backoff/10)))
			Expect(ch).NotTo(Receive())
		})
	})
})

var _ = Describe("failing queue with error handler", func() {
	ctx := context.Background()
	var q *memqueue.Queue
	ch := make(chan bool, 10)

	BeforeEach(func() {
		q = memqueue.NewQueue(&taskq.QueueConfig{
			Name:    "test",
			Storage: taskq.NewLocalStorage(),
		})
		task := taskq.RegisterTask("test", &taskq.TaskConfig{
			Handler: func() error {
				return errors.New("fake error")
			},
			FallbackHandler: func() {
				ch <- true
			},
			RetryLimit: 1,
		})
		q.AddJob(ctx, task.NewJob())

		err := q.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("error handler is called when handler fails", func() {
		Expect(ch).To(Receive())
		Expect(ch).NotTo(Receive())
	})
})

var _ = Describe("named message", func() {
	ctx := context.Background()
	var count int64

	BeforeEach(func() {
		q := memqueue.NewQueue(&taskq.QueueConfig{
			Name:    "test",
			Storage: taskq.NewLocalStorage(),
		})
		task := taskq.RegisterTask("test", &taskq.TaskConfig{
			Handler: func() {
				atomic.AddInt64(&count, 1)
			},
		})

		name := uuid.NewV4().String()

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				msg := task.NewJob()
				msg.Name = name
				q.AddJob(ctx, msg)
			}()
		}
		wg.Wait()

		err := q.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("is processed once", func() {
		n := atomic.LoadInt64(&count)
		Expect(n).To(Equal(int64(1)))
	})
})

var _ = Describe("CallOnce", func() {
	ctx := context.Background()
	var now time.Time
	delay := 3 * time.Second
	ch := make(chan time.Time, 10)

	BeforeEach(func() {
		now = time.Now()

		q := memqueue.NewQueue(&taskq.QueueConfig{
			Name:    "test",
			Storage: taskq.NewLocalStorage(),
		})
		task := taskq.RegisterTask("test", &taskq.TaskConfig{
			Handler: func() error {
				ch <- time.Now()
				return nil
			},
		})

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()

				msg := task.NewJob()
				msg.OnceInPeriod(delay)

				q.AddJob(ctx, msg)
			}()
		}
		wg.Wait()

		err := q.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("processes message once with delay", func() {
		Eventually(ch).Should(Receive(BeTemporally(">", now.Add(delay), time.Second)))
		Consistently(ch).ShouldNot(Receive())
	})
})

var _ = Describe("stress testing", func() {
	const n = 10000
	ctx := context.Background()
	var count int64

	BeforeEach(func() {
		q := memqueue.NewQueue(&taskq.QueueConfig{
			Name:    "test",
			Storage: taskq.NewLocalStorage(),
		})
		task := taskq.RegisterTask("test", &taskq.TaskConfig{
			Handler: func() {
				atomic.AddInt64(&count, 1)
			},
		})

		for i := 0; i < n; i++ {
			q.AddJob(ctx, task.NewJob())
		}

		err := q.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("handler is called for all messages", func() {
		nn := atomic.LoadInt64(&count)
		Expect(nn).To(Equal(int64(n)))
	})
})

var _ = Describe("stress testing failing queue", func() {
	const n = 100000
	ctx := context.Background()
	var errorCount int64

	BeforeEach(func() {
		q := memqueue.NewQueue(&taskq.QueueConfig{
			Name:                 "test",
			PauseErrorsThreshold: -1,
			Storage:              taskq.NewLocalStorage(),
		})
		task := taskq.RegisterTask("test", &taskq.TaskConfig{
			Handler: func() error {
				return errors.New("fake error")
			},
			FallbackHandler: func() {
				atomic.AddInt64(&errorCount, 1)
			},
			RetryLimit: 1,
		})

		for i := 0; i < n; i++ {
			q.AddJob(ctx, task.NewJob())
		}

		err := q.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("error handler is called for all messages", func() {
		nn := atomic.LoadInt64(&errorCount)
		Expect(nn).To(Equal(int64(n)))
	})
})

var _ = Describe("empty queue", func() {
	ctx := context.Background()
	var q *memqueue.Queue
	var task *taskq.Task
	var processed uint32

	BeforeEach(func() {
		processed = 0
		q = memqueue.NewQueue(&taskq.QueueConfig{
			Name:    "test",
			Storage: taskq.NewLocalStorage(),
		})
		task = taskq.RegisterTask("test", &taskq.TaskConfig{
			Handler: func() {
				atomic.AddUint32(&processed, 1)
			},
		})
	})

	AfterEach(func() {
		_ = q.Close()
	})

	It("can be closed", func() {
		err := q.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("stops processor", func() {
		err := q.Consumer().Stop()
		Expect(err).NotTo(HaveOccurred())
	})

	testEmptyQueue := func() {
		It("processes all messages", func() {
			err := q.Consumer().ProcessAll(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		It("processes one message", func() {
			err := q.Consumer().ProcessOne(ctx)
			Expect(err).To(MatchError("taskq: queue is empty"))

			err = q.Consumer().ProcessAll(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
	}

	Context("when processor is stopped", func() {
		BeforeEach(func() {
			err := q.Consumer().Stop()
			Expect(err).NotTo(HaveOccurred())
		})

		testEmptyQueue()

		Context("when there are messages in the queue", func() {
			BeforeEach(func() {
				for i := 0; i < 3; i++ {
					err := q.AddJob(ctx, task.NewJob())
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("processes all messages", func() {
				p := q.Consumer()

				err := p.ProcessAll(ctx)
				Expect(err).NotTo(HaveOccurred())

				n := atomic.LoadUint32(&processed)
				Expect(n).To(Equal(uint32(3)))
			})

			It("processes one message", func() {
				p := q.Consumer()

				err := p.ProcessOne(ctx)
				Expect(err).NotTo(HaveOccurred())

				n := atomic.LoadUint32(&processed)
				Expect(n).To(Equal(uint32(1)))

				err = p.ProcessAll(ctx)
				Expect(err).NotTo(HaveOccurred())

				n = atomic.LoadUint32(&processed)
				Expect(n).To(Equal(uint32(3)))
			})
		})
	})
})

// slot splits time into equal periods (called slots) and returns
// slot number for provided time.
func slot(period time.Duration) int64 {
	tm := time.Now()
	periodSec := int64(period / time.Second)
	if periodSec == 0 {
		return tm.Unix()
	}
	return tm.Unix() / periodSec
}
