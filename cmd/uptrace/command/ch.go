package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/go-clickhouse/chmigrate"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunconf"
)

func NewCHCommand(migrations *chmigrate.Migrations) *cli.Command {
	return &cli.Command{
		Name:  "ch",
		Usage: "ClickHouse management commands",
		Subcommands: []*cli.Command{
			{
				Name:  "wait",
				Usage: "wait until ClickHouse is up and running",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, chWait)
				},
			},
			{
				Name:  "init",
				Usage: "create migration tables",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, chInit, fx.Supply(migrations))
				},
			},
			{
				Name:  "migrate",
				Usage: "migrate database",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, chMigrate, fx.Supply(migrations))
				},
			},
			{
				Name:  "rollback",
				Usage: "rollback the last migration group",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, chRollback, fx.Supply(migrations))
				},
			},
			{
				Name:  "reset",
				Usage: "reset ClickHouse schema",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, chReset, fx.Supply(migrations))
				},
			},
			{
				Name:  "lock",
				Usage: "lock migrations",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, chLock, fx.Supply(migrations))
				},
			},
			{
				Name:  "unlock",
				Usage: "unlock migrations",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, chUnlock, fx.Supply(migrations))
				},
			},
			{
				Name:  "create_go",
				Usage: "create Go migration",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, chCreateGo, fx.Supply(migrations))
				},
			},
			{
				Name:  "create_sql",
				Usage: "create up and down SQL migrations",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, chCreateSQL, fx.Supply(migrations))
				},
			},
			{
				Name:  "status",
				Usage: "print migrations status",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, chStatus, fx.Supply(migrations))
				},
			},
		},
	}
}

func NewCHMigrator(conf *bunconf.Config, chdb *ch.DB, migrations *chmigrate.Migrations) *chmigrate.Migrator {
	chSchema := conf.CHSchema

	args := make(map[string]any)
	var options []chmigrate.MigratorOption

	args["CLUSTER"] = ch.Safe(chSchema.Cluster)
	args["CODEC"] = ch.Safe(defaultValue(chSchema.Compression, "Default"))

	if chSchema.Replicated {
		args["REPLICATED"] = ch.Safe("Replicated")

		cluster := chschema.FormatQuery("?", ch.Ident(chSchema.Cluster))
		args["ON_CLUSTER"] = ch.Safe("ON CLUSTER " + cluster)

		options = append(options,
			chmigrate.WithReplicated(chSchema.Replicated),
			chmigrate.WithOnCluster(chSchema.Cluster),
			chmigrate.WithDistributed(true))
	} else {
		args["REPLICATED"] = ch.Safe("")
		args["ON_CLUSTER"] = ch.Safe("")
	}

	args["SPANS_STORAGE"] = defaultValue(chSchema.Spans.StoragePolicy, "default")
	args["SPANS_TTL"] = ch.Safe(chSchema.Spans.TTLDelete)

	args["METRICS_STORAGE"] = defaultValue(chSchema.Metrics.StoragePolicy, "default")
	args["METRICS_TTL"] = ch.Safe(chSchema.Metrics.TTLDelete)

	fmter := chdb.Formatter()
	for k, v := range args {
		fmter = fmter.WithNamedArg(k, v)
	}

	return chmigrate.NewMigrator(chdb.WithFormatter(fmter), migrations, options...)
}

func defaultValue(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}

// Subcommands

func chWait(lc fx.Lifecycle, logger *otelzap.Logger, conf *bunconf.Config, ch *ch.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) {
		for {
			if err := ch.Ping(ctx); err != nil {
				logger.Info("ClickHouse is down",
					zap.Error(err),
					zap.String("addr", conf.CH.Addr),
					zap.String("user", conf.CH.User))
				time.Sleep(time.Second)
				continue
			}

			logger.Info("ClickHouse is up and runnining")
			break
		}
	}))
}

func chInit(lc fx.Lifecycle, migrations *chmigrate.Migrations, conf *bunconf.Config, ch *ch.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := NewCHMigrator(conf, ch, migrations)
		return migrator.Init(ctx)
	}))
}

func chMigrate(lc fx.Lifecycle, migrations *chmigrate.Migrations, conf *bunconf.Config, ch *ch.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := NewCHMigrator(conf, ch, migrations)

		ctx = bunconf.ContextWithConfig(ctx, conf)
		group, err := migrator.Migrate(ctx)
		if err != nil {
			return err
		}

		if group.ID == 0 {
			fmt.Printf("there are no new migrations to run\n")
			return nil
		}

		fmt.Printf("migrated to %s\n", group)
		return nil
	}))
}

func chRollback(lc fx.Lifecycle, migrations *chmigrate.Migrations, conf *bunconf.Config, ch *ch.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := NewCHMigrator(conf, ch, migrations)

		ctx = bunconf.ContextWithConfig(ctx, conf)
		group, err := migrator.Rollback(ctx)
		if err != nil {
			return err
		}

		if group.ID == 0 {
			fmt.Printf("there are no groups to roll back\n")
			return nil
		}

		fmt.Printf("rolled back %s\n", group)
		return nil
	}))
}

func chReset(lc fx.Lifecycle, migrations *chmigrate.Migrations, conf *bunconf.Config, ch *ch.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := NewCHMigrator(conf, ch, migrations)
		ctx = bunconf.ContextWithConfig(ctx, conf)

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

func chLock(lc fx.Lifecycle, migrations *chmigrate.Migrations, conf *bunconf.Config, ch *ch.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := NewCHMigrator(conf, ch, migrations)
		return migrator.Lock(ctx)
	}))
}

func chUnlock(lc fx.Lifecycle, migrations *chmigrate.Migrations, conf *bunconf.Config, ch *ch.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := NewCHMigrator(conf, ch, migrations)
		return migrator.Unlock(ctx)
	}))
}

func chCreateGo(
	lc fx.Lifecycle,
	c *cli.Context,
	migrations *chmigrate.Migrations,
	conf *bunconf.Config,
	ch *ch.DB,
) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := NewCHMigrator(conf, ch, migrations)

		name := strings.Join(c.Args().Slice(), "_")
		mf, err := migrator.CreateGoMigration(ctx, name)
		if err != nil {
			return err
		}
		fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)

		return nil
	}))
}

func chCreateSQL(
	lc fx.Lifecycle,
	c *cli.Context,
	migrations *chmigrate.Migrations,
	conf *bunconf.Config,
	ch *ch.DB,
) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := NewCHMigrator(conf, ch, migrations)

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

func chStatus(lc fx.Lifecycle, migrations *chmigrate.Migrations, conf *bunconf.Config, ch *ch.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		migrator := NewCHMigrator(conf, ch, migrations)

		ms, err := migrator.MigrationsWithStatus(ctx)
		if err != nil {
			return err
		}

		unapplied := ms.Unapplied()
		if len(unapplied) > 0 {
			fmt.Printf("You have %d unapplied migrations\n", len(unapplied))
		} else {
			fmt.Printf("The database is up to date\n")
		}
		fmt.Printf("Migrations: %s\n", ms)
		fmt.Printf("Unapplied migrations: %s\n", unapplied)
		fmt.Printf("Last migration group: %s\n", ms.LastGroup())

		return nil
	}))
}
