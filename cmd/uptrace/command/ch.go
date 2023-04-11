package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chmigrate"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
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
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					for {
						if err := app.CH.Ping(ctx); err != nil {
							conf := app.Config().CH
							app.Zap(ctx).Info("ClickHouse is down",
								zap.Error(err),
								zap.String("addr", conf.Addr),
								zap.String("user", conf.User),
								zap.String("password", conf.Password))
							time.Sleep(time.Second)
							continue
						}

						app.Zap(ctx).Info("ClickHouse is up and runnining")
						break
					}

					return nil
				},
			},
			{
				Name:  "init",
				Usage: "create migration tables",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := NewCHMigrator(app, migrations)
					return migrator.Init(ctx)
				},
			},
			{
				Name:  "migrate",
				Usage: "migrate database",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := NewCHMigrator(app, migrations)

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
				},
			},
			{
				Name:  "rollback",
				Usage: "rollback the last migration group",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := NewCHMigrator(app, migrations)

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
				},
			},
			{
				Name:  "reset",
				Usage: "reset ClickHouse schema",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := NewCHMigrator(app, migrations)

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

					if err := migrator.TruncateTable(ctx); err != nil {
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
				},
			},
			{
				Name:  "lock",
				Usage: "lock migrations",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := NewCHMigrator(app, migrations)
					return migrator.Lock(ctx)
				},
			},
			{
				Name:  "unlock",
				Usage: "unlock migrations",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := NewCHMigrator(app, migrations)
					return migrator.Unlock(ctx)
				},
			},
			{
				Name:  "create_go",
				Usage: "create Go migration",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := NewCHMigrator(app, migrations)

					name := strings.Join(c.Args().Slice(), "_")
					mf, err := migrator.CreateGoMigration(ctx, name)
					if err != nil {
						return err
					}
					fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)

					return nil
				},
			},
			{
				Name:  "create_sql",
				Usage: "create up and down SQL migrations",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := NewCHMigrator(app, migrations)

					name := strings.Join(c.Args().Slice(), "_")
					files, err := migrator.CreateSQLMigrations(ctx, name)
					if err != nil {
						return err
					}

					for _, mf := range files {
						fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "print migrations status",
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					migrator := NewCHMigrator(app, migrations)

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
				},
			},
		},
	}
}

func NewCHMigrator(app *bunapp.App, migrations *chmigrate.Migrations) *chmigrate.Migrator {
	chSchema := app.Config().CHSchema

	args := make(map[string]any)
	args["CLUSTER"] = ch.Safe(chSchema.Cluster)
	args["CODEC"] = ch.Safe(defaultValue(chSchema.Compression, "Default"))

	if chSchema.Replicated {
		args["REPLICATED"] = ch.Safe("Replicated")
	} else {
		args["REPLICATED"] = ch.Safe("")
	}

	if chSchema.Replicated {
		args["ON_CLUSTER"] = ch.Safe("ON CLUSTER " + chSchema.Cluster)
	} else {
		args["ON_CLUSTER"] = ch.Safe("")
	}

	args["SPANS_STORAGE"] = defaultValue(chSchema.Spans.StoragePolicy, "default")
	args["SPANS_TTL"] = ch.Safe(chSchema.Spans.TTLDelete)

	args["METRICS_STORAGE"] = defaultValue(chSchema.Metrics.StoragePolicy, "default")
	args["METRICS_TTL"] = ch.Safe(chSchema.Metrics.TTLDelete)

	fmter := app.CH.Formatter()
	for k, v := range args {
		fmter = fmter.WithNamedArg(k, v)
	}

	return chmigrate.NewMigrator(app.CH.WithFormatter(fmter), migrations)
}

func defaultValue(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}
