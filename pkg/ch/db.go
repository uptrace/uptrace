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
	cfg  *Config
	pool *chpool.ConnPool

	queryHooks []QueryHook

	fmter chschema.Formatter
	flags internal.Flag
	stats DBStats
}

func Connect(opts ...Option) *DB {
	db := &DB{
		cfg: defaultConfig(),
	}

	for _, opt := range opts {
		opt(db)
	}
	db.pool = newConnPool(db.cfg)

	return db
}

func newConnPool(cfg *Config) *chpool.ConnPool {
	poolcfg := cfg.Config
	poolcfg.Dialer = func(ctx context.Context) (net.Conn, error) {
		if cfg.TLSConfig != nil {
			return tls.DialWithDialer(
				cfg.netDialer(),
				cfg.Network,
				cfg.Addr,
				cfg.TLSConfig,
			)
		}
		return cfg.netDialer().DialContext(ctx, cfg.Network, cfg.Addr)
	}
	return chpool.New(&poolcfg)
}

// Close closes the database client, releasing any open resources.
//
// It is rare to Close a DB, as the DB handle is meant to be
// long-lived and shared between many goroutines.
func (db *DB) Close() error {
	return db.pool.Close()
}

func (db *DB) String() string {
	return fmt.Sprintf("DB<addr: %s>", db.cfg.Addr)
}

func (db *DB) Config() *Config {
	return db.cfg
}

func (db *DB) WithTimeout(d time.Duration) *DB {
	newcfg := *db.cfg
	newcfg.ReadTimeout = d
	newcfg.WriteTimeout = d

	clone := db.clone()
	clone.cfg = &newcfg

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

	var done chan struct{}

	if ctxDone := ctx.Done(); ctxDone != nil {
		done = make(chan struct{})
		go func() {
			select {
			case <-done:
				// fn has finished, skip cancel
			case <-ctxDone:
				db.cancelConn(ctx, cn)
				// Signal end of conn use.
				done <- struct{}{}
			}
		}()
	}

	defer func() {
		if done != nil {
			select {
			case <-done: // wait for cancel to finish request
			case done <- struct{}{}: // signal fn finish, skip cancel goroutine
			}
		}
		db.releaseConn(cn, err)
	}()

	// err is used in releaseConn above
	err = fn(cn)

	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		db.cancelConn(ctx, cn)
	}

	return err
}

