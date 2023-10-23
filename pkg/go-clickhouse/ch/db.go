package ch

import (
	"context"
	"crypto/tls"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/rs/dnscache"
	"github.com/uptrace/go-clickhouse/ch/chpool"
	"github.com/uptrace/go-clickhouse/ch/chproto"
	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/go-clickhouse/ch/internal"
)

type DBStats struct {
	Queries uint64
	Errors  uint64
}

type DB struct {
	conf *Config
	pool *chpool.ConnPool

	queryHooks []QueryHook

	fmter chschema.Formatter
	flags internal.Flag
	stats DBStats
}

func Connect(opts ...Option) *DB {
	db := newDB(defaultConfig(), opts...)
	if db.flags.Has(autoCreateDatabaseFlag) {
		db.autoCreateDatabase()
	}
	return db
}

func newDB(conf *Config, opts ...Option) *DB {
	db := &DB{
		conf: conf,
	}
	for _, opt := range opts {
		opt(db)
	}
	db.pool = newConnPool(db.conf)
	return db
}

func newConnPool(conf *Config) *chpool.ConnPool {
	poolConf := conf.Config

	resolver := new(dnscache.Resolver)
	dialer := conf.netDialer()
	poolConf.Dialer = func(ctx context.Context) (conn net.Conn, err error) {
		host, port, err := net.SplitHostPort(conf.Addr)
		if err != nil {
			return nil, err
		}

		ips, err := resolver.LookupHost(ctx, host)
		if err != nil {
			return nil, err
		}

		for _, ip := range ips {
			if conf.TLSConfig != nil {
				conn, err = tls.DialWithDialer(
					dialer,
					"tcp",
					net.JoinHostPort(ip, port),
					conf.TLSConfig,
				)
			} else {
				conn, err = dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, port))
			}
			if err == nil {
				break
			}
		}
		return conn, err
	}
	return chpool.New(&poolConf)
}

// Close closes the database client, releasing any open resources.
//
// It is rare to Close a DB, as the DB handle is meant to be
// long-lived and shared between many goroutines.
func (db *DB) Close() error {
	return db.pool.Close()
}

func (db *DB) String() string {
	return fmt.Sprintf("DB<addr: %s>", db.conf.Addr)
}

func (db *DB) Config() *Config {
	return db.conf
}

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
	return DBStats{
		Queries: atomic.LoadUint64(&db.stats.Queries),
		Errors:  atomic.LoadUint64(&db.stats.Errors),
	}
}

func (db *DB) autoCreateDatabase() {
	ctx := context.Background()

	switch err := db.Ping(ctx); err := err.(type) {
	case nil: // all is good
		return
	case *Error:
		if err.Code != 81 { // 81 - database does not exist
			return
		}
	default:
		// ignore the error
		return
	}

	conf := db.conf.clone()
	conf.Database = ""

	query := "CREATE DATABASE IF NOT EXISTS ?"
	if conf.Cluster != "" {
		query += " ON CLUSTER ?"
	}

	tmp := newDB(conf)
	defer tmp.Close()

	if _, err := tmp.Exec(query, Name(db.conf.Database), Name(db.conf.Cluster)); err != nil {
		internal.Logger.Printf("create database %q failed: %s", db.conf.Database, err)
	}
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
	if cn.Inited {
		return nil
	}
	cn.Inited = true

	return db.hello(ctx, cn)
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
	cn, err := db.getConn(ctx)
	if err != nil {
		return err
	}

	var fnErr error

	defer func() {
		db.releaseConn(cn, fnErr)
	}()

	fnErr = fn(cn)
	return fnErr
}

func (db *DB) Ping(ctx context.Context) error {
	return db.withConn(ctx, func(cn *chpool.Conn) error {
		if err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) {
			writePing(wr)
		}); err != nil {
			return err
		}
		return cn.WithReader(ctx, db.conf.ReadTimeout, func(rd *chproto.Reader) error {
			return readPong(rd)
		})
	})
}

func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.ExecContext(context.Background(), query, args...)
}

func (db *DB) ExecContext(
	ctx context.Context, query string, args ...any,
) (sql.Result, error) {
	query = db.FormatQuery(query, args...)
	ctx, evt := db.beforeQuery(ctx, nil, query, args, nil)
	res, err := db.exec(ctx, query)
	db.afterQuery(ctx, evt, res, err)
	return res, err
}

