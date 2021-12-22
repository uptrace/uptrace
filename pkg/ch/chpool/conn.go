package chpool

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	"github.com/uptrace/go-clickhouse/ch/chproto"
)

var noDeadline = time.Time{}

type Conn struct {
	netConn net.Conn

	rd *chproto.Reader
	wr *chproto.Writer

	ServerInfo chproto.ServerInfo

	pooled    bool
	Inited    bool
	createdAt time.Time
	usedAt    int64  // atomic
	closed    uint32 // atomic
}

func NewConn(netConn net.Conn) *Conn {
	cn := &Conn{
		netConn:   netConn,
		rd:        chproto.NewReader(netConn),
		wr:        chproto.NewWriter(netConn),
		createdAt: time.Now(),
	}
	return cn
}

func (cn *Conn) UsedAt() time.Time {
	unix := atomic.LoadInt64(&cn.usedAt)
	return time.Unix(unix, 0)
}

func (cn *Conn) SetUsedAt(tm time.Time) {
	atomic.StoreInt64(&cn.usedAt, tm.Unix())
}

func (cn *Conn) RemoteAddr() net.Addr {
	return cn.netConn.RemoteAddr()
}

func (cn *Conn) WithReader(
	ctx context.Context,
	timeout time.Duration,
	fn func(rd *chproto.Reader) error,
) error {
	if err := cn.netConn.SetReadDeadline(cn.deadline(ctx, timeout)); err != nil {
		return err
	}

	if err := fn(cn.rd); err != nil {
		return err
	}

	return nil
}

func (cn *Conn) WithWriter(
	ctx context.Context,
	timeout time.Duration,
	fn func(wb *chproto.Writer),
) error {
	if err := cn.netConn.SetWriteDeadline(cn.deadline(ctx, timeout)); err != nil {
		return err
	}

	fn(cn.wr)

	if err := cn.wr.Flush(); err != nil {
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

func (cn *Conn) Closed() bool {
	return atomic.LoadUint32(&cn.closed) == 1
}

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
