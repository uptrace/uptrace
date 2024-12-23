package ch

import (
	"context"
	"database/sql"
	"errors"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"github.com/uptrace/pkg/unsafeconv"
)

type InsertQuery struct {
	baseQuery
	selq  *SelectQuery
	where whereQuery
	block *Block
}

var _ Query = (*InsertQuery)(nil)

func NewInsertQuery(db *DB) *InsertQuery            { return &InsertQuery{baseQuery: baseQuery{db: db, conn: db}} }
func (q *InsertQuery) Model(model any) *InsertQuery { q.setTableModel(model); return q }
func (q *InsertQuery) Table(tables ...string) *InsertQuery {
	for _, table := range tables {
		q.addTable(chschema.UnsafeName(table))
	}
	return q
}
func (q *InsertQuery) TableExpr(query string, args ...any) *InsertQuery {
	q.addTable(chschema.SafeQuery(query, args))
	return q
}
func (q *InsertQuery) ModelTable(table string) *InsertQuery {
	q.modelTableName = chschema.UnsafeName(table)
	return q
}
func (q *InsertQuery) ModelTableExpr(query string, args ...any) *InsertQuery {
	q.modelTableName = chschema.SafeQuery(query, args)
	return q
}
func (q *InsertQuery) DistTable(table string) *InsertQuery { q.addTable(q.distTable(table)); return q }
func (q *InsertQuery) ModelDistTable(table string) *InsertQuery {
	q.modelTableName = q.distTable(table)
	return q
}
func (q *InsertQuery) From(selq *SelectQuery) *InsertQuery { q.selq = selq; return q }
func (q *InsertQuery) Block(block *Block) *InsertQuery     { q.block = block; return q }
func (q *InsertQuery) Setting(query string, args ...any) *InsertQuery {
	q.settings = append(q.settings, chschema.SafeQuery(query, args))
	return q
}
func (q *InsertQuery) Column(columns ...string) *InsertQuery {
	for _, column := range columns {
		q.addColumn(chschema.UnsafeName(column))
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
func (q *InsertQuery) Where(query string, args ...any) *InsertQuery {
	q.where.addFilter(chschema.SafeQueryWithSep(query, args, " AND "))
	return q
}
func (q *InsertQuery) WhereOr(query string, args ...any) *InsertQuery {
	q.where.addFilter(chschema.SafeQueryWithSep(query, args, " OR "))
	return q
}
func (q *InsertQuery) Operation() string { return "INSERT" }
func (q *InsertQuery) String() string {
	buf, err := q.AppendQuery(q.db.Formatter(), nil)
	if err != nil {
		return err.Error()
	}
	return unsafeconv.String(buf)
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
	if q.selq != nil {
		return q.appendInsertFrom(fmter, b, q.selq)
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
	b, err = q.appendSettings(fmter, b)
	if err != nil {
		return nil, err
	}
	b, err = q.appendValues(fmter, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
func (q *InsertQuery) appendInsertFrom(fmter chschema.Formatter, buf []byte, selq *SelectQuery) (_ []byte, err error) {
	buf = append(buf, " ("...)
	buf, err = q.appendColumns(fmter, buf)
	if err != nil {
		return nil, err
	}
	buf = append(buf, ") "...)
	buf, err = selq.AppendQuery(fmter, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
func (q *InsertQuery) appendValues(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	if !q.hasMultiTables() {
		return append(b, " FORMAT Native"...), nil
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
	if len(q.where.filters) > 0 {
		b = append(b, " WHERE "...)
		b, err = appendWhere(fmter, b, q.where.filters)
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}
func (q *InsertQuery) appendInsertTable(fmter chschema.Formatter, buf []byte) ([]byte, error) {
	if !q.modelTableName.IsZero() {
		return q.modelTableName.AppendQuery(fmter, buf)
	}
	if q.table != nil {
		return chschema.AppendName(buf, q.db.DistTable(q.table.CHInsertName)), nil
	}
	if len(q.tables) > 0 {
		return q.tables[0].AppendQuery(fmter, buf)
	}
	return nil, errors.New("ch: query does not have a table")
}
func (q *InsertQuery) Exec(ctx context.Context) (sql.Result, error) {
	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return nil, err
	}
	query := unsafeconv.String(queryBytes)
	ctx, evt, err := q.db.beforeQuery(ctx, q, query, nil, q.tableModel)
	if err != nil {
		return nil, err
	}
	var res *Result
	var retErr error
	if q.tableModel != nil {
		block := q.block
		if block == nil {
			block = NewBlock()
		}
		fields, err := q.getFields()
		if err != nil {
			return nil, err
		}
		q.tableModel.WriteToBlock(block, fields)
		res, retErr = q.db.insert(ctx, q.tableModel, query, block)
	} else {
		res, retErr = q.db.exec(ctx, query)
	}
	q.db.afterQuery(ctx, evt, res, retErr)
	return res, retErr
}
