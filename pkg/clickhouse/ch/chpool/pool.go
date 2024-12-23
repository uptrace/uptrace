package chpool

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var ErrClosed = errors.New("ch: database is closed")
var timers = sync.Pool{New: func() any {
	t := time.NewTimer(time.Hour)
	t.Stop()
	return t
}}

type BadConnError struct{ wrapped error }

var _ error = (*BadConnError)(nil)

func (e BadConnError) Error() string {
	s := "ch: Conn is in a bad state"
	if e.wrapped != nil {
		s += ": " + e.wrapped.Error()
	}
	return s
}
func (e BadConnError) Unwrap() error { return e.wrapped }

type Stats struct {
	Hits       uint32
	Misses     uint32
	Timeouts   uint32
	TotalConns uint32
	IdleConns  uint32
	StaleConns uint32
}
type Pooler interface {
	NewConn(context.Context) (*Conn, error)
	CloseConn(*Conn) error
	Get(context.Context) (*Conn, error)
	Put(*Conn)
	Remove(*Conn, error)
	Len() int
	IdleLen() int
	Stats() *Stats
	Close() error
}
type Config struct {
	Dialer          func(context.Context) (net.Conn, error)
	OnClose         func(*Conn) error
	PoolSize        int
	PoolTimeout     time.Duration
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}
type ConnPool struct {
	conf          *Config
	dialErrorsNum atomic.Uint32
	lastDialErr   atomic.Value
	closed        atomic.Uint32
	queue         chan struct{}
	stats         Stats
	connsMu       sync.Mutex
	conns         []*Conn
	idleConns     []*Conn
}

var _ Pooler = (*ConnPool)(nil)

