package ch

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/go-clickhouse/ch/internal"
)

type InsertQuery struct {
	whereBaseQuery
}

var _ Query = (*InsertQuery)(nil)

func NewInsertQuery(db *DB) *InsertQuery {
	return &InsertQuery{
		whereBaseQuery: whereBaseQuery{
			baseQuery: baseQuery{
				db: db,
			},
		},
	}
}

func (q *InsertQuery) Model(model any) *InsertQuery {
	q.setTableModel(model)
	return q
}

//------------------------------------------------------------------------------

func (q *InsertQuery) Table(tables ...string) *InsertQuery {
	for _, table := range tables {
		q.addTable(chschema.UnsafeIdent(table))
	}
	return q
}

func (q *InsertQuery) TableExpr(query string, args ...any) *InsertQuery {
	q.addTable(chschema.SafeQuery(query, args))
	return q
}

func (q *InsertQuery) ModelTableExpr(query string, args ...any) *InsertQuery {
	q.modelTableName = chschema.SafeQuery(query, args)
	return q
}

func (q *InsertQuery) Setting(query string, args ...any) *InsertQuery {
	q.settings = append(q.settings, chschema.SafeQuery(query, args))
	return q
}

//------------------------------------------------------------------------------

func (q *InsertQuery) Column(columns ...string) *InsertQuery {
	for _, column := range columns {
		q.addColumn(chschema.UnsafeIdent(column))
	}
	return q
}

func (q *InsertQuery) ColumnExpr(query string, args ...any) *InsertQuery {
	q.addColumn(chschema.SafeQuery(query, args))
	return q
}

func (q *InsertQuery) ExcludeColumn(columns ...string) *InsertQuery {
	q.excludeColumn(columns)
	return q
}

//------------------------------------------------------------------------------

func (q *InsertQuery) Where(query string, args ...any) *InsertQuery {
	q.addWhere(chschema.SafeQueryWithSep(query, args, " AND "))
	return q
}

func (q *InsertQuery) WhereOr(query string, args ...any) *InsertQuery {
	q.addWhere(chschema.SafeQueryWithSep(query, args, " OR "))
	return q
}

func (q *InsertQuery) WhereGroup(sep string, fn func(*WhereQuery)) *InsertQuery {
	q.addWhereGroup(sep, fn)
	return q
}

//------------------------------------------------------------------------------

func (q *InsertQuery) Operation() string {
	return "INSERT"
}

var _ chschema.QueryAppender = (*InsertQuery)(nil)

func (q *InsertQuery) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}

	b = append(b, "INSERT INTO "...)
	b, err = q.appendInsertTable(fmter, b)
	if err != nil {
		return nil, err
	}

	fields, err := q.getFields()
	if err != nil {
		return nil, err
	}
	if len(fields) > 0 {
		b = append(b, " ("...)
		b = appendColumns(b, "", fields)
		b = append(b, ")"...)
	}

	b, err = q.appendValues(fmter, b)
	if err != nil {
		return nil, err
	}

	b, err = q.appendSettings(fmter, b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (q *InsertQuery) appendValues(
	fmter chschema.Formatter, b []byte,
) (_ []byte, err error) {
	if !q.hasMultiTables() {
		return append(b, " VALUES"...), nil
	}

	b = append(b, " SELECT "...)

	fields, err := q.getFields()
	if err != nil {
		return nil, err
	}
	if len(fields) > 0 {
		b = appendColumns(b, "", fields)
	} else {
		b = append(b, "*"...)
	}

	b = append(b, " FROM "...)
	b, err = q.appendOtherTables(fmter, b)
	if err != nil {
		return nil, err
	}

	if len(q.where) > 0 {
		b = append(b, " WHERE "...)

		b, err = appendWhere(fmter, b, q.where)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (q *InsertQuery) appendInsertTable(fmter chschema.Formatter, b []byte) ([]byte, error) {
	if !q.modelTableName.IsZero() {
		return q.modelTableName.AppendQuery(fmter, b)
	}

	if q.table != nil {
		return fmter.AppendQuery(b, string(q.table.CHInsertName)), nil
	}
	if len(q.tables) > 0 {
		return q.tables[0].AppendQuery(fmter, b)
	}

	return nil, errors.New("ch: query does not have a table")
}

func (q *InsertQuery) Exec(ctx context.Context) (sql.Result, error) {
	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return nil, err
	}
	query := internal.String(queryBytes)

	fields, err := q.getFields()
	if err != nil {
		return nil, err
	}

	ctx, evt := q.db.beforeQuery(ctx, q, query, nil, q.tableModel)
	res, err := q.db.insert(ctx, q.tableModel, query, fields)
	q.db.afterQuery(ctx, evt, res, err)
	return res, err
}