func (db *DB) exec(ctx context.Context, query string) (*result, error) {
	var res *result
	var lastErr error
	for attempt := 0; attempt <= db.conf.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := internal.Sleep(ctx, db.retryBackoff()); err != nil {
				return nil, err
			}
		}

		res, lastErr = db._exec(ctx, query)
		if !db.shouldRetry(lastErr) {
			break
		}
	}
	return res, lastErr
}

func (db *DB) _exec(ctx context.Context, query string) (*result, error) {
	var res *result
	err := db.withConn(ctx, func(cn *chpool.Conn) error {
		if err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) {
			db.writeQuery(ctx, cn, wr, query)
			db.writeBlock(ctx, wr, nil)
		}); err != nil {
			return err
		}
		return cn.WithReader(ctx, db.conf.ReadTimeout, func(rd *chproto.Reader) error {
			var err error
			res, err = db.readDataBlocks(cn, rd)
			return err
		})
	})
	return res, err
}

func (db *DB) Query(query string, args ...any) (*Rows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

func (db *DB) QueryContext(
	ctx context.Context, query string, args ...any,
) (*Rows, error) {
	query = db.FormatQuery(query, args...)

	ctx, evt := db.beforeQuery(ctx, nil, query, args, nil)
	blocks, err := db.query(ctx, query)
	db.afterQuery(ctx, evt, nil, err)
	if err != nil {
		return nil, err
	}

	return newRows(ctx, blocks), nil
}

func (db *DB) QueryRow(query string, args ...any) *Row {
	return db.QueryRowContext(context.Background(), query, args...)
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *Row {
	rows, err := db.QueryContext(ctx, query, args...)
	return &Row{rows: rows, err: err}
}

func (db *DB) query(ctx context.Context, query string) (*blockIter, error) {
	var blocks *blockIter
	var lastErr error

	for attempt := 0; attempt <= db.conf.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := internal.Sleep(ctx, db.retryBackoff()); err != nil {
				return nil, err
			}
		}

		blocks, lastErr = db._query(ctx, query)
		if !db.shouldRetry(lastErr) {
			break
		}
	}

	return blocks, lastErr
}

func (db *DB) _query(ctx context.Context, query string) (*blockIter, error) {
	cn, err := db.getConn(ctx)
	if err != nil {
		return nil, err
	}

	if err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) {
		db.writeQuery(ctx, cn, wr, query)
		db.writeBlock(ctx, wr, nil)
	}); err != nil {
		db.releaseConn(cn, err)
		return nil, err
	}

	return newBlockIter(db, cn), nil
}

func (db *DB) insert(
	ctx context.Context, model TableModel, query string, fields []*chschema.Field,
) (*result, error) {
	block := model.Block(fields)

	var res *result
	var lastErr error

	for attempt := 0; attempt <= db.conf.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := internal.Sleep(ctx, db.retryBackoff()); err != nil {
				return nil, err
			}
		}

		res, lastErr = db._insert(ctx, model, query, block)
		if !db.shouldRetry(lastErr) {
			break
		}
	}

	return res, lastErr
}

func (db *DB) _insert(
	ctx context.Context, model TableModel, query string, block *chschema.Block,
) (*result, error) {
	var res *result
	err := db.withConn(ctx, func(cn *chpool.Conn) error {
		if err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) {
			db.writeQuery(ctx, cn, wr, query)
			db.writeBlock(ctx, wr, nil)
		}); err != nil {
			return err
		}

		if err := cn.WithReader(ctx, db.conf.ReadTimeout, func(rd *chproto.Reader) error {
			_, err := db.readSampleBlock(rd)
			return err
		}); err != nil {
			return err
		}

		if err := cn.WithWriter(ctx, db.conf.WriteTimeout, func(wr *chproto.Writer) {
			db.writeBlock(ctx, wr, block)
			db.writeBlock(ctx, wr, nil)
		}); err != nil {
			return err
		}

		return cn.WithReader(ctx, db.conf.ReadTimeout, func(rd *chproto.Reader) error {
			var err error
			res, err = readPacket(cn, rd)
			if err != nil {
				return err
			}
			res.affected = block.NumRow
			return nil
		})
	})
	return res, err
}

