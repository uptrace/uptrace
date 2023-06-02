package chmigrations

import (
	"context"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chmigrate"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *ch.DB) error {
		app := bunapp.AppFromContext(ctx)
		if app.Config().CHSchema.Cluster == "" {
			return nil
		}

		f, err := bunapp.FS().Open("sql/ch_recreate_distributed.up.sql")
		if err != nil {
			return err
		}
		return chmigrate.Exec(ctx, db, f)
	}, func(ctx context.Context, db *ch.DB) error {
		app := bunapp.AppFromContext(ctx)
		if app.Config().CHSchema.Cluster == "" {
			return nil
		}

		f, err := bunapp.FS().Open("sql/ch_recreate_distributed.up.sql")
		if err != nil {
			return err
		}
		return chmigrate.Exec(ctx, db, f)
	})
}
