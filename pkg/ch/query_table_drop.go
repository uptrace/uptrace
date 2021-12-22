package ch

import (
	"context"
	"database/sql"

	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/go-clickhouse/ch/internal"
)

type DropTableQuery struct {
	baseQuery

	ifExists bool
}

var _ Query = (*DropTableQuery)(nil)

func NewDropTableQuery(db *DB) *DropTableQuery {
	q := &DropTableQuery{
		baseQuery: baseQuery{
			db: db,
		},
	}
	return q
}

func (q *DropTableQuery) Model(model any) *DropTableQuery {
	q.setTableModel(model)
	return q
}

//------------------------------------------------------------------------------

func (q *DropTableQuery) Table(tables ...string) *DropTableQuery {
	for _, table := range tables {
		q.addTable(chschema.UnsafeIdent(table))
	}
	return q
}

func (q *DropTableQuery) TableExpr(query string, args ...any) *DropTableQuery {
	q.addTable(chschema.SafeQuery(query, args))
	return q
}

func (q *DropTableQuery) ModelTableExpr(query string, args ...any) *DropTableQuery {
	q.modelTableName = chschema.SafeQuery(query, args)
	return q
}

//------------------------------------------------------------------------------

func (q *DropTableQuery) IfExists() *DropTableQuery {
	q.ifExists = true
	return q
}

//------------------------------------------------------------------------------

func (q *DropTableQuery) Operation() string {
	return "DROP TABLE"
}

func (q *DropTableQuery) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}

	b = append(b, "DROP TABLE "...)
	if q.ifExists {
		b = append(b, "IF EXISTS "...)
	}

	b, err = q.appendTables(fmter, b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

//------------------------------------------------------------------------------

func (q *DropTableQuery) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return nil, err
	}
	query := internal.String(queryBytes)

	return q.exec(ctx, q, query)
}
