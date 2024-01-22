package unixtime

import (
	"time"
)

type Seconds uint32

func ToSeconds(t time.Time) Seconds {
	return Seconds(t.Unix())
}

func (s Seconds) String() string {
	return s.Time().String()
}

func (s Seconds) Add(d time.Duration) Seconds {
	if d > 0 {
		return s + Seconds(d.Seconds())
	}
	return s - Seconds(d.Seconds())
}

func (s Seconds) Time() time.Time {
	return time.Unix(int64(s), 0)
}
