package unixtime

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/segmentio/encoding/json"
	"github.com/uptrace/pkg/msgp"
	"github.com/uptrace/pkg/unsafeconv"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	Nanosecond  Nano = 1
	Microsecond      = 1000 * Nanosecond
	Millisecond      = 1000 * Microsecond
	Second           = 1000 * Millisecond
	Minute           = 60 * Second
	Hour             = 60 * Minute
	Day              = 24 * Hour
)

type Nano int64

var NowStub func() Nano

func Now() Nano {
	if NowStub != nil {
		return NowStub()
	}
	return Nano(now())
}
func Date(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) Nano {
	return ToNano(time.Date(year, month, day, hour, min, sec, nsec, loc))
}
func ToNano(t time.Time) Nano         { return Nano(t.UnixNano()) }
func Unix(sec int64, nsec int64) Nano { return Nano(sec*int64(time.Second) + nsec) }
func Parse(layout, value string) (Nano, error) {
	tm, err := time.Parse(layout, value)
	if err != nil {
		return 0, err
	}
	return ToNano(tm), nil
}
func (t Nano) String() string   { return t.Time().String() }
func (t Nano) Unix() int64      { return int64(t / Second) }
func (t Nano) UnixMicro() int64 { return int64(t / Microsecond) }
func (t Nano) UnixMilli() int64 { return int64(t / Millisecond) }
func (t Nano) UnixNano() int64  { return int64(t) }
func (t Nano) Truncate(d time.Duration) Nano {
	if d == 0 {
		return t
	}
	return t - t%Nano(d)
}
func (t Nano) Ceil(d time.Duration) Nano {
	r := time.Duration(t) % d
	if r == 0 {
		return t
	}
	return t.Add(d - r)
}
func (t Nano) Time() time.Time             { return time.Unix(0, int64(t)).In(time.UTC) }
func (t Nano) Sub(u Nano) time.Duration    { return time.Duration(t - u) }
func (t Nano) Add(d time.Duration) Nano    { return t + Nano(d) }
func (t Nano) Format(layout string) string { return t.Time().Format(layout) }

var _ driver.Valuer = (*Nano)(nil)

func (t Nano) Value() (driver.Value, error) { return t.Time(), nil }

var _ sql.Scanner = (*Nano)(nil)

func (t *Nano) Scan(src any) error {
	const layout = "2006-01-02 15:04:05.999999-07"
	switch src := src.(type) {
	case nil:
		return nil
	case []byte:
		got, err := Parse(layout, unsafeconv.String(src))
		if err != nil {
			return err
		}
		*t = got
		return nil
	case string:
		got, err := Parse(layout, src)
		if err != nil {
			return err
		}
		*t = got
		return nil
	case time.Time:
		*t = ToNano(src)
		return nil
	default:
		return fmt.Errorf("unixtime: can't Scan(%T) into Nano", src)
	}
}

var _ json.Unmarshaler = (*Nano)(nil)

func (t *Nano) UnmarshalJSON(b []byte) error {
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		var tm time.Time
		if err := json.Unmarshal(b, &tm); err != nil {
			return err
		}
		*t = Nano(tm.UnixNano())
		return nil
	}
	ms, err := strconv.ParseFloat(unsafeconv.String(b), 64)
	if err != nil {
		return err
	}
	*t = Nano(ms * 1e6)
	return nil
}

var _ json.Marshaler = (*Nano)(nil)

func (t Nano) MarshalJSON() ([]byte, error) { return json.Marshal(float64(t) / float64(Millisecond)) }

var _ msgp.IsZeroer = (*Nano)(nil)

func (t Nano) IsZero() bool { return t == 0 }

var _ msgp.Sizer = (*Nano)(nil)

func (Nano) MsgpackSize() int { return 8 }

var _ msgp.Appender = (*Nano)(nil)

func (t Nano) AppendMsgpack(b []byte, flags msgp.AppendFlags) (_ []byte, err error) {
	return msgp.AppendVarint(b, int64(t)), nil
}

var _ msgp.Parser = (*Nano)(nil)

func (t *Nano) ParseMsgpack(b []byte, flags msgp.ParseFlags) (_ []byte, err error) {
	n, bb, err := msgp.ParseInt64(b)
	if err != nil {
		if tm, bb, err := msgp.ParseTime(b); err == nil {
			*t = Nano(tm.UnixNano())
			return bb, nil
		}
		return nil, err
	}
	*t = Nano(n)
	return bb, nil
}
func Since(t Nano) time.Duration { return Now().Sub(t) }

var nowTime atomic.Int64

func NowFast() Nano {
	if NowStub != nil {
		return NowStub()
	}
	return Nano(nowTime.Load())
}
func init() {
	nowTime.Store(int64(Now()))
	go func() {
		for {
			time.Sleep(time.Millisecond)
			nowTime.Store(int64(Now()))
		}
	}()
}
