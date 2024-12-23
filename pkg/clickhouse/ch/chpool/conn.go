package chpool

import (
	"context"
	"database/sql/driver"
	"errors"
	"github.com/uptrace/pkg/clickhouse/ch/chproto"
	"net"
	"sync/atomic"
	"syscall"
	"time"
)

var noDeadline = time.Time{}

type Conn struct {
	netConn    net.Conn
	rd         *chproto.Reader
	wr         *chproto.Writer
	createdAt  time.Time
	usedAt     int64
	closed     uint32
	initOnce   doOnce
	ServerInfo chproto.ServerInfo
}

func NewConn(netConn net.Conn) *Conn {
	cn := &Conn{netConn: netConn, rd: chproto.NewReader(netConn), wr: chproto.NewWriter(netConn), createdAt: time.Now()}
	cn.SetUsedAt(time.Now())
	return cn
}
func (cn *Conn) InitOnce(fn func() error) error {
	if err := cn.initOnce.Do(fn); err != nil {
		return err
	}
	return nil
}
func (cn *Conn) UsedAt() time.Time {
	unix := atomic.LoadInt64(&cn.usedAt)
	return time.Unix(unix, 0)
}
func (cn *Conn) SetUsedAt(tm time.Time) { atomic.StoreInt64(&cn.usedAt, tm.Unix()) }
func (cn *Conn) LocalAddr() net.Addr    { return cn.netConn.LocalAddr() }
func (cn *Conn) RemoteAddr() net.Addr   { return cn.netConn.RemoteAddr() }
func (cn *Conn) Reader(ctx context.Context, timeout time.Duration) *chproto.Reader {
	_ = cn.netConn.SetReadDeadline(cn.deadline(ctx, timeout))
	return cn.rd
}
func (cn *Conn) WithReader(ctx context.Context, timeout time.Duration, fn func(rd *chproto.Reader) error) error {
	if err := cn.netConn.SetReadDeadline(cn.deadline(ctx, timeout)); err != nil {
		return err
	}
	return fn(cn.rd)
}
func (cn *Conn) WithWriter(ctx context.Context, timeout time.Duration, fn func(wb *chproto.Writer)) error {
	if err := cn.netConn.SetWriteDeadline(cn.deadline(ctx, timeout)); err != nil {
		return err
	}
	fn(cn.wr)
	if err := cn.wr.Flush(); err != nil {
		if errors.Is(err, syscall.EPIPE) {
			return driver.ErrBadConn
		}
		return err
	}
	return nil
}
func (cn *Conn) Close() error {
	if !atomic.CompareAndSwapUint32(&cn.closed, 0, 1) {
		return nil
	}
	return cn.netConn.Close()
}
func (cn *Conn) Closed() bool { return atomic.LoadUint32(&cn.closed) == 1 }
func (cn *Conn) deadline(ctx context.Context, timeout time.Duration) time.Time {
	tm := time.Now()
	cn.SetUsedAt(tm)
	if timeout > 0 {
		tm = tm.Add(timeout)
	}
	if ctx != nil {
		deadline, ok := ctx.Deadline()
		if ok {
			if timeout == 0 {
				return deadline
			}
			if deadline.Before(tm) {
				return deadline
			}
			return tm
		}
	}
	if timeout > 0 {
		return tm
	}
	return noDeadline
}

type doOnce struct{ done bool }

func (o *doOnce) Do(fn func() error) error {
	if o.done {
		return nil
	}
	o.done = true
	return fn()
}
