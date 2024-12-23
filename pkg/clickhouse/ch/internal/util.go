package internal

import (
	"context"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/rand"
	"log"
	"os"
	"reflect"
	"time"
)

var Tracer = otel.Tracer("github.com/uptrace/pkg/clickhouse")
var (
	Logger     = log.New(os.Stderr, "ch: ", log.LstdFlags|log.Lshortfile)
	Warn       = log.New(os.Stderr, "WARN: ch: ", log.LstdFlags|log.Lshortfile)
	Deprecated = log.New(os.Stderr, "DEPRECATED: ch: ", log.LstdFlags|log.Lshortfile)
)

func Sleep(ctx context.Context, dur time.Duration) error {
	if dur <= 0 {
		return nil
	}
	ctx, span := Tracer.Start(ctx, "Sleep")
	defer span.End()
	t := time.NewTimer(dur)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func Unwrap(err error) error {
	u, ok := err.(interface{ Unwrap() error })
	if !ok {
		return nil
	}
	return u.Unwrap()
}
func MakeSliceNextElemFunc(v reflect.Value) func() reflect.Value {
	if v.Kind() == reflect.Array {
		var pos int
		return func() reflect.Value { v := v.Index(pos); pos++; return v }
	}
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
		return func() reflect.Value {
			if v.Len() < v.Cap() {
				v.Set(v.Slice(0, v.Len()+1))
				elem := v.Index(v.Len() - 1)
				if elem.IsNil() {
					elem.Set(reflect.New(elemType))
				}
				return elem.Elem()
			}
			elem := reflect.New(elemType)
			v.Set(reflect.Append(v, elem))
			return elem.Elem()
		}
	}
	zero := reflect.Zero(elemType)
	return func() reflect.Value {
		if v.Len() < v.Cap() {
			v.Set(v.Slice(0, v.Len()+1))
			return v.Index(v.Len() - 1)
		}
		v.Set(reflect.Append(v, zero))
		return v.Index(v.Len() - 1)
	}
}
func RetryBackoff(minBackoff, maxBackoff time.Duration) time.Duration {
	backoff := minBackoff + time.Duration(rand.Int63n(int64(maxBackoff)))
	if backoff > maxBackoff {
		return maxBackoff
	}
	return backoff
}

type ErrorCounter struct {
	errs      []error
	i         int
	lastErr   error
	count     int
	threshold int
}

func NewErrorCounter(window int, threshold int) *ErrorCounter {
	return &ErrorCounter{errs: make([]error, window), threshold: threshold}
}
func (c *ErrorCounter) Add(err error) error {
	c.i = (c.i + 1) % len(c.errs)
	old := c.errs[c.i]
	c.errs[c.i] = err
	if err != nil {
		c.count++
		c.lastErr = err
	}
	if old != nil {
		c.count--
	}
	if c.count >= c.threshold {
		return c.lastErr
	}
	return nil
}
