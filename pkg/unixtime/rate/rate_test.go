package rate

import (
	"github.com/stretchr/testify/require"
	"github.com/uptrace/pkg/unixtime"
	"go4.org/syncutil"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	{
		lim := NewLimiter(Every(time.Second), 1)
		require.Equal(t, int64(0), lim.AllowN(unixtime.Now(), 0))
		require.Equal(t, int64(1), lim.AllowN(unixtime.Now(), 10))
		require.Equal(t, int64(0), lim.AllowN(unixtime.Now(), 10))
		require.Equal(t, int64(0), lim.AllowN(unixtime.Now(), 0))
	}
	{
		lim := NewLimiter(Every(time.Second), 3)
		require.Equal(t, int64(3), lim.AllowN(unixtime.Now(), 10))
		require.Equal(t, int64(0), lim.AllowN(unixtime.Now(), 10))
		require.Equal(t, int64(0), lim.AllowN(unixtime.Now(), 0))
	}
	{
		lim := NewLimiter(Every(time.Millisecond), 1)
		for i := 0; i < 10; i++ {
			require.Equal(t, int64(1), lim.AllowN(unixtime.Now(), 10))
			require.Equal(t, int64(0), lim.AllowN(unixtime.Now(), 10))
			time.Sleep(time.Millisecond)
		}
	}
}
func TestLimiterParallel(t *testing.T) {
	lim := NewLimiter(Every(time.Millisecond), 1)
	var allowed atomic.Int64
	var group syncutil.Group
	for i := 0; i < runtime.NumCPU(); i++ {
		group.Go(func() error {
			for i := 0; i < 1000; i++ {
				n := lim.AllowN(unixtime.Now(), 10)
				allowed.Add(int64(n))
				time.Sleep(time.Millisecond)
			}
			return nil
		})
	}
	group.Wait()
	require.InDelta(t, 1000, allowed.Load(), 25)
}
func TestLimiterBurst(t *testing.T) {
	const burst = 1_000_000
	lim := NewLimiter(Every(time.Millisecond), burst)
	for i := 0; i < burst; i++ {
		if got := lim.AllowN(unixtime.Now(), 1); got != 1 {
			require.Equal(t, 1, got, i)
		}
	}
	var allowed int64
	for i := 0; i < 1000; i++ {
		allowed += lim.AllowN(unixtime.Now(), 1)
	}
	require.InDelta(t, 50, allowed, 50)
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond)
		require.Equal(t, int64(1), lim.AllowN(unixtime.Now(), 1))
	}
}
func BenchmarkAllowNParallel(b *testing.B) {
	lim := NewLimiter(Every(time.Second), 1)
	now := unixtime.Now()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lim.AllowN(now, 1)
		}
	})
}
func BenchmarkAllowNSeq(b *testing.B) {
	lim := NewLimiter(Limit(b.N), int64(b.N))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lim.AllowN(unixtime.NowFast(), 1)
	}
}
