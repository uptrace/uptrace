package command

import (
	"fmt"
	"strings"
	"time"

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

					db := app.CH
					for i := 0; i < 30; i++ {
						if err := db.Ping(ctx); err != nil {
							app.Zap(ctx).Info("ClickHouse is down",
								zap.Error(err), zap.String("dsn", app.Config().CH.DSN))
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

					migrator := chmigrate.NewMigrator(app.CH, migrations)
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

					migrator := chmigrate.NewMigrator(app.CH, migrations)

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

					migrator := chmigrate.NewMigrator(app.CH, migrations)

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

					migrator := chmigrate.NewMigrator(app.CH, migrations)

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

					migrator := chmigrate.NewMigrator(app.CH, migrations)
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

					migrator := chmigrate.NewMigrator(app.CH, migrations)
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

					migrator := chmigrate.NewMigrator(app.CH, migrations)

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

					migrator := chmigrate.NewMigrator(app.CH, migrations)

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

					migrator := chmigrate.NewMigrator(app.CH, migrations)

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
