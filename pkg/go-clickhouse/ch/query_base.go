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

func (q *baseQuery) clone() baseQuery {
	clone := *q
	clone.with = lazyClone(clone.with)
	clone.tables = lazyClone(clone.tables)
	clone.columns = lazyClone(clone.columns)
	clone.settings = lazyClone(clone.settings)
	return clone
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

func (q *baseQuery) query(ctx context.Context, model Model, query string) (*result, error) {
	blocks, err := q.db.query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer blocks.Close()

	res := &result{
		model: model,
	}
	block := new(chschema.Block)
	if model, ok := model.(TableModel); ok {
		block.Table = model.Table()
	}

	for blocks.Next(block) {
		if err := model.ScanBlock(block); err != nil {
			return nil, err
		}
		res.affected += block.NumRow
	}
	if err := blocks.Err(); err != nil {
		return nil, err
	}

	if model, ok := model.(AfterScanRowHook); ok {
		if err := model.AfterScanRow(ctx); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (q *baseQuery) exec(
	ctx context.Context,
	iquery Query,
	query string,
) (sql.Result, error) {
	ctx, event := q.db.beforeQuery(ctx, iquery, query, nil, q.tableModel)
	res, err := q.db.exec(ctx, query)
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
			q.columns = append(q.columns, chschema.UnsafeName(f.CHName))
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

func (q *baseQuery) appendColumns(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	switch {
	case q.columns != nil:
		for i, f := range q.columns {
			if i > 0 {
				b = append(b, ", "...)
			}
			b, err = f.AppendQuery(fmter, b)
			if err != nil {
				return nil, err
			}
		}
	case q.table != nil:
		b = appendTableColumns(b, q.table.CHAlias, q.table.Fields)
	default:
		b = append(b, '*')
	}
	return b, nil
}

func appendTableColumns(b []byte, table chschema.Safe, fields []*chschema.Field) []byte {
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

//------------------------------------------------------------------------------

type whereQuery struct {
	filters []chschema.QueryWithSep
}

func (q *whereQuery) clone() whereQuery {
	clone := *q
	clone.filters = lazyClone(clone.filters)
	return clone
}

func (q *whereQuery) addFilter(filter chschema.QueryWithSep) {
	q.filters = append(q.filters, filter)
}

func (q *whereQuery) addGroup(sep string, filters []chschema.QueryWithSep) {
	if len(filters) == 0 {
		return
	}

	q.addFilter(chschema.SafeQueryWithSep("", nil, sep))
	q.addFilter(chschema.SafeQueryWithSep("", nil, "("))

	filters[0].Sep = ""
	q.filters = append(q.filters, filters...)

	q.addFilter(chschema.SafeQueryWithSep("", nil, ")"))
}

func appendWhere(
	fmter chschema.Formatter, b []byte, where []chschema.QueryWithSep,
) (_ []byte, err error) {
	for i, where := range where {
		if i > 0 || where.Sep == "(" {
			b = append(b, where.Sep...)
		}

		if where.Query == "" {
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

func lazyClone[S ~[]E, E any](s S) S {
	// Preserve nil in case it matters.
	if s == nil {
		return nil
	}
	return s[:len(s):len(s)]
}
