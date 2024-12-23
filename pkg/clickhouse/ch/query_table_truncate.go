package ch

import (
	"context"
	"database/sql"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"github.com/uptrace/pkg/unsafeconv"
)

type TruncateTableQuery struct {
	baseQuery
	ifExists bool
}

var _ Query = (*TruncateTableQuery)(nil)

func NewTruncateTableQuery(db *DB) *TruncateTableQuery {
	q := &TruncateTableQuery{baseQuery: baseQuery{db: db, conn: db}}
	return q
}
func (q *TruncateTableQuery) Model(model any) *TruncateTableQuery {
	q.setTableModel(model)
	return q
}
func (q *TruncateTableQuery) Table(tables ...string) *TruncateTableQuery {
	for _, table := range tables {
		q.addTable(chschema.UnsafeName(table))
	}
	return q
}
func (q *TruncateTableQuery) TableExpr(query string, args ...any) *TruncateTableQuery {
	q.addTable(chschema.SafeQuery(query, args))
	return q
}
func (q *TruncateTableQuery) IfExists() *TruncateTableQuery {
	q.ifExists = true
	return q
}
func (q *TruncateTableQuery) Operation() string { return "TRUNCATE TABLE" }
func (q *TruncateTableQuery) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}
	b = append(b, "TRUNCATE TABLE "...)
	if q.ifExists {
		b = append(b, "IF EXISTS "...)
	}
	b, err = q.appendTables(fmter, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
func (q *TruncateTableQuery) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return nil, err
	}
	query := unsafeconv.String(queryBytes)
	return q.conn.ExecContext(ctx, query)
}
