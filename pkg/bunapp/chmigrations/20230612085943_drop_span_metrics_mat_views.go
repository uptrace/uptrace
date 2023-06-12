package chmigrations

import (
	"context"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *ch.DB) error {
		return nil
	}, func(ctx context.Context, db *ch.DB) error {
		app := bunapp.AppFromContext(ctx)
		conf := app.Config()
		for i := range conf.MetricsFromSpans {
			metric := &conf.MetricsFromSpans[i]
			viewName := metric.ViewName()
			if _, err := db.ExecContext(ctx, "DROP VIEW IF EXISTS ?", ch.Ident(viewName)); err != nil {
				return err
			}
		}
		return nil
	})
}
