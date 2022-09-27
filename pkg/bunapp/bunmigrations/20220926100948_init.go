package bunmigrations

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/metrics"
)

func init() {
	models := []any{
		(*metrics.Metric)(nil),
		(*metrics.Dashboard)(nil),
		(*metrics.DashEntry)(nil),
		(*metrics.DashGauge)(nil),
		(*metrics.RuleAlerts)(nil),
	}

	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		for _, model := range models {
			if _, err := db.NewCreateTable().
				Model(model).
				WithForeignKeys().
				Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		for _, model := range models {
			if _, err := db.NewDropTable().
				Model(model).
				IfExists().
				Cascade().
				Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
