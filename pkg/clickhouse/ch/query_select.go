package ch

import (
	"context"
	"database/sql"
	"errors"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"github.com/uptrace/pkg/unsafeconv"
	"strconv"
	"strings"
	"sync"
)

type SelectQuery struct {
	baseQuery
	prewhere   whereQuery
	where      whereQuery
	sample     chschema.QueryWithArgs
	distinctOn []chschema.QueryWithArgs
	joins      []joinQuery
	group      []chschema.QueryWithArgs
	having     []chschema.QueryWithArgs
	order      []chschema.QueryWithArgs
	limit      int
	offset     int
	final      bool
	union      []union
}
type union struct {
	expr  string
	query *SelectQuery
}

var _ Query = (*SelectQuery)(nil)

func NewSelectQuery(db *DB) *SelectQuery { return &SelectQuery{baseQuery: baseQuery{db: db, conn: db}} }
func (q *SelectQuery) Clone() *SelectQuery {
	clone := *q
	clone.baseQuery = clone.baseQuery.clone()
	clone.prewhere = clone.prewhere.clone()
	clone.where = clone.where.clone()
	clone.distinctOn = cloneLazy(clone.distinctOn)
	clone.joins = cloneLazy(clone.joins)
	clone.group = cloneLazy(clone.group)
	clone.having = cloneLazy(clone.having)
	clone.order = cloneLazy(clone.order)
	return &clone
}
func (q *SelectQuery) Operation() string { return "SELECT" }
func (q *SelectQuery) Conn(conn IConn) *SelectQuery {
	q.setConn(conn)
	return q
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
	if fn != nil {
		return fn(q)
	}
	return q
}
func (q *SelectQuery) WithAlias(name, query string, args ...any) *SelectQuery {
	for i := range q.with {
		with := &q.with[i]
		if with.name == name {
			with.query = chschema.SafeQuery(query, args)
			return q
		}
	}
	q.with = append(q.with, withQuery{name: name, query: chschema.SafeQuery(query, args)})
	return q
}
func (q *SelectQuery) With(name string, subq chschema.QueryAppender) *SelectQuery {
	q.with = append(q.with, withQuery{name: name, query: subq, cte: true})
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
func (q *SelectQuery) Table(tables ...string) *SelectQuery {
	for _, table := range tables {
		q.addTable(chschema.UnsafeName(table))
	}
	return q
}
func (q *SelectQuery) TableExpr(query string, args ...any) *SelectQuery {
	q.addTable(chschema.SafeQuery(query, args))
	return q
}
func (q *SelectQuery) ModelTable(table string) *SelectQuery {
	q.modelTableName = chschema.UnsafeName(table)
	return q
}
func (q *SelectQuery) ModelTableExpr(query string, args ...any) *SelectQuery {
	q.modelTableName = chschema.SafeQuery(query, args)
	return q
}
func (q *SelectQuery) DistTable(table, alias string) *SelectQuery {
	q.addTable(q.allDistTable(table, alias))
	return q
}
func (q *SelectQuery) ModelDistTable(table, alias string) *SelectQuery {
	q.modelTableName = q.allDistTable(table, alias)
	return q
}
func (q *SelectQuery) Sample(query string, args ...any) *SelectQuery {
	q.sample = chschema.SafeQuery(query, args)
	return q
}
func (q *SelectQuery) Column(columns ...string) *SelectQuery {
	for _, column := range columns {
		q.addColumn(chschema.UnsafeName(column))
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
func (q *SelectQuery) UnionAll(other *SelectQuery) *SelectQuery {
	return q.addUnion(" UNION ALL ", other)
}
func (q *SelectQuery) addUnion(expr string, other *SelectQuery) *SelectQuery {
	q.union = append(q.union, union{expr: expr, query: other})
	return q
}
func (q *SelectQuery) Join(join string, args ...any) *SelectQuery {
	q.joins = append(q.joins, joinQuery{join: chschema.SafeQuery(join, args)})
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
func (q *SelectQuery) Prewhere(query string, args ...any) *SelectQuery {
	q.prewhere.addFilter(chschema.SafeQueryWithSep(query, args, " AND "))
	return q
}
func (q *SelectQuery) PrewhereOr(query string, args ...any) *SelectQuery {
	q.prewhere.addFilter(chschema.SafeQueryWithSep(query, args, " OR "))
	return q
}
func (q *SelectQuery) PrewhereGroup(sep string, fn func(*SelectQuery) *SelectQuery) *SelectQuery {
	saved := q.prewhere.filters
	q.prewhere.filters = nil
	q = fn(q)
	filters := q.prewhere.filters
	q.prewhere.filters = saved
	q.prewhere.addGroup(sep, filters)
	return q
}
func (q *SelectQuery) Where(query string, args ...any) *SelectQuery {
	q.where.addFilter(chschema.SafeQueryWithSep(query, args, " AND "))
	return q
}
func (q *SelectQuery) WhereOr(query string, args ...any) *SelectQuery {
	q.where.addFilter(chschema.SafeQueryWithSep(query, args, " OR "))
	return q
}
func (q *SelectQuery) WhereGroup(sep string, fn func(*SelectQuery) *SelectQuery) *SelectQuery {
	saved := q.where.filters
	q.where.filters = nil
	q = fn(q)
	filters := q.where.filters
	q.where.filters = saved
	q.where.addGroup(sep, filters)
	return q
}
func (q *SelectQuery) Group(columns ...string) *SelectQuery {
	for _, column := range columns {
		q.group = append(q.group, chschema.UnsafeName(column))
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
			q.order = append(q.order, chschema.UnsafeName(order))
			continue
		}
		field := order[:index]
		sort := order[index+1:]
		switch strings.ToUpper(sort) {
		case "ASC", "DESC", "ASC NULLS FIRST", "DESC NULLS FIRST", "ASC NULLS LAST", "DESC NULLS LAST":
			q.order = append(q.order, chschema.SafeQuery("? ?", []any{Name(field), Safe(sort)}))
		default:
			q.order = append(q.order, chschema.UnsafeName(order))
		}
	}
	return q
}
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
func (q *SelectQuery) Setting(query string, args ...any) *SelectQuery {
	q.settings = append(q.settings, chschema.SafeQuery(query, args))
	return q
}
func (q *SelectQuery) String() string {
	b, err := q.AppendQuery(q.db.Formatter(), nil)
	if err != nil {
		return err.Error()
	}
	return unsafeconv.String(b)
}
func (q *SelectQuery) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	return q.appendQuery(formatterWithModel(fmter, q), b, false)
}
func (q *SelectQuery) appendQuery(fmter chschema.Formatter, b []byte, count bool) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}
	cteCount := count && (len(q.group) > 0 || len(q.distinctOn) > 0)
	if cteCount {
		b = append(b, `WITH "_count_wrapper" AS (`...)
	}
	if len(q.union) > 0 {
		b = append(b, '(')
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
	if q.hasTables() {
		b = append(b, " FROM "...)
		b, err = q.appendTablesWithAlias(fmter, b)
		if err != nil {
			return nil, err
		}
	}
	if !q.sample.IsEmpty() {
		b = append(b, " SAMPLE "...)
		b, err = q.sample.AppendQuery(fmter, b)
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
	if len(q.prewhere.filters) > 0 {
		b = append(b, " PREWHERE "...)
		b, err = appendWhere(fmter, b, q.prewhere.filters)
		if err != nil {
			return nil, err
		}
	}
	if len(q.where.filters) > 0 {
		b = append(b, " WHERE "...)
		b, err = appendWhere(fmter, b, q.where.filters)
		if err != nil {
			return nil, err
		}
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
	}
	if len(q.union) > 0 {
		b = append(b, ')')
		for _, u := range q.union {
			b = append(b, u.expr...)
			b = append(b, '(')
			b, err = u.query.AppendQuery(fmter, b)
			if err != nil {
				return nil, err
			}
			b = append(b, ')')
		}
	}
	if cteCount {
		b = append(b, `) SELECT `...)
		b = append(b, "count()"...)
		b = append(b, ` FROM "_count_wrapper"`...)
	}
	b, err = q.appendSettings(fmter, b)
	if err != nil {
		return nil, err
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
			b = chschema.AppendName(b, with.name)
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
			b = chschema.AppendName(b, with.name)
		}
	}
	b = append(b, ' ')
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
func appendTableColumns(b []byte, table string, fields []*chschema.Field) []byte {
	for i, f := range fields {
		if i > 0 {
			b = append(b, ", "...)
		}
		if len(table) > 0 {
			b = chschema.AppendName(b, table)
			b = append(b, '.')
		}
		b = append(b, f.Column...)
	}
	return b
}
func (q *SelectQuery) Query(ctx context.Context) (*Rows, error) {
	if q.err != nil {
		return nil, q.err
	}
	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return nil, err
	}
	query := unsafeconv.String(queryBytes)
	return q.db.QueryContext(ctx, query)
}
func (q *SelectQuery) QueryBlocks(ctx context.Context) (*BlockIter, error) {
	if q.err != nil {
		return nil, q.err
	}
	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return nil, err
	}
	query := unsafeconv.String(queryBytes)
	return q.db.QueryBlocks(ctx, query)
}
func (q *SelectQuery) ScanResult(ctx context.Context, values ...any) (*Result, error) {
	return q.scan(ctx, false, values...)
}
func (q *SelectQuery) Scan(ctx context.Context, values ...any) error {
	_, err := q.scan(ctx, false, values...)
	return err
}
func (q *SelectQuery) ScanColumns(ctx context.Context, values ...any) error {
	_, err := q.scan(ctx, true, values...)
	return err
}
func (q *SelectQuery) scan(ctx context.Context, columnar bool, values ...any) (*Result, error) {
	if q.err != nil {
		return nil, q.err
	}
	model, err := q.newModel(values...)
	if err != nil {
		return nil, err
	}
	if columnar {
		model.(interface{ SetColumnar(bool) }).SetColumnar(true)
	}
	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return nil, err
	}
	query := unsafeconv.String(queryBytes)
	ctx, evt, err := q.db.beforeQuery(ctx, q, query, nil, model)
	if err != nil {
		return nil, err
	}
	res, err := q.query(ctx, model, query)
	q.db.afterQuery(ctx, evt, res, err)
	if err != nil {
		return nil, err
	}
	if !columnar && useQueryRowModel(model) {
		if res.affected == 0 {
			return nil, sql.ErrNoRows
		}
	}
	return res, nil
}
func useQueryRowModel(model Model) bool {
	if v, ok := model.(interface{ UseQueryRow() bool }); ok {
		return v.UseQueryRow()
	}
	return false
}
func (q *SelectQuery) EstimateRows(ctx context.Context) (int64, error) {
	tables, err := q.Estimate(ctx)
	if err != nil {
		return 0, err
	}
	var rows int64
	for i := range tables {
		rows += tables[i].Rows
	}
	return rows, nil
}

type TableEstimation struct {
	Database string
	Table    string
	Rows     int64
	Parts    int64
	Marks    int64
}

func (q *SelectQuery) Estimate(ctx context.Context) ([]TableEstimation, error) {
	if q.err != nil {
		return nil, q.err
	}
	var tables []TableEstimation
	model, err := q.newModel(&tables)
	if err != nil {
		return nil, err
	}
	var queryBytes []byte
	queryBytes = append(queryBytes, "EXPLAIN ESTIMATE "...)
	queryBytes, err = q.appendQuery(q.db.fmter, queryBytes, true)
	if err != nil {
		return nil, err
	}
	query := unsafeconv.String(queryBytes)
	if _, err := q.query(ctx, model, query); err != nil {
		return nil, err
	}
	return tables, nil
}
func (q *SelectQuery) Count(ctx context.Context) (int64, error) {
	if q.err != nil {
		return 0, q.err
	}
	queryBytes, err := q.appendQuery(q.db.fmter, nil, true)
	if err != nil {
		return 0, err
	}
	query := unsafeconv.String(queryBytes)
	var count int64
	if err := q.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}
func (q *SelectQuery) ScanAndCount(ctx context.Context, values ...any) (count int64, firstErr error) {
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