func (db *DB) cancelConn(ctx context.Context, cn *chpool.Conn) {
	if err := cn.WithWriter(ctx, db.cfg.WriteTimeout, func(wr *chproto.Writer) {
		writeCancel(wr)
	}); err != nil {
		internal.Logger.Printf("writeCancel failed: %s", err)
	}

	_ = cn.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	return db.withConn(ctx, func(cn *chpool.Conn) error {
		if err := cn.WithWriter(ctx, db.cfg.WriteTimeout, func(wr *chproto.Writer) {
			writePing(wr)
		}); err != nil {
			return err
		}
		return cn.WithReader(ctx, db.cfg.ReadTimeout, func(rd *chproto.Reader) error {
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
	res, err := db.query(ctx, nil, query)
	db.afterQuery(ctx, evt, res, err)
	return res, err
}

func (db *DB) Query(query string, args ...any) (*Rows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

func (db *DB) QueryContext(
	ctx context.Context, query string, args ...any,
) (*Rows, error) {
	rows := newRows()
	query = db.FormatQuery(query, args...)

	ctx, evt := db.beforeQuery(ctx, nil, query, args, nil)
	res, err := db.query(ctx, rows, query)
	db.afterQuery(ctx, evt, res, err)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (db *DB) QueryRow(query string, args ...any) *Row {
	return db.QueryRowContext(context.Background(), query, args...)
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *Row {
	rows, err := db.QueryContext(ctx, query, args...)
	return &Row{rows: rows, err: err}
}

func (db *DB) query(ctx context.Context, model Model, query string) (*result, error) {
	var res *result
	var lastErr error
	for attempt := 0; attempt <= db.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			lastErr = internal.Sleep(ctx, db.retryBackoff(attempt-1))
			if lastErr != nil {
				break
			}
		}

		res, lastErr = db._query(ctx, model, query)
		if !db.shouldRetry(lastErr) {
			break
		}
	}

	if lastErr == nil {
		if model, ok := model.(AfterScanRowHook); ok {
			if err := model.AfterScanRow(ctx); err != nil {
				lastErr = err
			}
		}
	}

	return res, lastErr
}

func (db *DB) _query(ctx context.Context, model Model, query string) (*result, error) {
	var res *result
	err := db.withConn(ctx, func(cn *chpool.Conn) error {
		if err := cn.WithWriter(ctx, db.cfg.WriteTimeout, func(wr *chproto.Writer) {
			db.writeQuery(wr, query)
			writeBlock(ctx, wr, nil)
		}); err != nil {
			return err
		}
		return cn.WithReader(ctx, db.cfg.ReadTimeout, func(rd *chproto.Reader) error {
			var err error
			res, err = readDataBlocks(rd, model)
			return err
		})
	})
	return res, err
}

func (db *DB) insert(
	ctx context.Context, model TableModel, query string, fields []*chschema.Field,
) (*result, error) {
	block := model.Block(fields)

	var res *result
	var lastErr error

	for attempt := 0; attempt <= db.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			lastErr = internal.Sleep(ctx, db.retryBackoff(attempt-1))
			if lastErr != nil {
				break
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
		if err := cn.WithWriter(ctx, db.cfg.WriteTimeout, func(wr *chproto.Writer) {
			db.writeQuery(wr, query)
			writeBlock(ctx, wr, nil)
		}); err != nil {
			return err
		}

		if err := cn.WithReader(ctx, db.cfg.ReadTimeout, func(rd *chproto.Reader) error {
			_, err := readSampleBlock(rd)
			return err
		}); err != nil {
			return err
		}

		if err := cn.WithWriter(ctx, db.cfg.WriteTimeout, func(wr *chproto.Writer) {
			writeBlock(ctx, wr, block)
			writeBlock(ctx, wr, nil)
		}); err != nil {
			return err
		}

		return cn.WithReader(ctx, db.cfg.ReadTimeout, func(rd *chproto.Reader) error {
			var err error
			res, err = readPacket(rd)
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
	case driver.ErrBadConn:
		return true
	case nil, context.Canceled, context.DeadlineExceeded:
		return false
	}

	if err, ok := err.(*Error); ok {
		// https://github.com/ClickHouse/ClickHouse/blob/master/src/Common/ErrorCodes.cpp
		const (
			timeoutExceeded            = 159
			tooManySimultaneousQueries = 202
			memoryLimitExceeded        = 241
		)

		switch err.Code {
		case timeoutExceeded, tooManySimultaneousQueries, memoryLimitExceeded:
			return true
		}
	}

	return false
}

func (db *DB) retryBackoff(attempt int) time.Duration {
	return internal.RetryBackoff(
		attempt, db.cfg.MinRetryBackoff, db.cfg.MaxRetryBackoff)
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
	blocks []*chschema.Block

	block      *chschema.Block
	blockIndex int
	rowIndex   int
}

func newRows() *Rows {
	return new(Rows)
}

func (rs *Rows) Close() error {
	return nil
}

func (rs *Rows) ColumnTypes() ([]*sql.ColumnType, error) {
	return nil, errors.New("not implemented")
}

func (rs *Rows) Columns() ([]string, error) {
	return nil, errors.New("not implemented")
}

func (rs *Rows) Err() error {
	return nil
}

func (rs *Rows) Next() bool {
	if rs.block != nil && rs.rowIndex < rs.block.NumRow {
		rs.rowIndex++
		return true
	}

	for rs.blockIndex < len(rs.blocks) {
		rs.block = rs.blocks[rs.blockIndex]
		rs.blockIndex++
		if rs.block.NumRow > 0 {
			rs.rowIndex = 1
			return true
		}
	}

	return false
}

func (rs *Rows) NextResultSet() bool {
	return false
}

func (rs *Rows) Scan(dest ...any) error {
	if rs.block == nil {
		return errors.New("ch: Scan called without calling Next")
	}

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

func (rs *Rows) ScanBlock(block *chschema.Block) error {
	rs.blocks = append(rs.blocks, block)
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