func New(conf *Config) *ConnPool {
	p := &ConnPool{conf: conf, queue: make(chan struct{}, conf.PoolSize), conns: make([]*Conn, 0, conf.PoolSize), idleConns: make([]*Conn, 0, conf.PoolSize)}
	for i := 0; i < conf.PoolSize; i++ {
		p.queue <- struct{}{}
	}
	return p
}
func (p *ConnPool) NewConn(ctx context.Context) (*Conn, error) {
	cn, err := p.dialConn(ctx)
	if err != nil {
		return nil, err
	}
	p.connsMu.Lock()
	p.conns = append(p.conns, cn)
	p.connsMu.Unlock()
	return cn, nil
}
func (p *ConnPool) dialConn(ctx context.Context) (*Conn, error) {
	if p.isClosed() {
		return nil, ErrClosed
	}
	if p.dialErrorsNum.Load() >= uint32(p.conf.PoolSize) {
		if h, ok := p.lastDialErr.Load().(errorHolder); ok && h.error != nil {
			return nil, h.error
		}
	}
	netConn, err := p.conf.Dialer(ctx)
	if err != nil {
		p.lastDialErr.Store(errorHolder{err})
		if p.dialErrorsNum.Add(1) == uint32(p.conf.PoolSize) {
			go p.tryDial()
		}
		return nil, err
	}
	cn := NewConn(netConn)
	return cn, nil
}
func (p *ConnPool) tryDial() {
	for {
		if p.isClosed() {
			return
		}
		conn, err := p.conf.Dialer(context.Background())
		if err != nil {
			slog.Error("ch: tryDial failed", slog.Any("error", err))
			p.lastDialErr.Store(errorHolder{err})
			time.Sleep(time.Second)
			continue
		}
		p.dialErrorsNum.Store(0)
		_ = conn.Close()
		return
	}
}
func (p *ConnPool) Get(ctx context.Context) (*Conn, error) {
	if p.isClosed() {
		return nil, ErrClosed
	}
	if err := p.waitTurn(ctx); err != nil {
		return nil, err
	}
	for {
		p.connsMu.Lock()
		cn := p.popIdle()
		p.connsMu.Unlock()
		if cn == nil {
			break
		}
		if !p.isHealthyConn(cn) {
			_ = p.CloseConn(cn)
			continue
		}
		atomic.AddUint32(&p.stats.Hits, 1)
		return cn, nil
	}
	atomic.AddUint32(&p.stats.Misses, 1)
	newcn, err := p.NewConn(ctx)
	if err != nil {
		_ = p.freeTurn()
		return nil, err
	}
	return newcn, nil
}
func (p *ConnPool) waitTurn(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	select {
	case <-p.queue:
		return nil
	default:
	}
	timer := timers.Get().(*time.Timer)
	timer.Reset(p.conf.PoolTimeout)
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		timers.Put(timer)
		return ctx.Err()
	case <-p.queue:
		if !timer.Stop() {
			<-timer.C
		}
		timers.Put(timer)
		return nil
	case <-timer.C:
		timers.Put(timer)
		atomic.AddUint32(&p.stats.Timeouts, 1)
		return nil
	}
}
func (p *ConnPool) freeTurn() bool {
	select {
	case p.queue <- struct{}{}:
		return true
	default:
		return false
	}
}
func (p *ConnPool) popIdle() *Conn {
	if len(p.idleConns) == 0 {
		return nil
	}
	idx := len(p.idleConns) - 1
	cn := p.idleConns[idx]
	p.idleConns = p.idleConns[:idx]
	return cn
}
func (p *ConnPool) Put(cn *Conn) {
	if cn.rd.Buffered() > 0 {
		slog.Warn("ch: Conn has unread data")
		p.Remove(cn, BadConnError{})
		return
	}
	var shouldCloseConn bool
	withLock(&p.connsMu, func() {
		if p.conf.MaxIdleConns == 0 || len(p.idleConns) < p.conf.MaxIdleConns {
			p.idleConns = append(p.idleConns, cn)
		} else {
			p.removeConn(cn)
			shouldCloseConn = true
		}
	})
	if !p.freeTurn() || shouldCloseConn {
		_ = p.closeConn(cn)
	}
}
func (p *ConnPool) Remove(cn *Conn, reason error) {
	p.removeConnAcquiringLock(cn)
	_ = p.freeTurn()
	_ = p.closeConn(cn)
}
func (p *ConnPool) CloseConn(cn *Conn) error {
	p.removeConnAcquiringLock(cn)
	return p.closeConn(cn)
}
func (p *ConnPool) removeConnAcquiringLock(cn *Conn) {
	p.connsMu.Lock()
	p.removeConn(cn)
	p.connsMu.Unlock()
}
func (p *ConnPool) removeConn(needle *Conn) {
	for i, found := range p.conns {
		if found == needle {
			p.conns = append(p.conns[:i], p.conns[i+1:]...)
			return
		}
	}
}
func (p *ConnPool) closeConn(cn *Conn) error {
	if p.conf.OnClose != nil {
		_ = p.conf.OnClose(cn)
	}
	return cn.Close()
}
func (p *ConnPool) Len() int {
	p.connsMu.Lock()
	n := len(p.conns)
	p.connsMu.Unlock()
	return n
}
func (p *ConnPool) IdleLen() int {
	p.connsMu.Lock()
	n := len(p.idleConns)
	p.connsMu.Unlock()
	return n
}
func (p *ConnPool) Stats() *Stats {
	return &Stats{Hits: atomic.LoadUint32(&p.stats.Hits), Misses: atomic.LoadUint32(&p.stats.Misses), Timeouts: atomic.LoadUint32(&p.stats.Timeouts), TotalConns: uint32(p.Len()), IdleConns: uint32(p.IdleLen()), StaleConns: atomic.LoadUint32(&p.stats.StaleConns)}
}
func (p *ConnPool) isClosed() bool { return p.closed.Load() != 0 }
func (p *ConnPool) Close() error {
	if !p.closed.CompareAndSwap(0, 1) {
		return ErrClosed
	}
	var firstErr error
	withLock(&p.connsMu, func() {
		for _, cn := range p.conns {
			if err := p.closeConn(cn); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		p.conns = nil
		p.idleConns = nil
	})
	return firstErr
}
func (p *ConnPool) isHealthyConn(cn *Conn) bool {
	now := time.Now()
	if p.conf.ConnMaxLifetime > 0 && now.Sub(cn.createdAt) >= p.conf.ConnMaxLifetime {
		return false
	}
	if p.conf.ConnMaxIdleTime > 0 && now.Sub(cn.UsedAt()) >= p.conf.ConnMaxIdleTime {
		atomic.AddUint32(&p.stats.IdleConns, 1)
		return false
	}
	if err := connCheck(cn.netConn); err != nil {
		slog.Warn("ch: connCheck failed", "error", err)
		return false
	}
	cn.SetUsedAt(now)
	return true
}
func withLock(mu *sync.Mutex, fn func()) {
	mu.Lock()
	defer mu.Unlock()
	fn()
}

type errorHolder struct{ error }
