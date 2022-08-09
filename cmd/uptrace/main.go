package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	kitzap "github.com/go-kit/kit/log/zap"
	"github.com/go-kit/log"
	"github.com/klauspost/compress/gzhttp"
	_ "github.com/mostynb/go-grpc-compression/snappy"
	_ "github.com/mostynb/go-grpc-compression/zstd"
	"github.com/prometheus/client_golang/prometheus"
	promconfig "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/notifier"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/rules"
	"github.com/prometheus/prometheus/util/strutil"
	"github.com/rs/cors"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chmigrate"
	"github.com/uptrace/uptrace"
	"github.com/uptrace/uptrace/pkg"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunapp/migrations"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/run"
	"gopkg.in/yaml.v2"

	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/tracing"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
			// command.AlertManager,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
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

		conf := app.Config()
		logger := app.Logger()

		// Paths should be resolved relatively to the main Uptrace config file.
		if err := os.Chdir(conf.BaseDir); err != nil {
			logger.Error("os.Chdir failed", zap.Error(err))
		}

		projects := app.Config().Projects
		project := &projects[len(projects)-1]

		fmt.Printf("current working dir         %s\n", conf.BaseDir)
		fmt.Printf("reading YAML config from    %s\n", conf.FileName)
		fmt.Printf("read the docs at            https://uptrace.dev/get/\n")
		fmt.Printf("changelog                   https://github.com/uptrace/uptrace/blob/master/CHANGELOG.md\n")
		fmt.Println()

		fmt.Printf("OTLP/gRPC (listen.grpc)     %s\n", conf.GRPCDsn(project))
		fmt.Printf("OTLP/HTTP (listen.http)     %s\n", conf.HTTPDsn(project))
		fmt.Printf("Open UI (site.addr)         %s\n", conf.SitePath("/"))
		fmt.Println()

		httpLn, err := net.Listen("tcp", conf.Listen.HTTP)
		if err != nil {
			logger.Error("net.Listen failed (edit listen.http YAML option)",
				zap.Error(err), zap.String("addr", conf.Listen.HTTP))
			return err
		}

		grpcLn, err := net.Listen("tcp", conf.Listen.GRPC)
		if err != nil {
			logger.Error("net.Listen failed (edit listen.grpc YAML option)",
				zap.Error(err), zap.String("addr", conf.Listen.GRPC))
			return err
		}

		if err := app.CH.Ping(ctx); err != nil {
			logger.Error("ClickHouse ping failed (edit ch.dsn YAML option)",
				zap.Error(err), zap.String("dsn", app.Config().CH.DSN))
		}

		if err := runMigrations(ctx, app.CH); err != nil {
			logger.Error("ClickHouse migrations failed",
				zap.Error(err))
		}

		promLogger := kitzap.NewZapSugarLogger(logger.Logger, zapcore.InfoLevel)
		promConfigReady := make(chan struct{})
		ctxNotify, cancelNotify := context.WithCancel(context.Background())

		app.NotifierManager = notifier.NewManager(&notifier.Options{
			QueueCapacity: 100,
			Registerer:    prometheus.DefaultRegisterer,
		}, promLogger)

		discoveryManagerNotifyReady := make(chan struct{})
		discoveryManagerNotify := discovery.NewManager(
			ctxNotify,
			log.With(promLogger, "component", "discovery manager notify"),
			discovery.Name("notify"),
		)

		app.QueryEngine = promql.NewEngine(promql.EngineOpts{
			Logger:             log.With(promLogger, "component", "query engine"),
			Reg:                prometheus.DefaultRegisterer,
			MaxSamples:         50000000,
			Timeout:            time.Minute,
			ActiveQueryTracker: nil,
			LookbackDelta:      5 * time.Minute,
			NoStepSubqueryIntervalFn: func(rangeMillis int64) int64 {
				return time.Duration(promconfig.DefaultGlobalConfig.EvaluationInterval).Milliseconds()
			},
			// EnableAtModifier and EnableNegativeOffset have to be
			// always on for regular PromQL as of Prometheus v2.33.
			EnableAtModifier:     true,
			EnableNegativeOffset: true,
			EnablePerStepStats:   false,
		})

		promStorage := metrics.NewPromStorage(ctx, app, 0)
		externalURL, err := url.Parse(conf.Prometheus.ExternalURL)
		if err != nil {
			return err
		}

		app.RuleManager = rules.NewManager(&rules.ManagerOptions{
			Appendable:      promStorage,
			Queryable:       promStorage,
			QueryFunc:       rules.EngineQueryFunc(app.QueryEngine, promStorage),
			NotifyFunc:      sendAlerts(app.NotifierManager, conf.Prometheus.ExternalURL),
			Context:         context.Background(),
			ExternalURL:     externalURL,
			Registerer:      prometheus.DefaultRegisterer,
			Logger:          log.With(promLogger, "component", "rule manager"),
			OutageTolerance: time.Hour,
			ForGracePeriod:  10 * time.Minute,
			ResendDelay:     time.Minute,
		})

		org.Init(ctx, app)
		tracing.Init(ctx, app)
		metrics.Init(ctx, app)

		reloaders := []reloader{
			{
				name: "notify",
				fn:   app.NotifierManager.ApplyConfig,
			},
			{
				name: "notify_sd",
				fn: func(conf *promconfig.Config) error {
					c := make(map[string]discovery.Configs)
					for k, v := range conf.AlertingConfig.AlertmanagerConfigs.ToMap() {
						c[k] = v.ServiceDiscoveryConfigs
					}
					discoveryManagerNotify.ApplyConfig(c)
					close(discoveryManagerNotifyReady)
					return nil
				},
			},
			{
				name: "rules",
				fn: func(conf *promconfig.Config) error {
					// Get all rule files matching the configuration paths.
					var files []string
					for _, pat := range conf.RuleFiles {
						fs, err := filepath.Glob(pat)
						if err != nil {
							// The only error can be a bad pattern.
							return fmt.Errorf("error retrieving rule files for %s: %w", pat, err)
						}
						files = append(files, fs...)
					}
					return app.RuleManager.Update(
						time.Duration(conf.GlobalConfig.EvaluationInterval),
						files,
						conf.GlobalConfig.ExternalLabels,
						externalURL.String(),
						nil,
					)
				},
			},
		}

		var g run.Group

		{
			handleStaticFiles(app.Router(), uptrace.DistFS())
			handler := app.HTTPHandler()
			handler = gzhttp.GzipHandler(handler)
			handler = httputil.DecompressHandler{Next: handler}
			handler = httputil.NewTraceparentHandler(handler)
			handler = otelhttp.NewHandler(handler, "")
			handler = cors.AllowAll().Handler(handler)
			handler = httputil.PanicHandler{Next: handler}

			httpServer := &http.Server{
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
				IdleTimeout:  60 * time.Second,
				Handler:      handler,
			}

			g.Add(func() error {
				return httpServer.Serve(httpLn)
			}, func(err error) {
				if err := httpServer.Shutdown(ctx); err != nil {
					logger.Error("httpServer.Shutdown", zap.Error(err))
				}
			})
		}

		{
			grpcServer := app.GRPCServer()

			g.Add(func() error {
				return grpcServer.Serve(grpcLn)
			}, func(err error) {
				grpcServer.Stop()
			})
		}

		g.Add(
			func() error {
				err := discoveryManagerNotify.Run()
				logger.Info("Notify discovery manager stopped")
				return err
			},
			func(err error) {
				logger.Info("Stopping notify discovery manager...")
				cancelNotify()
			},
		)

		g.Add(
			func() error {
				select {
				case <-promConfigReady:
				case <-app.Done():
					return nil
				}

				app.NotifierManager.Run(discoveryManagerNotify.SyncCh())
				logger.Info("Notifier manager stopped")
				return nil
			},
			func(err error) {
				app.NotifierManager.Stop()
			},
		)

		g.Add(func() error {
			select {
			case <-promConfigReady:
			case <-app.Done():
				return nil
			}

			app.RuleManager.Run()
			return nil
		}, func(err error) {
			app.RuleManager.Stop()
		})

		{
			hup := make(chan os.Signal, 1)
			signal.Notify(hup, syscall.SIGHUP)

			g.Add(func() error {
				for {
					select {
					case <-hup:
						if err := reloadPromConfig(conf.Prometheus.Config, reloaders); err != nil {
							logger.Error("reloadPromConfig failed", zap.Error(err))
						}
					case <-app.Done():
						return nil
					}
				}
			}, func(err error) {})
		}

		{
			term := make(chan os.Signal, 1)
			signal.Notify(term, os.Interrupt, syscall.SIGTERM)

			g.Add(func() error {
				if err := reloadPromConfig(conf.Prometheus.Config, reloaders); err != nil {
					logger.Error("reloadPromConfig failed", zap.Error(err))
				}
				close(promConfigReady)

				select {
				case <-term:
				case <-app.Done():
				}

				return nil
			}, func(err error) {
				app.Stop()
			})
		}

		go genSampleTrace()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return g.Run(ctx)
	},
}

