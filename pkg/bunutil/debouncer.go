package bunutil

import (
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
