package base

import (
	"context"
	"fmt"
	"sync"

	"github.com/vmihailenco/taskq/v4"
)

type Factory struct {
	m sync.Map
}

func (f *Factory) Register(queue taskq.Queue) error {
	name := queue.Name()
	_, loaded := f.m.LoadOrStore(name, queue)
	if loaded {
		return fmt.Errorf("queue=%q already exists", name)
	}
	return nil
}

func (f *Factory) Unregister(name string) {
	f.m.Delete(name)
}

func (f *Factory) Reset() {
	f.m = sync.Map{}
}

func (f *Factory) Range(fn func(queue taskq.Queue) bool) {
	f.m.Range(func(_, value interface{}) bool {
		return fn(value.(taskq.Queue))
	})
}

func (f *Factory) StartConsumers(ctx context.Context) error {
	return f.forEachQueue(func(q taskq.Queue) error {
		return q.Consumer().Start(ctx)
	})
}

func (f *Factory) StopConsumers() error {
	return f.forEachQueue(func(q taskq.Queue) error {
		return q.Consumer().Stop()
	})
}

func (f *Factory) Close() error {
	return f.forEachQueue(func(q taskq.Queue) error {
		return q.Close()
	})
}

func (f *Factory) forEachQueue(fn func(taskq.Queue) error) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	f.Range(func(q taskq.Queue) bool {
		wg.Add(1)
		go func(q taskq.Queue) {
			defer wg.Done()
			err := fn(q)
			select {
			case errCh <- err:
			default:
			}
		}(q)
		return true
	})
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
