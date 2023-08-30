package bunutil

import (
	"sync"
	"time"
)

type Debouncer struct {
	timer *time.Timer
}

func NewDebouncer() *Debouncer {
	return new(Debouncer)
}

func (d *Debouncer) Run(after time.Duration, f func()) {
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(after, f)
}

type OnceMap struct {
	mu    sync.Mutex
	table map[string]*sync.Once
}

func (m *OnceMap) Do(key string, fn func()) {
	m.mu.Lock()
	once, ok := m.table[key]
	if !ok {
		once = new(sync.Once)
		if m.table == nil {
			m.table = make(map[string]*sync.Once)
		}
		m.table[key] = once
	}
	m.mu.Unlock()
	once.Do(fn)
}
