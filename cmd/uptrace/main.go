package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/compress/gzhttp"
	"github.com/rs/cors"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/chmigrate"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace"
	"github.com/uptrace/uptrace/pkg"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunapp/migrations"
	"github.com/uptrace/uptrace/pkg/httputil"
	_ "github.com/uptrace/uptrace/pkg/tracing"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func main() {
	app := &cli.App{
		Name:  "uptrace",
		Usage: "Uptrace CLI",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Value:   "config/uptrace.yml",
				Usage:   "load YAML configuration from `FILE`",
				EnvVars: []string{"UPTRACE_CONFIG"},
			},
		},

		Commands: []*cli.Command{
			versionCommand,
			serveCommand,
			newCHCommand(migrations.Migrations),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var versionCommand = &cli.Command{
	Name:  "version",
	Usage: "print Uptrace version",
	Action: func(c *cli.Context) error {
		fmt.Println(pkg.Version())
		return nil
	},
}

var serveCommand = &cli.Command{
	Name:  "serve",
	Usage: "run HTTP and gRPC APIs",
	Action: func(c *cli.Context) error {
		ctx, app, err := bunapp.StartCLI(c)
		if err != nil {
			return err
		}
		defer app.Stop()

		cfg := app.Config()
		project := &app.Config().Projects[0]

		fmt.Printf("reading YAML config from    %s\n", cfg.Filepath)
		fmt.Printf("OTLP/gRPC (listen.grpc)     %s\n", cfg.GRPCDsn(project))
		fmt.Printf("OTLP/HTTP (listen.http)     %s\n", cfg.HTTPDsn(project))
		fmt.Println()

		fmt.Printf("read the docs at            https://docs.uptrace.dev/guide/os.html#otlp\n")
		fmt.Printf("changelog                   https://github.com/uptrace/uptrace/blob/master/CHANGELOG.md\n")
		fmt.Println()

		fmt.Printf("Open UI (listen.http)       %s\n", cfg.SiteAddr())
		fmt.Println()

		httpLn, err := net.Listen("tcp", cfg.Listen.HTTP)
		if err != nil {
			otelzap.L().Error("net.Listen failed (edit listen.http YAML option)",
				zap.Error(err), zap.String("addr", cfg.Listen.HTTP))
			return err
		}

		grpcLn, err := net.Listen("tcp", cfg.Listen.GRPC)
		if err != nil {
			otelzap.L().Error("net.Listen failed (edit listen.grpc YAML option)",
				zap.Error(err), zap.String("addr", cfg.Listen.GRPC))
			return err
		}

		if err := app.CH().Ping(ctx); err != nil {
			otelzap.L().Error("ClickHouse ping failed (edit ch.dsn YAML option)",
				zap.Error(err), zap.String("dsn", app.Config().CH.DSN))
		}

		serveVueApp(app)
		handler := app.HTTPHandler()
		handler = gzhttp.GzipHandler(handler)
		handler = httputil.DecompressHandler{Next: handler}
		handler = otelhttp.NewHandler(handler, "")
		handler = cors.AllowAll().Handler(handler)
		handler = httputil.PanicHandler{Next: handler}

		httpServer := &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  60 * time.Second,
			Handler:      handler,
		}

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := httpServer.Serve(httpLn); err != nil && err.Error() != "http: Server closed" {
				app.Zap(ctx).Error("httpServer.Serve failed", zap.Error(err))
			}
		}()

		go func() {
			if err := app.GRPCServer().Serve(grpcLn); err != nil {
				app.Zap(ctx).Error("grpcServer.Serve failed", zap.Error(err))
			}
		}()

		genSampleTrace()

		fmt.Println(bunapp.WaitExitSignal())

		if err := httpServer.Shutdown(ctx); err != nil {
			return err
		}

		wg.Wait()
		return nil
	},
}

func serveVueApp(app *bunapp.App) {
	router := app.Router()
	fsys := http.FS(uptrace.DistFS())
	fileServer := http.FileServer(fsys)

	notFoundMiddleware := func(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
		return func(w http.ResponseWriter, req bunrouter.Request) error {
			path := req.URL.Path
			if path == "/" || strings.Contains(path, "/api/") {
				return next(w, req)
			}

			if _, err := fsys.Open(req.URL.Path); err != nil {
				req.URL.Path = "/"
				router.ServeHTTP(w, req.Request)
				return nil
			}

			return next(w, req)
		}
	}

	router.NewGroup("/*path",
		bunrouter.WithMiddleware(notFoundMiddleware),
		bunrouter.WithGroup(func(group *bunrouter.Group) {
			group.GET("", bunrouter.HTTPHandler(fileServer))
		}))
}

func newCHCommand(migrations *chmigrate.Migrations) *cli.Command {
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

					db := app.CH()
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

					migrator := chmigrate.NewMigrator(app.CH(), migrations)
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

					migrator := chmigrate.NewMigrator(app.CH(), migrations)

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

					migrator := chmigrate.NewMigrator(app.CH(), migrations)

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

					migrator := chmigrate.NewMigrator(app.CH(), migrations)

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

					migrator := chmigrate.NewMigrator(app.CH(), migrations)
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

					migrator := chmigrate.NewMigrator(app.CH(), migrations)
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

					migrator := chmigrate.NewMigrator(app.CH(), migrations)

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

					migrator := chmigrate.NewMigrator(app.CH(), migrations)

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

					migrator := chmigrate.NewMigrator(app.CH(), migrations)

					ms, err := migrator.MigrationsWithStatus(ctx)
					if err != nil {
						return err
					}
					fmt.Printf("migrations: %s\n", ms)
					fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
					fmt.Printf("last migration group: %s\n", ms.LastGroup())

					return nil
				},
			},
		},
	}
}

func genSampleTrace() {
	ctx := context.Background()

	tracer := otel.Tracer("github.com/uptrace/uptrace")

	ctx, main := tracer.Start(ctx, "sample-trace")
	defer main.End()

	_, child1 := tracer.Start(ctx, "child1-of-main")
	child1.SetAttributes(attribute.String("key1", "value1"))
	child1.RecordError(errors.New("error1"))
	child1.End()

	_, child2 := tracer.Start(ctx, "child2-of-main")
	child2.SetAttributes(attribute.Int("key2", 42), attribute.Float64("key3", 123.456))
	child2.End()
}