func (db *DB) NewSelect() *SelectQuery {
	return NewSelectQuery(db)
}

func (db *DB) NewRaw(query string, args ...any) *RawQuery {
	return NewRawQuery(db, query, args...)
}

func (db *DB) NewInsert() *InsertQuery {
	return NewInsertQuery(db)
}

func (db *DB) NewCreateTable() *CreateTableQuery {
	return NewCreateTableQuery(db)
}

func (db *DB) NewDropTable() *DropTableQuery {
	return NewDropTableQuery(db)
}

func (db *DB) NewTruncateTable() *TruncateTableQuery {
	return NewTruncateTableQuery(db)
}

func (db *DB) NewCreateView() *CreateViewQuery {
	return NewCreateViewQuery(db)
}

func (db *DB) NewDropView() *DropViewQuery {
	return NewDropViewQuery(db)
}

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

func (db *DB) Formatter() chschema.Formatter {
	return db.fmter
}

func (db *DB) WithFormatter(fmter chschema.Formatter) *DB {
	clone := db.clone()
	clone.fmter = fmter
	return clone
}

func (db *DB) shouldRetry(err error) bool {
	switch err {
	case nil, context.Canceled, context.DeadlineExceeded:
		return false
	case driver.ErrBadConn:
		return true
	}

	if err, ok := err.(*Error); ok {
		// https://github.com/ClickHouse/ClickHouse/blob/master/src/Common/ErrorCodes.cpp
		const (
			timeoutExceeded            = 159
			tooSlow                    = 160
			tooManySimultaneousQueries = 202
			memoryLimitExceeded        = 241
			cannotDecompress           = 271
		)

		switch err.Code {
		case timeoutExceeded,
			tooSlow,
			tooManySimultaneousQueries,
			memoryLimitExceeded,
			cannotDecompress:
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

func (db *DB) makeQueryBytes() []byte {
	// TODO: make this configurable?
	return make([]byte, 0, 4096)
}

//------------------------------------------------------------------------------

// Rows is the result of a query. Its cursor starts before the first row of the result set.
// Use Next to advance from row to row.
type Rows struct {
	ctx    context.Context
	blocks *blockIter
	block  *chschema.Block

	rowIndex int
	hasNext  bool
	closed   bool
}

func newRows(ctx context.Context, blocks *blockIter) *Rows {
	return &Rows{
		ctx:    ctx,
		blocks: blocks,
		block:  new(chschema.Block),
	}
}

func (rs *Rows) Close() error {
	if !rs.closed {
		for rs.blocks.Next(rs.ctx, rs.block) {
		}
		rs.close()
	}
	return nil
}

func (rs *Rows) close() {
	rs.closed = true
	_ = rs.blocks.Close()
}

func (rs *Rows) ColumnTypes() ([]*sql.ColumnType, error) {
	return nil, errors.New("not implemented")
}

func (rs *Rows) Columns() ([]string, error) {
	return nil, errors.New("not implemented")
}

func (rs *Rows) Err() error {
	return rs.blocks.Err()
}

func (rs *Rows) Next() bool {
	if rs.closed {
		return false
	}

	for rs.rowIndex >= rs.block.NumRow {
		if !rs.blocks.Next(rs.ctx, rs.block) {
			rs.close()
			return false
		}
		rs.rowIndex = 0
	}

	rs.hasNext = true
	rs.rowIndex++
	return true
}

func (rs *Rows) NextResultSet() bool {
	return false
}

func (rs *Rows) Scan(dest ...any) error {
	if rs.closed {
		return rs.Err()
	}

	if !rs.hasNext {
		return errors.New("ch: Scan called without calling Next")
	}
	rs.hasNext = false

	if rs.block.NumColumn != len(dest) {
		return fmt.Errorf("ch: got %d columns, but Scan has %d values",
			rs.block.NumColumn, len(dest))
	}

	for i, col := range rs.block.Columns {
		if err := col.ConvertAssign(rs.rowIndex-1, reflect.ValueOf(dest[i]).Elem()); err != nil {
			return err
		}
	}

	return nil
}

type Row struct {
	rows *Rows
	err  error
}

func (r *Row) Err() error {
	return r.err
}

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
