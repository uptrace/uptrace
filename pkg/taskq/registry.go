package taskq

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var Tasks TaskMap

type TaskMap struct {
	m sync.Map
}

func (r *TaskMap) Get(name string) *Task {
	if v, ok := r.m.Load(name); ok {
		return v.(*Task)
	}
	if v, ok := r.m.Load("*"); ok {
		return v.(*Task)
	}
	return nil
}

func (r *TaskMap) Register(name string, opt *TaskConfig) (*Task, error) {
	opt.init()

	task := &Task{
		name:    name,
		opt:     opt,
		handler: NewHandler(opt.Handler),
	}

	if opt.FallbackHandler != nil {
		task.fallbackHandler = NewHandler(opt.FallbackHandler)
	}

	_, loaded := r.m.LoadOrStore(name, task)
	if loaded {
		return nil, fmt.Errorf("task=%q already exists", name)
	}
	return task, nil
}

func (r *TaskMap) Unregister(task *Task) {
	r.m.Delete(task.Name())
}

func (r *TaskMap) Reset() {
	r.m = sync.Map{}
}

func (r *TaskMap) Range(fn func(name string, task *Task) bool) {
	r.m.Range(func(key, value interface{}) bool {
		return fn(key.(string), value.(*Task))
	})
}

func (r *TaskMap) HandleJob(ctx context.Context, msg *Job) error {
	task := r.Get(msg.TaskName)
	if task == nil {
		msg.Delay = r.delay(msg, nil, unknownTaskOpt)
		return fmt.Errorf("taskq: unknown task=%q", msg.TaskName)
	}

	opt := task.Options()
	if opt.DeferFunc != nil {
		defer opt.DeferFunc()
	}

	msgErr := task.HandleJob(ctx, msg)
	if msgErr == nil {
		return nil
	}

	msg.Delay = r.delay(msg, msgErr, opt)
	return msgErr
}

func (r *TaskMap) delay(msg *Job, msgErr error, opt *TaskConfig) time.Duration {
	if msg.ReservedCount >= opt.RetryLimit {
		return 0
	}
	if delayer, ok := msgErr.(Delayer); ok {
		return delayer.Delay()
	}
	return exponentialBackoff(opt.MinBackoff, opt.MaxBackoff, msg.ReservedCount)
}
