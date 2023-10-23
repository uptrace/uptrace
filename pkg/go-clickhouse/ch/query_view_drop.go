package ch

import (
	"context"
	"database/sql"

	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/go-clickhouse/ch/internal"
)

type DropViewQuery struct {
	baseQuery

	ifExists  bool
	view      chschema.QueryWithArgs
	onCluster chschema.QueryWithArgs
}

var _ Query = (*DropViewQuery)(nil)

func NewDropViewQuery(db *DB) *DropViewQuery {
	q := &DropViewQuery{
		baseQuery: baseQuery{
			db: db,
		},
	}
	return q
}

func (q *DropViewQuery) Model(model any) *DropViewQuery {
	q.setTableModel(model)
	return q
}

func (q *DropViewQuery) Apply(fn func(*DropViewQuery) *DropViewQuery) *DropViewQuery {
	return fn(q)
}

//------------------------------------------------------------------------------

func (q *DropViewQuery) IfExists() *DropViewQuery {
	q.ifExists = true
	return q
}

func (q *DropViewQuery) View(view string) *DropViewQuery {
	q.view = chschema.UnsafeName(view)
	return q
}

func (q *DropViewQuery) ViewExpr(query string, args ...any) *DropViewQuery {
	q.view = chschema.SafeQuery(query, args)
	return q
}

func (q *DropViewQuery) OnCluster(cluster string) *DropViewQuery {
	q.onCluster = chschema.UnsafeName(cluster)
	return q
}

func (q *DropViewQuery) OnClusterExpr(query string, args ...any) *DropViewQuery {
	q.onCluster = chschema.SafeQuery(query, args)
	return q
}

//------------------------------------------------------------------------------

func (q *DropViewQuery) Operation() string {
	return "DROP TABLE"
}

func (q *DropViewQuery) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}

	b = append(b, "DROP VIEW "...)
	if q.ifExists {
		b = append(b, "IF EXISTS "...)
	}

	b, err = q.view.AppendQuery(fmter, b)
	if err != nil {
		return nil, err
	}

	if !q.onCluster.IsEmpty() {
		b = append(b, " ON CLUSTER "...)
		b, err = q.onCluster.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

//------------------------------------------------------------------------------

func (q *DropViewQuery) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return nil, err
	}
	query := internal.String(queryBytes)

	return q.exec(ctx, q, query)
}
