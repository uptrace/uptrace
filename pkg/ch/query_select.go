package ch

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/go-clickhouse/ch/internal"
)

type SelectQuery struct {
	whereBaseQuery

	distinctOn []chschema.QueryWithArgs
	joins      []joinQuery
	group      []chschema.QueryWithArgs
	having     []chschema.QueryWithArgs
	order      []chschema.QueryWithArgs
	limit      int
	offset     int
	final      bool
}

var _ Query = (*SelectQuery)(nil)

func NewSelectQuery(db *DB) *SelectQuery {
	return &SelectQuery{
		whereBaseQuery: whereBaseQuery{
			baseQuery: baseQuery{
				db: db,
			},
		},
	}
}

func (q *SelectQuery) Operation() string {
	return "SELECT"
}

func (q *SelectQuery) Model(model any) *SelectQuery {
	q.setTableModel(model)
	return q
}

func (q *SelectQuery) Err(err error) *SelectQuery {
	q.setErr(err)
	return q
}

func (q *SelectQuery) Apply(fn func(*SelectQuery) *SelectQuery) *SelectQuery {
	return fn(q)
}

func (q *SelectQuery) WithAlias(name, query string, args ...any) *SelectQuery {
	for i := range q.with {
		with := &q.with[i]
		if with.name == name {
			with.query = chschema.SafeQuery(query, args)
			return q
		}
	}

	q.with = append(q.with, withQuery{
		name:  name,
		query: chschema.SafeQuery(query, args),
	})
	return q
}

func (q *SelectQuery) With(name string, subq chschema.QueryAppender) *SelectQuery {
	q.with = append(q.with, withQuery{
		name:  name,
		query: subq,
		cte:   true,
	})
	return q
}

func (q *SelectQuery) Distinct() *SelectQuery {
	q.distinctOn = make([]chschema.QueryWithArgs, 0)
	return q
}

func (q *SelectQuery) DistinctOn(query string, args ...any) *SelectQuery {
	q.distinctOn = append(q.distinctOn, chschema.SafeQuery(query, args))
	return q
}

//------------------------------------------------------------------------------

func (q *SelectQuery) Table(tables ...string) *SelectQuery {
	for _, table := range tables {
		q.addTable(chschema.UnsafeIdent(table))
	}
	return q
}

func (q *SelectQuery) TableExpr(query string, args ...any) *SelectQuery {
	q.addTable(chschema.SafeQuery(query, args))
	return q
}

func (q *SelectQuery) ModelTableExpr(query string, args ...any) *SelectQuery {
	q.modelTableName = chschema.SafeQuery(query, args)
	return q
}

//------------------------------------------------------------------------------

func (q *SelectQuery) Column(columns ...string) *SelectQuery {
	for _, column := range columns {
		q.addColumn(chschema.UnsafeIdent(column))
	}
	return q
}

func (q *SelectQuery) ColumnExpr(query string, args ...any) *SelectQuery {
	q.addColumn(chschema.SafeQuery(query, args))
	return q
}

func (q *SelectQuery) ExcludeColumn(columns ...string) *SelectQuery {
	q.excludeColumn(columns)
	return q
}

//------------------------------------------------------------------------------

func (q *SelectQuery) Join(join string, args ...any) *SelectQuery {
	q.joins = append(q.joins, joinQuery{
		join: chschema.SafeQuery(join, args),
	})
	return q
}

func (q *SelectQuery) JoinOn(cond string, args ...any) *SelectQuery {
	return q.joinOn(cond, args, " AND ")
}

func (q *SelectQuery) JoinOnOr(cond string, args ...any) *SelectQuery {
	return q.joinOn(cond, args, " OR ")
}

func (q *SelectQuery) joinOn(cond string, args []any, sep string) *SelectQuery {
	if len(q.joins) == 0 {
		q.err = errors.New("ch: query has no joins")
		return q
	}
	j := &q.joins[len(q.joins)-1]
	j.on = append(j.on, chschema.SafeQueryWithSep(cond, args, sep))
	return q
}

//------------------------------------------------------------------------------

func (q *SelectQuery) Where(query string, args ...any) *SelectQuery {
	q.addWhere(chschema.SafeQueryWithSep(query, args, " AND "))
	return q
}

