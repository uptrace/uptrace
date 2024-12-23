package ch

import (
	"context"
	"crypto/tls"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/uptrace/pkg/clickhouse/ch/chpool"
	"github.com/uptrace/pkg/clickhouse/ch/chproto"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"github.com/uptrace/pkg/clickhouse/ch/internal"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

type DBStats struct {
	Queries uint64
	Errors  uint64
}
type DB struct {
	conf       *Config
	pool       *chpool.ConnPool
	queryHooks []QueryHook
	fmter      chschema.Formatter
	flags      internal.Flag
	stats      DBStats
}

func Connect(opts ...Option) *DB {
	db := &DB{conf: defaultConfig()}
	for _, opt := range opts {
		opt(db)
	}
	db.pool = newConnPool(db.conf)
	for k, v := range db.conf.NamedArgs {
		db.fmter = db.fmter.WithNamedArg(k, v)
	}
	return db
}
func newConnPool(conf *Config) *chpool.ConnPool {
	poolConf := conf.Config
	dialer := conf.netDialer()
	poolConf.Dialer = func(ctx context.Context) (net.Conn, error) {
		ctx, span := internal.Tracer.Start(ctx, "ch.Dial")
		defer span.End()
		if conf.TLSConfig != nil {
			return tls.DialWithDialer(dialer, conf.Network, conf.Addr, conf.TLSConfig)
		}
		return dialer.DialContext(ctx, conf.Network, conf.Addr)
	}
	return chpool.New(&poolConf)
}
func (db *DB) Close() error         { return db.pool.Close() }
func (db *DB) String() string       { return fmt.Sprintf("DB<addr: %s>", db.conf.Addr) }
func (db *DB) Equal(other *DB) bool { return db.pool == other.pool }
func (db *DB) Config() *Config      { return db.conf }
func (db *DB) WithTimeout(d time.Duration) *DB {
	newconf := *db.conf
	newconf.ReadTimeout = d
	newconf.WriteTimeout = d
	clone := db.clone()
	clone.conf = &newconf
	return clone
}
func (db *DB) clone() *DB {
	clone := *db
	l := len(db.queryHooks)
	clone.queryHooks = db.queryHooks[:l:l]
	return &clone
}
func (db *DB) Stats() DBStats {
	return DBStats{Queries: atomic.LoadUint64(&db.stats.Queries), Errors: atomic.LoadUint64(&db.stats.Errors)}
}
func (db *DB) AllDistTable(tableName string) string {
	if db.conf.Distributed {
		return "dist_all_" + tableName
	}
	return tableName
}
func (db *DB) DistTable(tableName string) string {
	if db.conf.Distributed {
		return "dist_" + tableName
	}
	return tableName
}
func (db *DB) getConn(ctx context.Context) (*chpool.Conn, error) {
	cn, err := db.pool.Get(ctx)
	if err != nil {
		return nil, err
	}
	if err := db.initConn(ctx, cn); err != nil {
		db.pool.Remove(cn, err)
		if err := internal.Unwrap(err); err != nil {
			return nil, err
		}
		return nil, err
	}
	return cn, nil
}
func (db *DB) initConn(ctx context.Context, cn *chpool.Conn) error {
	return cn.InitOnce(func() error {
		if err := db.hello(ctx, cn); err != nil {
			return err
		}
		return nil
	})
}
func (db *DB) releaseConn(cn *chpool.Conn, err error) {
	if isBadConn(err, false) || cn.Closed() {
		db.pool.Remove(cn, err)
	} else {
		db.pool.Put(cn)
	}
}
func (db *DB) withConn(ctx context.Context, fn func(*chpool.Conn) error) error {
	err := db._withConn(ctx, fn)
	atomic.AddUint64(&db.stats.Queries, 1)
	if err != nil {
		atomic.AddUint64(&db.stats.Errors, 1)
	}
	return err
}
func (db *DB) _withConn(ctx context.Context, fn func(*chpool.Conn) error) error {
	cn, release, err := db.useConn(ctx)
	if err != nil {
		return err
	}
	defer func() { release(err) }()
	err = fn(cn)
	return err
}
func (db *DB) useConn(ctx context.Context) (*chpool.Conn, func(err error), error) {
	cn, err := db.getConn(ctx)
	if err != nil {
		return nil, nil, err
	}
	var doneOrCanceled chan struct{}
	if ctxDone := ctx.Done(); ctxDone != nil {
		doneOrCanceled = make(chan struct{})
		go func() {
			select {
			case <-doneOrCanceled:
			case <-ctxDone:
				db.cancelConn(ctx, cn)
				doneOrCanceled <- struct{}{}
			}
		}()
	}
	release := func(err error) {
		var canceled bool
		if doneOrCanceled != nil {
			select {
			case <-doneOrCanceled:
				canceled = true
			case doneOrCanceled <- struct{}{}:
			}
		}
		if !canceled {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				db.cancelConn(ctx, cn)
				canceled = true
			}
		}
		if canceled {
			db.pool.Remove(cn, err)
			return
		}
		db.releaseConn(cn, err)
	}
	return cn, release, err
}
func (db *DB) cancelConn(ctx context.Context, cn *chpool.Conn) {
	ctx = context.WithoutCancel(ctx)
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		ctx, span = internal.Tracer.Start(ctx, "ch-write-cancel")
		defer span.End()
	}
	if err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) { writeCancel(wr) }); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
