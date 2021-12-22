package ch_test

import (
	"fmt"
	"path/filepath"
	"testing"

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