func (q *SelectQuery) WhereOr(query string, args ...any) *SelectQuery {
	q.addWhere(chschema.SafeQueryWithSep(query, args, " OR "))
	return q
}

func (q *SelectQuery) WhereGroup(sep string, fn func(*WhereQuery)) *SelectQuery {
	q.addWhereGroup(sep, fn)
	return q
}

//------------------------------------------------------------------------------

func (q *SelectQuery) Group(columns ...string) *SelectQuery {
	for _, column := range columns {
		q.group = append(q.group, chschema.UnsafeIdent(column))
	}
	return q
}

func (q *SelectQuery) GroupExpr(group string, args ...any) *SelectQuery {
	q.group = append(q.group, chschema.SafeQuery(group, args))
	return q
}

func (q *SelectQuery) Having(having string, args ...any) *SelectQuery {
	q.having = append(q.having, chschema.SafeQuery(having, args))
	return q
}

func (q *SelectQuery) Order(orders ...string) *SelectQuery {
	for _, order := range orders {
		if order == "" {
			continue
		}

		index := strings.IndexByte(order, ' ')
		if index == -1 {
			q.order = append(q.order, chschema.UnsafeIdent(order))
			continue
		}

		field := order[:index]
		sort := order[index+1:]

		switch strings.ToUpper(sort) {
		case "ASC", "DESC", "ASC NULLS FIRST", "DESC NULLS FIRST",
			"ASC NULLS LAST", "DESC NULLS LAST":
			q.order = append(q.order, chschema.SafeQuery("? ?", []any{
				Ident(field),
				Safe(sort),
			}))
		default:
			q.order = append(q.order, chschema.UnsafeIdent(order))
		}
	}
	return q
}

// Order adds sort order to the Query.
func (q *SelectQuery) OrderExpr(order string, args ...any) *SelectQuery {
	q.order = append(q.order, chschema.SafeQuery(order, args))
	return q
}

func (q *SelectQuery) Limit(limit int) *SelectQuery {
	q.limit = limit
	return q
}

func (q *SelectQuery) Offset(offset int) *SelectQuery {
	q.offset = offset
	return q
}

func (q *SelectQuery) Final() *SelectQuery {
	q.final = true
	return q
}

//------------------------------------------------------------------------------

func (q *SelectQuery) String() string {
	b, err := q.AppendQuery(q.db.fmter, nil)
	if err != nil {
		return err.Error()
	}
	return internal.String(b)
}

func (q *SelectQuery) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	return q.appendQuery(formatterWithModel(fmter, q), b, false)
}