func (db *DB) Ping(ctx context.Context) error {
	return db.withConn(ctx, func(cn *chpool.Conn) error {
		if err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) { writePing(wr) }); err != nil {
			return err
		}
		return cn.WithReader(ctx, db.conf.ReadTimeout, func(rd *chproto.Reader) error { return readPong(rd) })
	})
}
func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.ExecContext(context.Background(), query, args...)
}
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	query = db.FormatQuery(query, args...)
	ctx, evt, err := db.beforeQuery(ctx, nil, query, args, nil)
	if err != nil {
		return nil, err
	}
	res, err := db.exec(ctx, query)
	db.afterQuery(ctx, evt, res, err)
	return res, err
}
func (db *DB) exec(ctx context.Context, query string) (*Result, error) {
	var res *Result
	var lastErr error
	for attempt := 0; attempt <= db.conf.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := internal.Sleep(ctx, db.retryBackoff()); err != nil {
				return nil, err
			}
		}
		res, lastErr = db._exec(ctx, query)
		if lastErr == nil || !db.shouldRetryError(lastErr) {
			break
		}
		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			span.AddEvent("log", trace.WithAttributes(attribute.String("log.severity", "INFO"), attribute.String("log.message", "retrying exec"), attribute.String("error", lastErr.Error())))
		}
	}
	return res, lastErr
}
func (db *DB) _exec(ctx context.Context, query string) (*Result, error) {
	var res *Result
	err := db.withConn(ctx, func(cn *chpool.Conn) error {
		if err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) { db.writeQuery(ctx, cn, wr, query); db.writeBlock(ctx, wr, nil) }); err != nil {
			return err
		}
		return cn.WithReader(ctx, db.conf.ReadTimeout, func(rd *chproto.Reader) error { var err error; res, err = db.readDataBlocks(cn, rd); return err })
	})
	return res, err
}
func (db *DB) QueryRow(query string, args ...any) *Row {
	return db.QueryRowContext(context.Background(), query, args...)
}
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *Row {
	rows, err := db.QueryContext(ctx, query, args...)
	return &Row{rows: rows, err: err}
}
func (db *DB) Query(query string, args ...any) (*Rows, error) {
	return db.QueryContext(context.Background(), query, args...)
}
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*Rows, error) {
	blocks, err := db.QueryBlocks(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return newRows(ctx, blocks), nil
}
func (db *DB) QueryBlocks(ctx context.Context, query string, args ...any) (*BlockIter, error) {
	query = db.FormatQuery(query, args...)
	var blocks *BlockIter
	var lastErr error
	for attempt := 0; attempt <= db.conf.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := internal.Sleep(ctx, db.retryBackoff()); err != nil {
				return nil, err
			}
		}
		ctx, evt, err := db.beforeQuery(ctx, nil, query, args, nil)
		if err != nil {
			return nil, err
		}
		blocks, lastErr = db._query(ctx, query)
		db.afterQuery(ctx, evt, nil, lastErr)
		if lastErr == nil || !db.shouldRetryError(lastErr) {
			break
		}
		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			span.AddEvent("log", trace.WithAttributes(attribute.String("log.severity", "INFO"), attribute.String("log.message", "retrying query"), attribute.String("error", lastErr.Error())))
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return blocks, nil
}
func (db *DB) _query(ctx context.Context, query string) (*BlockIter, error) {
	cn, release, err := db.useConn(ctx)
	if err != nil {
		return nil, err
	}
	if err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) { db.writeQuery(ctx, cn, wr, query); db.writeBlock(ctx, wr, nil) }); err != nil {
		release(err)
		return nil, err
	}
	blocks, err := newBlockIter(ctx, db, cn, release)
	if err != nil {
		release(err)
		return nil, err
	}
	return blocks, nil
}
func (db *DB) insert(ctx context.Context, model TableModel, query string, block *Block) (*Result, error) {
	var res *Result
	var lastErr error
	for attempt := 0; attempt <= db.conf.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := internal.Sleep(ctx, db.retryBackoff()); err != nil {
				return nil, err
			}
		}
		res, lastErr = db._insert(ctx, model, query, block)
		if lastErr == nil || !db.shouldRetryError(lastErr) {
			break
		}
		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			span.AddEvent("log", trace.WithAttributes(attribute.String("log.severity", "INFO"), attribute.String("log.message", "retrying insert"), attribute.String("error", lastErr.Error())))
		}
	}
	return res, lastErr
}
func (db *DB) _insert(ctx context.Context, model TableModel, query string, block *Block) (*Result, error) {
	var res *Result
	err := db.withConn(ctx, func(cn *chpool.Conn) error {
		if err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) { db.writeQuery(ctx, cn, wr, query); db.writeBlock(ctx, wr, nil) }); err != nil {
			return err
		}
		if err := cn.WithReader(ctx, db.conf.ReadTimeout, func(rd *chproto.Reader) error { return db.readSampleBlock(rd) }); err != nil {
			return err
		}
		if err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) { db.writeBlock(ctx, wr, block); db.writeBlock(ctx, wr, nil) }); err != nil {
			return err
		}
		return cn.WithReader(ctx, db.conf.ReadTimeout, func(rd *chproto.Reader) error {
			var err error
			res, err = readInsertResult(cn, rd)
			if err != nil {
				return err
			}
			res.affected = int64(block.NumRow)
			return nil
		})
	})
	return res, err
}
func (db *DB) NewSelect() *SelectQuery               { return NewSelectQuery(db) }
func (db *DB) NewWindow() *WindowQuery               { return NewWindowQuery(db) }
func (db *DB) NewInsert() *InsertQuery               { return NewInsertQuery(db) }
func (db *DB) NewAlterDelete() *AlterDeleteQuery     { return NewAlterDeleteQuery(db) }
func (db *DB) NewCreateTable() *CreateTableQuery     { return NewCreateTableQuery(db) }
func (db *DB) NewDropTable() *DropTableQuery         { return NewDropTableQuery(db) }
func (db *DB) NewTruncateTable() *TruncateTableQuery { return NewTruncateTableQuery(db) }
func (db *DB) ResetModel(ctx context.Context, models ...any) error {
	for _, model := range models {
		if _, err := db.NewDropTable().Model(model).IfExists().Exec(ctx); err != nil {
			return err
		}
		if _, err := db.NewCreateTable().Model(model).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
func (db *DB) WithNamedArg(name string, value interface{}) *DB {
	clone := db.clone()
	clone.fmter = clone.fmter.WithNamedArg(name, value)
	return clone
}
func (db *DB) Formatter() chschema.Formatter { return db.fmter }
func (db *DB) WithFormatter(fmter chschema.Formatter) *DB {
	clone := db.clone()
	clone.fmter = fmter
	return clone
}
func (db *DB) shouldRetryError(err error) bool {
	switch err {
	case nil, context.Canceled, context.DeadlineExceeded:
		return false
	case driver.ErrBadConn:
		return true
	}
	if err, ok := err.(*Error); ok {
		const (
			timeoutExceeded            = 159
			tooSlow                    = 160
			tooManySimultaneousQueries = 202
			cannotDecompress           = 271
		)
		switch err.Code {
		case timeoutExceeded, tooSlow, tooManySimultaneousQueries, cannotDecompress:
			return true
		}
	}
	return false
}
func (db *DB) retryBackoff() time.Duration {
	return internal.RetryBackoff(db.conf.MinRetryBackoff, db.conf.MaxRetryBackoff)
}
func (db *DB) FormatQuery(query string, args ...any) string {
	return db.fmter.FormatQuery(query, args...)
}
func (db *DB) makeQueryBytes() []byte { return make([]byte, 0, 4096) }

type Rows struct {
	ctx      context.Context
	blocks   *BlockIter
	block    Block
	rowIndex int
	hasNext  bool
	closed   bool
}

func newRows(ctx context.Context, blocks *BlockIter) *Rows { return &Rows{ctx: ctx, blocks: blocks} }
func (rs *Rows) Close() error {
	if rs.closed {
		return nil
	}
	for rs.blocks.Next(&rs.block) {
	}
	rs.close()
	return nil
}
func (rs *Rows) close()                                  { rs.closed = true; rs.block = Block{}; _ = rs.blocks.Close() }
func (rs *Rows) ColumnTypes() ([]*sql.ColumnType, error) { return nil, errors.New("not implemented") }
func (rs *Rows) Columns() ([]string, error)              { return nil, errors.New("not implemented") }
func (rs *Rows) Err() error                              { return rs.blocks.Err() }
func (rs *Rows) Next() bool {
	if rs.closed {
		return false
	}
	for rs.rowIndex >= rs.block.NumRow {
		if !rs.blocks.Next(&rs.block) {
			rs.close()
			return false
		}
		rs.rowIndex = 0
	}
	rs.hasNext = true
	rs.rowIndex++
	return true
}
func (rs *Rows) NextResultSet() bool { return false }
func (rs *Rows) Scan(dest ...any) error {
	if rs.closed {
		return rs.Err()
	}
	if !rs.hasNext {
		return errors.New("ch: Scan called without calling Next")
	}
	rs.hasNext = false
	if rs.block.NumColumn != len(dest) {
		return fmt.Errorf("ch: got %d columns, but Scan has %d values", rs.block.NumColumn, len(dest))
	}
	for i, col := range rs.block.Columns {
		x := dest[i]
		typ := reflect.TypeOf(x).Elem()
		ptr := (*eface)(unsafe.Pointer(&x)).ptr
		if err := col.ConvertAssign(rs.rowIndex-1, typ, ptr); err != nil {
			return err
		}
		runtime.KeepAlive(x)
	}
	return nil
}
func (rs *Rows) Result() *Result { return rs.blocks.Result() }

type Row struct {
	rows *Rows
	err  error
}

func (r *Row) Err() error { return r.err }
func (r *Row) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	defer r.rows.Close()
	if r.rows.Next() {
		return r.rows.Scan(dest...)
	}
	return sql.ErrNoRows
}
