package rate

import (
	"github.com/puzpuzpuz/xsync/v3"
	"github.com/uptrace/pkg/unixtime"
	"math"
	"sync/atomic"
	"time"
)

const Inf = Limit(math.MaxFloat64)
const InfDuration = time.Duration(math.MaxInt64)

func Every(interval time.Duration) Limit {
	if interval <= 0 {
		return Inf
	}
	return 1 / Limit(interval.Seconds())
}

type Limit float64

func NewLimit(n int64, interval time.Duration) Limit { return Every(interval / time.Duration(n)) }
func (limit Limit) durationFromTokens(tokens int64) time.Duration {
	if limit <= 0 {
		return InfDuration
	}
	seconds := float64(tokens) / float64(limit)
	return time.Duration(float64(time.Second) * seconds)
}
func (limit Limit) tokensFromDuration(d time.Duration) int64 {
	if limit <= 0 {
		return 0
	}
	return int64(d.Seconds() * float64(limit))
}

type Limiter struct {
	limit    Limit
	burst    int64
	lastTime atomic.Int64
}

func NewLimiter(r Limit, b int64) *Limiter { return &Limiter{limit: r, burst: b} }
func (l *Limiter) AllowN(now unixtime.Nano, amount int64) int64 {
	if amount == 0 {
		return 0
	}
	if amount > l.burst {
		amount = l.burst
	}
	for i := 0; i < 10; i++ {
		lastTime := unixtime.Nano(l.lastTime.Load())
		elapsed := now.Sub(lastTime)
		if elapsed <= 0 {
			return 0
		}
		tokens := l.limit.tokensFromDuration(elapsed)
		if tokens <= 0 {
			return 0
		}
		var newTime unixtime.Nano
		if tokens > l.burst {
			tokens = amount
			newTime = now
			if rem := l.burst - amount; rem > 0 {
				newTime = newTime.Add(-l.limit.durationFromTokens(rem))
			}
		} else {
			if tokens > amount {
				tokens = amount
			}
			newTime = lastTime.Add(l.limit.durationFromTokens(tokens))
		}
		if l.lastTime.CompareAndSwap(int64(lastTime), int64(newTime)) {
			return tokens
		}
	}
	return 0
}

type LimiterMap struct {
	new   func(key uint64) *Limiter
	table *xsync.MapOf[uint64, *Limiter]
}

func NewLimiterMap(new func(key uint64) *Limiter) *LimiterMap {
	return &LimiterMap{new: new, table: xsync.NewMapOf[uint64, *Limiter]()}
}
func (m *LimiterMap) Get(key uint64) *Limiter {
	if lim, ok := m.table.Load(key); ok {
		return lim
	}
	lim := m.new(key)
	if found, ok := m.table.LoadOrStore(key, lim); ok {
		return found
	}
	return lim
}