func (q *SelectQuery) appendQuery(
	fmter chschema.Formatter, b []byte, count bool,
) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}

	cteCount := count && (len(q.group) > 0 || len(q.distinctOn) > 0)
	if cteCount {
		b = append(b, `WITH "_count_wrapper" AS (`...)
	}

	if len(q.with) > 0 {
		b, err = q.appendWith(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	b = append(b, "SELECT "...)

	if len(q.distinctOn) > 0 {
		b = append(b, "DISTINCT ON ("...)
		for i, app := range q.distinctOn {
			if i > 0 {
				b = append(b, ", "...)
			}
			b, err = app.AppendQuery(fmter, b)
		}
		b = append(b, ") "...)
	} else if q.distinctOn != nil {
		b = append(b, "DISTINCT "...)
	}

	if count && !cteCount {
		b = append(b, "count()"...)
	} else {
		b, err = q.appendColumns(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	if q.tableModel != nil || len(q.tables) > 0 {
		b = append(b, " FROM "...)
		b, err = q.appendTablesWithAlias(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	for _, j := range q.joins {
		b = append(b, ' ')
		b, err = j.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	b, err = q.appendWhere(fmter, b)
	if err != nil {
		return nil, err
	}

	if len(q.group) > 0 {
		b = append(b, " GROUP BY "...)
		for i, f := range q.group {
			if i > 0 {
				b = append(b, ", "...)
			}
			b, err = f.AppendQuery(fmter, b)
			if err != nil {
				return nil, err
			}
		}
	}

	if len(q.having) > 0 {
		b = append(b, " HAVING "...)
		for i, f := range q.having {
			if i > 0 {
				b = append(b, " AND "...)
			}
			b = append(b, '(')
			b, err = f.AppendQuery(fmter, b)
			if err != nil {
				return nil, err
			}
			b = append(b, ')')
		}
	}

	if !count {
		if len(q.order) > 0 {
			b = append(b, " ORDER BY "...)
			for i, f := range q.order {
				if i > 0 {
					b = append(b, ", "...)
				}
				b, err = f.AppendQuery(fmter, b)
				if err != nil {
					return nil, err
				}
			}
		}

		if q.limit > 0 {
			b = append(b, " LIMIT "...)
			b = strconv.AppendInt(b, int64(q.limit), 10)
		}
		if q.offset > 0 {
			b = append(b, " OFFSET "...)
			b = strconv.AppendInt(b, int64(q.offset), 10)
		}
		if q.final {
			b = append(b, " FINAL"...)
		}
	} else if cteCount {
		b = append(b, `) SELECT `...)
		b = append(b, "count()"...)
		b = append(b, ` FROM "_count_wrapper"`...)
	}

	return b, nil
}

func (q *SelectQuery) appendWith(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	b = append(b, "WITH "...)
	for i, with := range q.with {
		if i > 0 {
			b = append(b, ", "...)
		}

		if with.cte {
			b = chschema.AppendIdent(b, with.name)
			b = append(b, " AS "...)
			b = append(b, "("...)
		}

		b, err = with.query.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}

		if with.cte {
			b = append(b, ")"...)
		} else {
			b = append(b, " AS "...)
			b = chschema.AppendIdent(b, with.name)
		}
	}
	b = append(b, ' ')
	return b, nil
}

func (q *SelectQuery) appendColumns(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
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

func (q *SelectQuery) Scan(ctx context.Context, values ...any) error {
	return q.scan(ctx, false, values...)
}

func (q *SelectQuery) ScanColumns(ctx context.Context, values ...any) error {
	return q.scan(ctx, true, values...)
}

func (q *SelectQuery) scan(ctx context.Context, columnar bool, values ...any) error {
	if q.err != nil {
		return q.err
	}

	model, err := q.newModel(values...)
	if err != nil {
		return err
	}

	if columnar {
		model.(interface{ SetColumnar(bool) }).SetColumnar(true)
	}

	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return err
	}
	query := internal.String(queryBytes)

	ctx, evt := q.db.beforeQuery(ctx, q, query, nil, model)
	res, err := q.db.query(ctx, model, query)
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

func useQueryRowModel(model Model) bool {
	if v, ok := model.(interface{ UseQueryRow() bool }); ok {
		return v.UseQueryRow()
	}
	return false
}

// Count returns number of rows matching the query using count aggregate function.
func (q *SelectQuery) Count(ctx context.Context) (int, error) {
	if q.err != nil {
		return 0, q.err
	}

	queryBytes, err := q.appendQuery(q.db.fmter, nil, true)
	if err != nil {
		return 0, err
	}
	query := internal.String(queryBytes)

	var count uint
	err = q.db.QueryRowContext(ctx, query).Scan(&count)
	return int(count), err
}

// SelectAndCount runs Select and Count in two goroutines,
// waits for them to finish and returns the result. If query limit is -1
// it does not select any data and only counts the results.
func (q *SelectQuery) ScanAndCount(
	ctx context.Context, values ...any,
) (count int, firstErr error) {
	if q.err != nil {
		return 0, q.err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	if q.limit >= 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := q.Scan(ctx, values...)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		count, err = q.Count(ctx)
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg.Wait()
	return count, firstErr
}

//------------------------------------------------------------------------------

type joinQuery struct {
	join chschema.QueryWithArgs
	on   []chschema.QueryWithSep
}

func (j *joinQuery) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	b = append(b, ' ')

	b, err = j.join.AppendQuery(fmter, b)
	if err != nil {
		return nil, err
	}

	if len(j.on) > 0 {
		b = append(b, " ON "...)
		for i, on := range j.on {
			if i > 0 {
				b = append(b, on.Sep...)
			}

			b = append(b, '(')
			b, err = on.AppendQuery(fmter, b)
			if err != nil {
				return nil, err
			}
			b = append(b, ')')
		}
	}

	return b, nil
}
