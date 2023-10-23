package internal

import (
	"context"
	"log"
	"math/rand"
	"os"
	"reflect"
	"time"
)

var (
	Logger     = log.New(os.Stderr, "ch: ", log.LstdFlags|log.Lshortfile)
	Warn       = log.New(os.Stderr, "WARN: ch: ", log.LstdFlags|log.Lshortfile)
	Deprecated = log.New(os.Stderr, "DEPRECATED: ch: ", log.LstdFlags|log.Lshortfile)
)

func Sleep(ctx context.Context, dur time.Duration) error {
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
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}

func MakeSliceNextElemFunc(v reflect.Value) func() reflect.Value {
	if v.Kind() == reflect.Array {
		var pos int
		return func() reflect.Value {
			v := v.Index(pos)
			pos++
			return v
		}
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
