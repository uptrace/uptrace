package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunconf"
)

func NewBunCommand(migrations *migrate.Migrations) *cli.Command {
	return &cli.Command{
		Name:  "pg",
		Usage: "PostgreSQL management commands",
		Subcommands: []*cli.Command{
			{
				Name:  "wait",
				Usage: "wait until PostgreSQL is up and running",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, pgWait)
				},
			},
			{
				Name:  "init",
				Usage: "create migration tables",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, pgInit, fx.Supply(migrations))
				},
			},
			{
				Name:  "migrate",
				Usage: "migrate database",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, pgMigrate, fx.Supply(migrations))
				},
			},
			{
				Name:  "rollback",
				Usage: "rollback the last migration group",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, pgRollback, fx.Supply(migrations))
				},
			},
			{
				Name:  "reset",
				Usage: "reset database schema",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, pgReset, fx.Supply(migrations))
				},
			},
			{
				Name:  "lock",
				Usage: "lock migrations",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, pgLock, fx.Supply(migrations))
				},
			},
			{
				Name:  "unlock",
				Usage: "unlock migrations",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, pgUnlock, fx.Supply(migrations))
				},
			},
			{
				Name:  "create_go",
				Usage: "create Go migration",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, pgCreateGo, fx.Supply(migrations), fx.Supply(c))
				},
			},
			{
				Name:  "create_sql",
				Usage: "create up and down SQL migrations",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, pgCreateSQL, fx.Supply(migrations), fx.Supply(c))
				},
			},
			{
				Name:  "status",
				Usage: "print migrations status",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, pgStatus, fx.Supply(migrations))
				},
			},
			{
				Name:  "mark_applied",
				Usage: "mark migrations as applied without actually running them",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, pgMarkApplied, fx.Supply(migrations))
				},
			},
		},
	}
}

func pgWait(lc fx.Lifecycle, logger *otelzap.Logger, conf *bunconf.Config, pg *bun.DB) {
	lc.Append(fx.StartHook(func(_ context.Context) {
		for {
			if err := pg.Ping(); err != nil {
				logger.Info("PostgreSQL is down",
					zap.Error(err),
					zap.String("addr", conf.PG.Addr),
					zap.String("user", conf.PG.User))
				time.Sleep(time.Second)
				continue
			}

			logger.Info("PostgreSQL is up and runnining")
			break
		}
	}))
}

func pgInit(lc fx.Lifecycle, migrations *migrate.Migrations, pg *bun.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := migrate.NewMigrator(pg, migrations)
		return migrator.Init(ctx)
	}))
}

func pgMigrate(lc fx.Lifecycle, migrations *migrate.Migrations, pg *bun.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := migrate.NewMigrator(pg, migrations)

		group, err := migrator.Migrate(ctx)
		if err != nil {
			return err
		}
		if group.IsZero() {
			fmt.Printf("there are no new migrations to run (database is up to date)\n")
			return nil
		}
		fmt.Printf("migrated to %s\n", group)
		return nil
	}))
}

func pgRollback(lc fx.Lifecycle, migrations *migrate.Migrations, pg *bun.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := migrate.NewMigrator(pg, migrations)

		group, err := migrator.Rollback(ctx)
		if err != nil {
			return err
		}
		if group.IsZero() {
			fmt.Printf("there are no groups to roll back\n")
			return nil
		}
		fmt.Printf("rolled back %s\n", group)
		return nil
	}))
}

func pgReset(lc fx.Lifecycle, migrations *migrate.Migrations, pg *bun.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := migrate.NewMigrator(pg, migrations)

		if err := migrator.Init(ctx); err != nil {
			return err
		}

		for {
			group, err := migrator.Rollback(ctx)
			if err != nil {
				return err
			}
			if group.ID == 0 {
				break
			}
		}

		if err := migrator.Reset(ctx); err != nil {
			return err
		}

		group, err := migrator.Migrate(ctx)
		if err != nil {
			return err
		}

		if group.ID == 0 {
			fmt.Printf("there are no new migrations to run\n")
			return nil
		}

		return nil
	}))
}

func pgLock(lc fx.Lifecycle, migrations *migrate.Migrations, pg *bun.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := migrate.NewMigrator(pg, migrations)
		return migrator.Lock(ctx)
	}))
}

func pgUnlock(lc fx.Lifecycle, migrations *migrate.Migrations, pg *bun.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := migrate.NewMigrator(pg, migrations)
		return migrator.Unlock(ctx)
	}))
}

func pgCreateGo(lc fx.Lifecycle, c *cli.Context, migrations *migrate.Migrations, pg *bun.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := migrate.NewMigrator(pg, migrations)

		name := strings.Join(c.Args().Slice(), "_")
		mf, err := migrator.CreateGoMigration(ctx, name)
		if err != nil {
			return err
		}
		fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
		return nil
	}))
}

func pgCreateSQL(lc fx.Lifecycle, c *cli.Context, migrations *migrate.Migrations, pg *bun.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := migrate.NewMigrator(pg, migrations)

		name := strings.Join(c.Args().Slice(), "_")
		files, err := migrator.CreateSQLMigrations(ctx, name)
		if err != nil {
			return err
		}

		for _, mf := range files {
			fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
		}

		return nil
	}))
}

func pgStatus(lc fx.Lifecycle, migrations *migrate.Migrations, pg *bun.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := migrate.NewMigrator(pg, migrations)

		ms, err := migrator.MigrationsWithStatus(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("migrations: %s\n", ms)
		fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
		fmt.Printf("last migration group: %s\n", ms.LastGroup())
		return nil
	}))
}

func pgMarkApplied(lc fx.Lifecycle, migrations *migrate.Migrations, pg *bun.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := migrate.NewMigrator(pg, migrations)

		group, err := migrator.Migrate(ctx, migrate.WithNopMigration())
		if err != nil {
			return err
		}
		if group.IsZero() {
			fmt.Printf("there are no new migrations to mark as applied\n")
			return nil
		}
		fmt.Printf("marked as applied %s\n", group)
		return nil
	}))
}
