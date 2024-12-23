package ch

import (
	"context"
	"database/sql"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
)

type RawQuery struct {
	baseQuery
	query string
	args  []any
}

func (db *DB) Raw(query string, args ...any) *RawQuery {
	return &RawQuery{baseQuery: baseQuery{db: db, conn: db}, query: query, args: args}
}
func (q *RawQuery) Scan(ctx context.Context, dest ...any) error { return q.scan(ctx, dest, false) }
func (q *RawQuery) ScanColumns(ctx context.Context, dest ...any) error {
	return q.scan(ctx, dest, true)
}
func (q *RawQuery) scan(ctx context.Context, dest []any, columnar bool) error {
	if q.err != nil {
		return q.err
	}
	model, err := q.newModel(dest...)
	if err != nil {
		return err
	}
	query := q.db.FormatQuery(q.query, q.args...)
	ctx, evt, err := q.db.beforeQuery(ctx, q, query, nil, model)
	if err != nil {
		return err
	}
	res, err := q.baseQuery.query(ctx, model, query)
	q.db.afterQuery(ctx, evt, res, err)
	if err != nil {
		return err
	}
	if !columnar && useQueryRowModel(model) {
		if res.affected == 0 {
			return sql.ErrNoRows
		}
	}
	return nil
}
func (q *RawQuery) AppendQuery(fmter chschema.Formatter, b []byte) ([]byte, error) {
	return fmter.AppendQuery(b, q.query, q.args...), nil
}
func (q *RawQuery) Operation() string { return "SELECT" }
