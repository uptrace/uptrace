package ch

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/go-clickhouse/ch/internal"
)

type withQuery struct {
	name  string
	query chschema.QueryAppender
	cte   bool
}

type baseQuery struct {
	db  *DB
	err error

	tableModel TableModel
	table      *chschema.Table

	with           []withQuery
	modelTableName chschema.QueryWithArgs
	tables         []chschema.QueryWithArgs
	columns        []chschema.QueryWithArgs
	settings       []chschema.QueryWithArgs

	flags internal.Flag
}

func (q *baseQuery) DB() *DB {
	return q.db
}

func (q *baseQuery) GetModel() Model {
	return q.tableModel
}

func (q *baseQuery) GetTableName() string {
	if q.table != nil {
		return q.table.Name
	}

	for _, wq := range q.with {
		if v, ok := wq.query.(Query); ok {
			if model := v.GetModel(); model != nil {
				return v.GetTableName()
			}
		}
	}

	if q.modelTableName.Query != "" {
		return q.modelTableName.Query
	}

	if len(q.tables) > 0 {
		b, _ := q.tables[0].AppendQuery(q.db.fmter, nil)
		if len(b) < 64 {
			return string(b)
		}
	}

	return ""
}

func (q *baseQuery) setConn(db *DB) {
	q.db = db
}

func (q *baseQuery) setErr(err error) {
	if q.err == nil {
		q.err = err
	}
}

func (q *baseQuery) setTableModel(model any) {
	tm, err := newTableModel(q.db, model)
	if err != nil {
		q.setErr(err)
		return
	}
	q.tableModel = tm
	q.table = tm.Table()
}

func (q *baseQuery) newModel(values ...any) (Model, error) {
	if len(values) > 0 {
		return newModel(q.db, values...)
	}
	return q.tableModel, nil
}

func (q *baseQuery) exec(
	ctx context.Context,
	iquery Query,
	query string,
) (sql.Result, error) {
	ctx, event := q.db.beforeQuery(ctx, iquery, query, nil, q.tableModel)
	res, err := q.db.query(ctx, nil, query)
	q.db.afterQuery(ctx, event, res, err)
	return res, err
}

//------------------------------------------------------------------------------

func (q *baseQuery) AppendNamedArg(fmter chschema.Formatter, b []byte, name string) ([]byte, bool) {
	return b, false
}

func appendColumns(b []byte, table Safe, fields []*chschema.Field) []byte {
	for i, f := range fields {
		if i > 0 {
			b = append(b, ", "...)
		}

		if len(table) > 0 {
			b = append(b, table...)
			b = append(b, '.')
		}
		b = append(b, f.Column...)
	}
	return b
}

func formatterWithModel(
	fmter chschema.Formatter, model chschema.NamedArgAppender,
) chschema.Formatter {
	return fmter.WithArg(model)
}

//------------------------------------------------------------------------------

func (q *baseQuery) addTable(table chschema.QueryWithArgs) {
	q.tables = append(q.tables, table)
}

func (q *baseQuery) modelHasTableName() bool {
	if !q.modelTableName.IsZero() {
		return q.modelTableName.Query != ""
	}
	return q.table != nil
}

func (q *baseQuery) hasTables() bool {
	return q.modelHasTableName() || len(q.tables) > 0
}

func (q *baseQuery) appendTables(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	return q._appendTables(fmter, b, false)
}

func (q *baseQuery) appendTablesWithAlias(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	return q._appendTables(fmter, b, true)
}

