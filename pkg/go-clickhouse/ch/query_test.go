package ch_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/chschema"
)

func TestQuery(t *testing.T) {
	type Model struct {
		ID     uint64
		String string
		Bytes  []byte
	}

	queries := []func(db *ch.DB) chschema.QueryAppender{
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewCreateTable().Model((*Model)(nil))
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewDropTable().Model((*Model)(nil))
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewSelect().Model((*Model)(nil))
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewSelect().Model((*Model)(nil)).ExcludeColumn("bytes")
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewInsert().Model(new(Model))
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewTruncateTable().Model(new(Model))
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewSelect().
				Model((*Model)(nil)).
				Setting("max_rows_to_read = 100")
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewSelect().
				Model((*Model)(nil)).
				Setting("max_rows_to_read = 100").
				Setting("read_overflow_mode = 'break'")
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewInsert().
				TableExpr("dest").
				TableExpr("src").
				Where("_part = ?", "part_name").
				Setting("max_threads = 1").
				Setting("max_insert_threads = 1").
				Setting("max_execution_time = 0")
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewSelect().
				Model((*Model)(nil)).
				Sample("?", 1000)
		},
		func(db *ch.DB) chschema.QueryAppender {
			type Model struct {
				ch.CHModel `ch:"table:spans,partition:toYYYYMM(time)"`

				ID   uint64
				Text string    `ch:",lc"` // low cardinality column
				Time time.Time `ch:",pk"` // ClickHouse primary key for order by
			}
			return db.NewCreateTable().Model((*Model)(nil)).
				TTL("time + INTERVAL 30 DAY DELETE").
				Partition("toDate(time)").
				Order("id").
				Setting("ttl_only_drop_parts = 1")
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewDropView().View("view_name")
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewDropView().IfExists().ViewExpr("view_name")
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewCreateView().
				Materialized().
				IfNotExists().
				View("view_name").
				To("dest_table").
				Column("col1").
				ColumnExpr("col1 AS alias").
				TableExpr("src_table AS alias").
				Where("foo = bar").
				Group("group1").
				GroupExpr("group2, group3").
				OrderExpr("order2, order3")
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewSelect().
				Model((*Model)(nil)).
				Final()
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewSelect().
				Model((*Model)(nil)).
				Where("id = ?", 1).
				Final()
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewSelect().
				Model((*Model)(nil)).
				Where("id = ?", 1).
				Final().
				Group("id").
				OrderExpr("id")
		},
		func(db *ch.DB) chschema.QueryAppender {
			q1 := db.NewSelect().Model(new(Model)).Where("1")
			q2 := db.NewSelect().Model(new(Model))
			return q1.Union(q2)
		},
		func(db *ch.DB) chschema.QueryAppender {
			q1 := db.NewSelect().Model(new(Model)).Where("1")
			q2 := db.NewSelect().Model(new(Model))
			return q1.UnionAll(q2)
		},
		func(db *ch.DB) chschema.QueryAppender {
			return db.NewCreateTable().
				Table("my-table_dist").
				As("my-table").
				Engine("Distributed(?, currentDatabase(), ?, rand())",
					ch.Name("my-cluster"), ch.Name("my-table")).
				OnCluster("my-cluster").
				IfNotExists()
		},
	}

	db := chDB()
	defer db.Close()

	snapshotsDir := filepath.Join("testdata", "snapshots")
	snapshot := cupaloy.New(cupaloy.SnapshotSubdirectory(snapshotsDir))

	for i, fn := range queries {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			q := fn(db)

			query, err := q.AppendQuery(db.Formatter(), nil)
			if err != nil {
				snapshot.SnapshotT(t, err.Error())
			} else {
				snapshot.SnapshotT(t, string(query))
			}
		})
	}
}