func runMigrations(ctx context.Context, db *ch.DB) error {
	migrator := chmigrate.NewMigrator(db, migrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		return err
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}

	if group.ID == 0 { // no migrations
		return nil
	}

	fmt.Printf("migrated to %s\n", group)
	return nil
}

func handleStaticFiles(router *bunrouter.Router, fsys fs.FS) {
	httpFS := http.FS(fsys)
	fileServer := http.FileServer(httpFS)

	router.GET("/*path", func(w http.ResponseWriter, req bunrouter.Request) error {
		if _, err := httpFS.Open(req.URL.Path); err == nil {
			fileServer.ServeHTTP(w, req.Request)
			return nil
		}

		if !strings.HasPrefix(req.URL.Path, "/api") {
			req.URL.Path = "/"
			fileServer.ServeHTTP(w, req.Request)
			return nil
		}

		http.NotFound(w, req.Request)
		return nil
	})
}

type reloader struct {
	name string
	fn   func(*promconfig.Config) error
}

func reloadPromConfig(configFile string, reloaders []reloader) error {
	configFile, err := filepath.Abs(configFile)
	if err != nil {
		return err
	}

	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	conf := new(promconfig.Config)

	if err := yaml.Unmarshal(configBytes, conf); err != nil {
		return err
	}

	for _, reloader := range reloaders {
		if err := reloader.fn(conf); err != nil {
			return err
		}
	}
	return nil
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

func genSampleTrace() {
	ctx := context.Background()

	tracer := otel.Tracer("github.com/uptrace/uptrace")

	ctx, main := tracer.Start(ctx, "sample-trace")
	defer main.End()

	_, child1 := tracer.Start(ctx, "child1-of-main")
	child1.SetAttributes(attribute.String("key1", "value1"))
	child1.RecordError(errors.New("oh my error1"))
	child1.End()

	_, child2 := tracer.Start(ctx, "child2-of-main")
	child2.SetAttributes(attribute.Int("key2", 42), attribute.Float64("key3", 123.456))
	child2.End()
}

//------------------------------------------------------------------------------

type sender interface {
	Send(alerts ...*notifier.Alert)
}

// sendAlerts implements the rules.NotifyFunc for a Notifier.
func sendAlerts(s sender, externalURL string) rules.NotifyFunc {
	return func(ctx context.Context, expr string, alerts ...*rules.Alert) {
		var res []*notifier.Alert

		for _, alert := range alerts {
			a := &notifier.Alert{
				StartsAt:     alert.FiredAt,
				Labels:       alert.Labels,
				Annotations:  alert.Annotations,
				GeneratorURL: externalURL + strutil.TableLinkForExpression(expr),
			}
			if !alert.ResolvedAt.IsZero() {
				a.EndsAt = alert.ResolvedAt
			} else {
				a.EndsAt = alert.ValidUntil
			}
			res = append(res, a)
		}

		if len(alerts) > 0 {
			s.Send(res...)
		}
	}
}
