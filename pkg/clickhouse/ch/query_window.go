package ch

import (
	"github.com/uptrace/pkg/clickhouse/ch/chschema"
	"github.com/uptrace/pkg/unsafeconv"
)

type WindowQuery struct {
	baseQuery
	partition []chschema.QueryWithArgs
	order     []chschema.QueryWithArgs
	rows      chschema.QueryWithArgs
}

func NewWindowQuery(db *DB) *WindowQuery { return &WindowQuery{baseQuery: baseQuery{db: db, conn: db}} }
func (q *WindowQuery) Partition(partitions ...string) *WindowQuery {
	for _, partition := range partitions {
		q.partition = append(q.partition, chschema.UnsafeName(partition))
	}
	return q
}
func (q *WindowQuery) PartitionExpr(partition string, args ...any) *WindowQuery {
	q.partition = append(q.partition, chschema.SafeQuery(partition, args))
	return q
}
func (q *WindowQuery) Order(orders ...string) *WindowQuery {
	for _, order := range orders {
		q.order = append(q.order, chschema.UnsafeName(order))
	}
	return q
}
func (q *WindowQuery) OrderExpr(order string, args ...any) *WindowQuery {
	q.order = append(q.order, chschema.SafeQuery(order, args))
	return q
}
func (q *WindowQuery) Rows(query string, args ...any) *WindowQuery {
	q.rows = chschema.SafeQuery(query, args)
	return q
}
func (q *WindowQuery) String() string {
	b, err := q.AppendQuery(q.db.Formatter(), nil)
	if err != nil {
		return err.Error()
	}
	return unsafeconv.String(b)
}
func (q *WindowQuery) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}
	if len(q.partition) > 0 {
		b = append(b, "PARTITION BY "...)
		for i, f := range q.partition {
			if i > 0 {
				b = append(b, ", "...)
			}
			b, err = f.AppendQuery(fmter, b)
			if err != nil {
				return nil, err
			}
		}
	}
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
	if !q.rows.IsEmpty() {
		b = append(b, " ROWS "...)
		b, err = q.rows.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}