func (q *baseQuery) _appendTables(
	fmter chschema.Formatter, b []byte, withAlias bool,
) (_ []byte, err error) {
	startLen := len(b)

	if q.modelHasTableName() {
		if !q.modelTableName.IsZero() {
			b, err = q.modelTableName.AppendQuery(fmter, b)
			if err != nil {
				return nil, err
			}
		} else {
			b = fmter.AppendQuery(b, string(q.table.CHName))
			if withAlias && q.table.CHAlias != q.table.CHName {
				b = append(b, " AS "...)
				b = append(b, q.table.CHAlias...)
			}
		}
	}

	for _, table := range q.tables {
		if len(b) > startLen {
			b = append(b, ", "...)
		}
		b, err = table.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (q *baseQuery) appendFirstTable(fmter chschema.Formatter, b []byte) ([]byte, error) {
	return q._appendFirstTable(fmter, b, false)
}

func (q *baseQuery) appendFirstTableWithAlias(
	fmter chschema.Formatter, b []byte,
) ([]byte, error) {
	return q._appendFirstTable(fmter, b, true)
}

func (q *baseQuery) _appendFirstTable(
	fmter chschema.Formatter, b []byte, withAlias bool,
) ([]byte, error) {
	if !q.modelTableName.IsZero() {
		return q.modelTableName.AppendQuery(fmter, b)
	}

	if q.table != nil {
		b = fmter.AppendQuery(b, string(q.table.CHName))
		if withAlias {
			b = append(b, " AS "...)
			b = append(b, q.table.CHAlias...)
		}
		return b, nil
	}

	if len(q.tables) > 0 {
		return q.tables[0].AppendQuery(fmter, b)
	}

	return nil, errors.New("ch: query does not have a table")
}

func (q *baseQuery) hasMultiTables() bool {
	if q.modelHasTableName() {
		return len(q.tables) >= 1
	}
	return len(q.tables) >= 2
}

func (q *baseQuery) appendOtherTables(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	tables := q.tables
	if !q.modelHasTableName() {
		tables = tables[1:]
	}
	for i, table := range tables {
		if i > 0 {
			b = append(b, ", "...)
		}
		b, err = table.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}

//------------------------------------------------------------------------------

func (q *baseQuery) addColumn(column chschema.QueryWithArgs) {
	q.columns = append(q.columns, column)
}

func (q *baseQuery) excludeColumn(columns []string) {
	if q.columns == nil {
		for _, f := range q.table.Fields {
			q.columns = append(q.columns, chschema.UnsafeIdent(f.CHName))
		}
	}

	if len(columns) == 1 && columns[0] == "*" {
		q.columns = make([]chschema.QueryWithArgs, 0)
		return
	}

	for _, column := range columns {
		if !q._excludeColumn(column) {
			q.setErr(fmt.Errorf("ch: can't find column=%q", column))
			return
		}
	}
}

func (q *baseQuery) _excludeColumn(column string) bool {
	for i, col := range q.columns {
		if col.Args == nil && col.Query == column {
			q.columns = append(q.columns[:i], q.columns[i+1:]...)
			return true
		}
	}
	return false
}

func (q *baseQuery) getFields() ([]*chschema.Field, error) {
	if len(q.columns) == 0 {
		if q.table == nil {
			return nil, nil
		}
		return q.table.Fields, nil
	}

	fields := make([]*chschema.Field, 0, len(q.columns))
	for _, col := range q.columns {
		if col.Args != nil {
			continue
		}

		field, err := q.table.Field(col.Query)
		if err != nil {
			return nil, err
		}

		fields = append(fields, field)
	}
	return fields, nil
}

func (q *baseQuery) appendSettings(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	if len(q.settings) > 0 {
		b = append(b, " SETTINGS "...)
		for i, opt := range q.settings {
			if i > 0 {
				b = append(b, ", "...)
			}
			b, err = opt.AppendQuery(fmter, b)
			if err != nil {
				return nil, err
			}
		}
	}
	return b, nil
}

//------------------------------------------------------------------------------

type WhereQuery struct {
	where []chschema.QueryWithSep
}

func (q *WhereQuery) addWhere(where chschema.QueryWithSep) {
	q.where = append(q.where, where)
}

func (q *WhereQuery) WhereGroup(sep string, fn func(*WhereQuery)) {
	q.addWhereGroup(sep, fn)
}

func (q *WhereQuery) addWhereGroup(sep string, fn func(*WhereQuery)) {
	q2 := new(WhereQuery)
	fn(q2)

	if len(q2.where) > 0 {
		q2.where[0].Sep = ""

		q.addWhere(chschema.SafeQueryWithSep("", nil, sep+"("))
		q.where = append(q.where, q2.where...)
		q.addWhere(chschema.SafeQueryWithSep("", nil, ")"))
	}
}

//------------------------------------------------------------------------------

type whereBaseQuery struct {
	baseQuery
	WhereQuery
}

func (q *whereBaseQuery) mustAppendWhere(fmter chschema.Formatter, b []byte) ([]byte, error) {
	if len(q.where) == 0 {
		err := errors.New("ch: Update and Delete queries require at least one Where")
		return nil, err
	}
	return q.appendWhere(fmter, b)
}

func (q *whereBaseQuery) appendWhere(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	if len(q.where) == 0 {
		return b, nil
	}

	b = append(b, " WHERE "...)

	b, err = appendWhere(fmter, b, q.where)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func appendWhere(
	fmter chschema.Formatter, b []byte, where []chschema.QueryWithSep,
) (_ []byte, err error) {
	for i, where := range where {
		if i > 0 || where.Sep == "(" {
			b = append(b, where.Sep...)
		}

		if where.Query == "" && where.Args == nil {
			continue
		}

		b = append(b, '(')
		b, err = where.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
		b = append(b, ')')
	}
	return b, nil
}
