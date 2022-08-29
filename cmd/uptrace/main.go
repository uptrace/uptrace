package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/klauspost/compress/gzhttp"
	_ "github.com/mostynb/go-grpc-compression/snappy"
	_ "github.com/mostynb/go-grpc-compression/zstd"
	"github.com/rs/cors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chmigrate"
	"github.com/uptrace/uptrace"
	"github.com/uptrace/uptrace/cmd/uptrace/command"
	"github.com/uptrace/uptrace/pkg"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunapp/bunmigrations"
	"github.com/uptrace/uptrace/pkg/bunapp/chmigrations"
	"github.com/uptrace/uptrace/pkg/metrics/alerting"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/run"
	"golang.org/x/net/http2"

	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/tracing"
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
				Value:   "",
				Usage:   "load YAML configuration from `FILE`",
				EnvVars: []string{"UPTRACE_CONFIG"},
			},
		},

		Commands: []*cli.Command{
			versionCommand,
			serveCommand,
			command.NewCHCommand(chmigrations.Migrations),
			command.NewBunCommand(bunmigrations.Migrations),
			command.NewTemplateCommand(),
			command.NewCHSchemaCommand(),
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
		logger := app.Logger

		projects := app.Config().Projects
		project := &projects[len(projects)-1]

		fmt.Printf("read the docs at            https://uptrace.dev/get/\n")
		fmt.Printf("changelog                   https://github.com/uptrace/uptrace/blob/master/CHANGELOG.md\n")
		fmt.Printf("Telegram chat               https://t.me/uptrace\n")
		fmt.Println()

		fmt.Printf("reading YAML config from    %s\n", conf.Path)
		fmt.Printf("OTLP/gRPC (listen.grpc)     %s\n", conf.GRPCDsn(project))
		fmt.Printf("OTLP/HTTP (listen.http)     %s\n", conf.HTTPDsn(project))
		fmt.Println()

		fmt.Printf("Open UI (site.addr)         %s\n", conf.SitePath("/"))
		fmt.Println()

		httpLn, err := net.Listen("tcp", conf.Listen.HTTP.Addr)
		if err != nil {
			logger.Error("net.Listen failed (edit listen.http YAML option)",
				zap.Error(err), zap.String("addr", conf.Listen.HTTP.Addr))
			return err
		}

		grpcLn, err := net.Listen("tcp", conf.Listen.GRPC.Addr)
		if err != nil {
			logger.Error("net.Listen failed (edit listen.grpc YAML option)",
				zap.Error(err), zap.String("addr", conf.Listen.GRPC.Addr))
			return err
		}

		if err := app.DB.Ping(); err != nil {
			logger.Error("SQLite Ping failed (edit sqlite.file YAML option)",
				zap.Error(err))
		}
		if err := app.CH.Ping(ctx); err != nil {
			logger.Error("ClickHouse Ping failed (edit ch.dsn YAML option)",
				zap.Error(err), zap.String("dsn", app.Config().CH.DSN))
		}

		if err := runBunMigrations(ctx, app.DB); err != nil {
			logger.Error("SQLite migrations failed",
				zap.Error(err))
		}
		if err := runCHMigrations(ctx, app.CH); err != nil {
			logger.Error("ClickHouse migrations failed",
				zap.Error(err))
		}
		if err := createUptraceMetrics(ctx, app); err != nil {
			logger.Error("createUptraceMetrics failed",
				zap.Error(err))
		}

		org.Init(ctx, app)
		tracing.Init(ctx, app)
		metrics.Init(ctx, app)

		var group run.Group

		{
			handleStaticFiles(app.Router(), uptrace.DistFS())
			handler := app.HTTPHandler()
			handler = gzhttp.GzipHandler(handler)
			handler = httputil.DecompressHandler{Next: handler}
			handler = httputil.NewTraceparentHandler(handler)
			handler = otelhttp.NewHandler(handler, "")
			handler = cors.AllowAll().Handler(handler)
			handler = httputil.PanicHandler{Next: handler}

			if conf.Listen.HTTP.TLS != nil {
				tlsConf, err := conf.Listen.HTTP.TLS.TLSConfig()
				if err != nil {
					return err
				}
				tlsConf.NextProtos = []string{http2.NextProtoTLS, "http/1.1"}
				httpLn = tls.NewListener(httpLn, tlsConf)
			}

			httpServer := &http.Server{
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
				IdleTimeout:  60 * time.Second,
				Handler:      handler,
			}

			group.Add(func() error {
				return httpServer.Serve(httpLn)
			}, func(err error) {
				if err := httpServer.Shutdown(ctx); err != nil {
					logger.Error("httpServer.Shutdown", zap.Error(err))
				}
			})
		}

		{
			grpcServer := app.GRPCServer()

			group.Add(func() error {
				return grpcServer.Serve(grpcLn)
			}, func(err error) {
				grpcServer.Stop()
			})
		}

		startAlerting(&group, app)

		{
			term := make(chan os.Signal, 1)
			signal.Notify(term, os.Interrupt, syscall.SIGTERM)

			group.Add(func() error {
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

		return group.Run(ctx)
	},
}

func runBunMigrations(ctx context.Context, db *bun.DB) error {
	migrator := migrate.NewMigrator(db, bunmigrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		return err
	}

	missing, err := migrator.MissingMigrations(ctx)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		panic("migrations have been changed\n" +
			"run `uptrace db reset` to reset the database before continuing")
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

func runCHMigrations(ctx context.Context, db *ch.DB) error {
	migrator := chmigrate.NewMigrator(db, chmigrations.Migrations)

	if err := migrator.Init(ctx); err != nil {
		return err
	}

	missing, err := migrator.MissingMigrations(ctx)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		panic("migrations have been changed\n" +
			"run `uptrace ch reset` to reset the database before continuing")
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

func createUptraceMetrics(ctx context.Context, app *bunapp.App) error {
	projects := app.Config().Projects
	for i := range projects {
		project := &projects[i]

		if _, err := metrics.UpsertMetric(ctx, app, &metrics.Metric{
			ProjectID:  project.ID,
			Name:       "uptrace.spans.duration",
			Unit:       "microseconds",
			Instrument: metrics.HistogramInstrument,
		}); err != nil {
			return err
		}
	}
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

func startAlerting(group *run.Group, app *bunapp.App) {
	type Project struct {
		id    uint32
		rules []alerting.RuleConfig
	}

	conf := app.Config()
	projectMap := make(map[uint32]*Project)
	var numValidRule int

	for i := range conf.Alerting.Rules {
		rule := &conf.Alerting.Rules[i]

		if err := rule.Validate(); err != nil {
			app.Logger.Error("rule.Validate failed", zap.Error(err))
			continue
		}

		numValidRule++

		for _, projectID := range rule.Projects {
			project, ok := projectMap[projectID]
			if !ok {
				project = &Project{
					id: projectID,
				}
				projectMap[projectID] = project
			}

			project.rules = append(project.rules, rule.RuleConfig())
		}
	}

	app.Logger.Info("starting monitoring metrics...",
		zap.Int("rules", numValidRule))

	for _, project := range projectMap {
		man := alerting.NewManager(&alerting.ManagerConfig{
			Engine:   metrics.NewAlertingEngine(app, project.id),
			Rules:    project.rules,
			AlertMan: metrics.NewAlertManager(app.DB, app.Notifier, project.id),
			Logger:   app.Logger.Logger,
		})

		group.Add(func() error {
			man.Run()
			return nil
		}, func(err error) {
			man.Stop()
		})
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
