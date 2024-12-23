package ch

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"github.com/uptrace/pkg/unsafeconv"
)

type AlterDeleteQuery struct {
	baseQuery
	onCluster chschema.QueryWithArgs
	partition string
	where     whereQuery
}

var _ Query = (*AlterDeleteQuery)(nil)

func NewAlterDeleteQuery(db *DB) *AlterDeleteQuery {
	q := &AlterDeleteQuery{baseQuery: baseQuery{db: db, conn: db}}
	return q
}
func (q *AlterDeleteQuery) Model(model any) *AlterDeleteQuery {
	q.setTableModel(model)
	return q
}
func (q *AlterDeleteQuery) Table(tables ...string) *AlterDeleteQuery {
	for _, table := range tables {
		q.addTable(chschema.UnsafeName(table))
	}
	return q
}
func (q *AlterDeleteQuery) TableExpr(query string, args ...any) *AlterDeleteQuery {
	q.addTable(chschema.SafeQuery(query, args))
	return q
}
func (q *AlterDeleteQuery) OnCluster(cluster string) *AlterDeleteQuery {
	q.onCluster = chschema.UnsafeName(cluster)
	return q
}
func (q *AlterDeleteQuery) Setting(query string, args ...any) *AlterDeleteQuery {
	q.settings = append(q.settings, chschema.SafeQuery(query, args))
	return q
}
func (q *AlterDeleteQuery) Partition(partition string) *AlterDeleteQuery {
	q.partition = partition
	return q
}
func (q *AlterDeleteQuery) Where(query string, args ...any) *AlterDeleteQuery {
	q.where.addFilter(chschema.SafeQueryWithSep(query, args, " AND "))
	return q
}
func (q *AlterDeleteQuery) WhereOr(query string, args ...any) *AlterDeleteQuery {
	q.where.addFilter(chschema.SafeQueryWithSep(query, args, " OR "))
	return q
}
func (q *AlterDeleteQuery) WhereGroup(sep string, fn func(*AlterDeleteQuery) *AlterDeleteQuery) *AlterDeleteQuery {
	saved := q.where.filters
	q.where.filters = nil
	q = fn(q)
	filters := q.where.filters
	q.where.filters = saved
	q.where.addGroup(sep, filters)
	return q
}
func (q *AlterDeleteQuery) Operation() string { return "DELETE" }
func (q *AlterDeleteQuery) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}
	b = append(b, "ALTER TABLE "...)
	b, err = q.appendTables(fmter, b)
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
	b = append(b, " DELETE"...)
	if q.partition != "" {
		b = append(b, " IN PARTITION ID "...)
		b = chschema.AppendString(b, q.partition)
	}
	if len(q.where.filters) == 0 {
		return nil, fmt.Errorf("ALTER TABLE DELETE requires WHERE")
	}
	b = append(b, " WHERE "...)
	b, err = appendWhere(fmter, b, q.where.filters)
	if err != nil {
		return nil, err
	}
	b, err = q.appendSettings(fmter, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
func (q *AlterDeleteQuery) Exec(ctx context.Context, dest ...any) (sql.Result, error) {
	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return nil, err
	}
	query := unsafeconv.String(queryBytes)
	return q.conn.ExecContext(ctx, query)
}
