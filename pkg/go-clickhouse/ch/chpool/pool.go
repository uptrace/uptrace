package chpool

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uptrace/go-clickhouse/ch/internal"
)

var ErrClosed = errors.New("ch: database is closed")

var timers = sync.Pool{
	New: func() any {
		t := time.NewTimer(time.Hour)
		t.Stop()
		return t
	},
}

//------------------------------------------------------------------------------

type BadConnError struct {
	wrapped error
}

var _ error = (*BadConnError)(nil)

func (e BadConnError) Error() string {
	s := "ch: Conn is in a bad state"
	if e.wrapped != nil {
		s += ": " + e.wrapped.Error()
	}
	return s
}

func (e BadConnError) Unwrap() error {
	return e.wrapped
}

//------------------------------------------------------------------------------

// Stats contains pool state information and accumulated stats.
type Stats struct {
	Hits     uint32 // number of times free connection was found in the pool
	Misses   uint32 // number of times free connection was NOT found in the pool
	Timeouts uint32 // number of times a wait timeout occurred

	TotalConns uint32 // number of total connections in the pool
	IdleConns  uint32 // number of idle connections in the pool
	StaleConns uint32 // number of stale connections removed from the pool
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
	Dialer  func(context.Context) (net.Conn, error)
	OnClose func(*Conn) error

	PoolSize        int
	PoolTimeout     time.Duration
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}

type ConnPool struct {
	conf *Config

	dialErrorsNum uint32 // atomic

	_closed uint32 // atomic

	lastDialErrorMu sync.RWMutex
	lastDialError   error

	queue chan struct{}

	stats Stats

	connsMu   sync.Mutex
	conns     []*Conn
	idleConns []*Conn
}

var _ Pooler = (*ConnPool)(nil)

func New(conf *Config) *ConnPool {
	p := &ConnPool{
		conf: conf,

		queue:     make(chan struct{}, conf.PoolSize),
		conns:     make([]*Conn, 0, conf.PoolSize),
		idleConns: make([]*Conn, 0, conf.PoolSize),
	}

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
	if p.closed() {
		return nil, ErrClosed
	}

	if atomic.LoadUint32(&p.dialErrorsNum) >= uint32(p.conf.PoolSize) {
		return nil, p.getLastDialError()
	}

	netConn, err := p.conf.Dialer(ctx)
	if err != nil {
		p.setLastDialError(err)
		if atomic.AddUint32(&p.dialErrorsNum, 1) == uint32(p.conf.PoolSize) {
			go p.tryDial()
		}
		return nil, err
	}

	cn := NewConn(netConn)
	return cn, nil
}

func (p *ConnPool) tryDial() {
	for {
		if p.closed() {
			return
		}

		conn, err := p.conf.Dialer(context.TODO())
		if err != nil {
			p.setLastDialError(err)
			time.Sleep(time.Second)
			continue
		}

		atomic.StoreUint32(&p.dialErrorsNum, 0)
		_ = conn.Close()
		return
	}
}

func (p *ConnPool) setLastDialError(err error) {
	p.lastDialErrorMu.Lock()
	p.lastDialError = err
	p.lastDialErrorMu.Unlock()
}

func (p *ConnPool) getLastDialError() error {
	p.lastDialErrorMu.RLock()
	err := p.lastDialError
	p.lastDialErrorMu.RUnlock()
	return err
}

// Get returns an existing connection from the pool or creates a new one.
func (p *ConnPool) Get(ctx context.Context) (*Conn, error) {
	if p.closed() {
		return nil, ErrClosed
	}

	err := p.waitTurn(ctx)
	if err != nil {
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
		internal.Logger.Printf("Conn has unread data")
		p.Remove(cn, BadConnError{})
		return
	}

	var shouldCloseConn bool

	p.connsMu.Lock()

	if p.conf.MaxIdleConns == 0 || len(p.idleConns) < p.conf.MaxIdleConns {
		p.idleConns = append(p.idleConns, cn)
	} else {
		p.removeConn(cn)
		shouldCloseConn = true
	}

	p.connsMu.Unlock()

	if !p.freeTurn() || shouldCloseConn {
		_ = p.closeConn(cn)
	}
}

func (p *ConnPool) Remove(cn *Conn, reason error) {
	p.removeConnWithLock(cn)
	_ = p.freeTurn()
	_ = p.closeConn(cn)
}

func (p *ConnPool) CloseConn(cn *Conn) error {
	p.removeConnWithLock(cn)
	return p.closeConn(cn)
}

func (p *ConnPool) removeConnWithLock(cn *Conn) {
	p.connsMu.Lock()
	defer p.connsMu.Unlock()
	p.removeConn(cn)
}

func (p *ConnPool) removeConn(cn *Conn) {
	for i, c := range p.conns {
		if c == cn {
			p.conns = append(p.conns[:i], p.conns[i+1:]...)
			break
		}
	}
}

func (p *ConnPool) closeConn(cn *Conn) error {
	if p.conf.OnClose != nil {
		_ = p.conf.OnClose(cn)
	}
	return cn.Close()
}

// Len returns total number of connections.
func (p *ConnPool) Len() int {
	p.connsMu.Lock()
	n := len(p.conns)
	p.connsMu.Unlock()
	return n
}

// IdleLen returns number of idle connections.
func (p *ConnPool) IdleLen() int {
	p.connsMu.Lock()
	n := len(p.idleConns)
	p.connsMu.Unlock()
	return n
}

func (p *ConnPool) Stats() *Stats {
	return &Stats{
		Hits:     atomic.LoadUint32(&p.stats.Hits),
		Misses:   atomic.LoadUint32(&p.stats.Misses),
		Timeouts: atomic.LoadUint32(&p.stats.Timeouts),

		TotalConns: uint32(p.Len()),
		IdleConns:  uint32(p.IdleLen()),
		StaleConns: atomic.LoadUint32(&p.stats.StaleConns),
	}
}

func (p *ConnPool) closed() bool {
	return atomic.LoadUint32(&p._closed) == 1
}

func (p *ConnPool) Close() error {
	if !atomic.CompareAndSwapUint32(&p._closed, 0, 1) {
		return ErrClosed
	}

	var firstErr error
	p.connsMu.Lock()
	for _, cn := range p.conns {
		if err := p.closeConn(cn); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	p.conns = nil
	p.idleConns = nil
	p.connsMu.Unlock()

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
		return false
	}

	cn.SetUsedAt(now)
	return true
}
