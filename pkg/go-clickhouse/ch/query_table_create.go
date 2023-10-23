package ch

import (
	"context"
	"database/sql"

	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/go-clickhouse/ch/internal"
)

type CreateTableQuery struct {
	baseQuery

	ifNotExists bool
	as          chschema.QueryWithArgs
	onCluster   chschema.QueryWithArgs
	engine      chschema.QueryWithArgs
	ttl         chschema.QueryWithArgs
	partition   chschema.QueryWithArgs
	order       chschema.QueryWithArgs
}

var _ Query = (*CreateTableQuery)(nil)

func NewCreateTableQuery(db *DB) *CreateTableQuery {
	return &CreateTableQuery{
		baseQuery: baseQuery{
			db: db,
		},
	}
}

func (q *CreateTableQuery) Model(model any) *CreateTableQuery {
	q.setTableModel(model)
	return q
}

func (q *CreateTableQuery) Apply(fn func(*CreateTableQuery) *CreateTableQuery) *CreateTableQuery {
	return fn(q)
}

//------------------------------------------------------------------------------

func (q *CreateTableQuery) Table(tables ...string) *CreateTableQuery {
	for _, table := range tables {
		q.addTable(chschema.UnsafeName(table))
	}
	return q
}

func (q *CreateTableQuery) TableExpr(query string, args ...any) *CreateTableQuery {
	q.addTable(chschema.SafeQuery(query, args))
	return q
}

func (q *CreateTableQuery) ModelTable(table string) *CreateTableQuery {
	q.modelTableName = chschema.UnsafeName(table)
	return q
}

func (q *CreateTableQuery) ModelTableExpr(query string, args ...any) *CreateTableQuery {
	q.modelTableName = chschema.SafeQuery(query, args)
	return q
}

func (q *CreateTableQuery) As(table string) *CreateTableQuery {
	q.as = chschema.UnsafeName(table)
	return q
}

func (q *CreateTableQuery) ColumnExpr(query string, args ...any) *CreateTableQuery {
	q.addColumn(chschema.SafeQuery(query, args))
	return q
}

//------------------------------------------------------------------------------

func (q *CreateTableQuery) IfNotExists() *CreateTableQuery {
	q.ifNotExists = true
	return q
}

func (q *CreateTableQuery) OnCluster(cluster string) *CreateTableQuery {
	q.onCluster = chschema.UnsafeName(cluster)
	return q
}

func (q *CreateTableQuery) Engine(query string, args ...any) *CreateTableQuery {
	q.engine = chschema.SafeQuery(query, args)
	return q
}

func (q *CreateTableQuery) TTL(query string, args ...any) *CreateTableQuery {
	q.ttl = chschema.SafeQuery(query, args)
	return q
}

func (q *CreateTableQuery) Partition(query string, args ...any) *CreateTableQuery {
	q.partition = chschema.SafeQuery(query, args)
	return q
}

func (q *CreateTableQuery) Order(query string, args ...any) *CreateTableQuery {
	q.order = chschema.SafeQuery(query, args)
	return q
}

func (q *CreateTableQuery) Setting(query string, args ...any) *CreateTableQuery {
	q.settings = append(q.settings, chschema.SafeQuery(query, args))
	return q
}

//------------------------------------------------------------------------------

func (q *CreateTableQuery) Operation() string {
	return "CREATE TABLE"
}

var _ chschema.QueryAppender = (*CreateTableQuery)(nil)

func (q *CreateTableQuery) AppendQuery(fmter chschema.Formatter, b []byte) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}
	b = append(b, "CREATE TABLE "...)
	if q.ifNotExists {
		b = append(b, "IF NOT EXISTS "...)
	}

	b, err = q.appendFirstTable(fmter, b)
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

	if !q.as.IsEmpty() {
		b = append(b, " AS "...)
		b, err = q.as.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	if q.table != nil {
		b = append(b, " ("...)

		for i, field := range q.table.Fields {
			if i > 0 {
				b = append(b, ", "...)
			}

			b = append(b, field.CHName...)
			b = append(b, " "...)
			b = append(b, field.CHType...)
			if field.NotNull {
				b = append(b, " NOT NULL"...)
			}
			if field.CHDefault != "" {
				b = append(b, " DEFAULT "...)
				b = append(b, field.CHDefault...)
			}
		}

		for i, col := range q.columns {
			if i > 0 || len(q.table.Fields) > 0 {
				b = append(b, ", "...)
			}
			b, err = col.AppendQuery(fmter, b)
			if err != nil {
				return nil, err
			}
		}

		b = append(b, ")"...)
	}

	b = append(b, " Engine = "...)

	if !q.engine.IsZero() {
		b, err = q.engine.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	} else if q.table.CHEngine != "" {
		b = append(b, q.table.CHEngine...)
	} else {
		b = append(b, "MergeTree()"...)
	}

	b, err = q.appendPartition(fmter, b)
	if err != nil {
		return nil, err
	}

	if !q.order.IsZero() {
		b = append(b, " ORDER BY ("...)
		b, err = q.order.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
		b = append(b, ')')
	} else if q.table != nil {
		if len(q.table.PKs) > 0 {
			b = append(b, " ORDER BY ("...)
			for i, pk := range q.table.PKs {
				if i > 0 {
					b = append(b, ", "...)
				}
				b = append(b, pk.CHName...)
			}
			b = append(b, ')')
		} else if q.table.CHEngine == "" {
			b = append(b, " ORDER BY tuple()"...)
		}
	}

	if !q.ttl.IsZero() {
		b = append(b, " TTL "...)
		b, err = q.ttl.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	b, err = q.appendSettings(fmter, b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (q *CreateTableQuery) appendPartition(fmter chschema.Formatter, b []byte) ([]byte, error) {
	if q.partition.IsZero() && (q.table == nil || q.table.CHPartition == "") {
		return b, nil
	}

	b = append(b, " PARTITION BY "...)
	if !q.partition.IsZero() {
		return q.partition.AppendQuery(fmter, b)
	}
	return append(b, q.table.CHPartition...), nil
}

func (q *CreateTableQuery) Exec(ctx context.Context) (sql.Result, error) {
	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return nil, err
	}
	query := internal.String(queryBytes)

	return q.exec(ctx, q, query)
}
