package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauspost/compress/gzhttp"
	_ "github.com/mostynb/go-grpc-compression/snappy"
	_ "github.com/mostynb/go-grpc-compression/zstd"
	"github.com/rs/cors"
	"github.com/uptrace/go-clickhouse/ch"
	uptracego "github.com/uptrace/uptrace-go/uptrace"
	"github.com/urfave/cli/v2"
	"github.com/vmihailenco/taskq/extra/oteltaskq/v4"
	"github.com/vmihailenco/taskq/v4"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace"
	"github.com/uptrace/uptrace/cmd/uptrace/command"
	"github.com/uptrace/uptrace/pkg"
	"github.com/uptrace/uptrace/pkg/alerting"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunapp/chmigrations"
	"github.com/uptrace/uptrace/pkg/bunapp/pgmigrations"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/run"
	"github.com/uptrace/uptrace/pkg/tracing"
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
			command.NewBunCommand(pgmigrations.Migrations),
			command.NewTemplateCommand(),
			command.NewCHSchemaCommand(),
			command.NewConfigCommand(),
			command.NewEmailCommand(),
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
		fxApp, err := bunapp.NewApp(c.String("config"),
			org.Module,
			alerting.Module,
			metrics.Module,
			tracing.Module,

			fx.Invoke(initPostgres),
			fx.Invoke(initClickhouse),
			fx.Invoke(loadInitialData),
			fx.Invoke(runMainQueue),
			fx.Invoke(syncDashboards),
			fx.Invoke(runGRPCServer),
			fx.Invoke(runHTTPServer),
			fx.Invoke(initOpentelemetry),
			fx.Invoke(func() {
				go genSampleTrace()
			}),
			fx.Invoke(func(lc fx.Lifecycle, logger *otelzap.Logger) {
				lc.Append(fx.StopHook(func() {
					_ = logger.Sync()
				}))
			}),
			fx.Invoke(showInfo),
		)
		if err != nil {
			return err
		}

		fxApp.Run()
		return nil
	},
}

func runHTTPServer(
	lc fx.Lifecycle,
	conf *bunconf.Config,
	logger *otelzap.Logger,
	router bunapp.RouterParams,
) {
	handleStaticFiles(conf, router.RouterGroup, uptrace.DistFS())
	handler := http.Handler(router.Router)
	handler = gzhttp.GzipHandler(handler)
	handler = httputil.DecompressHandler{Next: handler}
	handler = httputil.NewTraceparentHandler(handler)
	handler = otelhttp.NewHandler(handler, "")
	handler = cors.AllowAll().Handler(handler)
	handler = httputil.PanicHandler{Next: handler}

	srv := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: time.Minute,
		IdleTimeout:  3 * time.Minute,
		Handler:      handler,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			l, err := net.Listen("tcp", conf.Listen.HTTP.Addr)
			if err != nil {
				logger.Error(
					"net.Listen failed (edit listen.http YAML option)",
					zap.Error(err), zap.String("addr", conf.Listen.HTTP.Addr),
				)
				return err
			}

			if conf.Listen.HTTP.TLS != nil {
				tlsConf, err := conf.Listen.HTTP.TLS.TLSConfig()
				if err != nil {
					return err
				}
				tlsConf.NextProtos = []string{http2.NextProtoTLS, "http/1.1"}
				l = tls.NewListener(l, tlsConf)
			}

			go func() {
				_ = srv.Serve(l)
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}

func runGRPCServer(
	group *run.Group,
	conf *bunconf.Config,
	srv *grpc.Server,
	logger *otelzap.Logger,
) error {
	ln, err := net.Listen("tcp", conf.Listen.GRPC.Addr)
	if err != nil {
		logger.Error("net.Listen failed (edit listen.grpc YAML option)",
			zap.String("addr", conf.Listen.GRPC.Addr),
			zap.Error(err))
		return err
	}

	group.Add("grpc.Serve", func() error {
		return srv.Serve(ln)
	})
	group.OnStop(func(context.Context, error) error {
		srv.Stop()
		return nil
	})

	return nil
}

func runMainQueue(lc fx.Lifecycle, logger *otelzap.Logger, mainQueue taskq.Queue) error {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				ctx := context.Background()
				consumer := mainQueue.Consumer()
				consumer.AddHook(oteltaskq.NewHook())

				if err := consumer.Start(ctx); err != nil {
					logger.Error("consumer.Start() failed", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			mainQueue.Close()
			return nil
		},
	})

	return nil
}

func showInfo(conf *bunconf.Config, logger *otelzap.Logger) {
	fmt.Printf("read the docs at            https://uptrace.dev/get/\n")
	fmt.Printf("changelog                   https://github.com/uptrace/uptrace/blob/master/CHANGELOG.md\n")
	fmt.Printf("Telegram chat               https://t.me/uptrace\n")
	fmt.Printf("Slack chat                  https://join.slack.com/t/uptracedev/shared_invite/zt-1xr19nhom-cEE3QKSVt172JdQLXgXGvw\n")
	fmt.Println()

	fmt.Printf("Open UI (site.addr)         %s\n", conf.SiteURL("/"))
	fmt.Println()

	logger.Info("starting Uptrace...",
		zap.String("version", pkg.Version()),
		zap.String("config", conf.Path))
}

func initPostgres(lc fx.Lifecycle, logger *otelzap.Logger, pg *bun.DB) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := pg.Ping(); err != nil {
				return fmt.Errorf("PostgreSQL Ping failed: %w", err)
			}

			return bunapp.WithGlobalLock(ctx, pg, func() error {
				return runPGMigrations(ctx, logger, pg)
			})
		},
	})
}

func runPGMigrations(ctx context.Context, logger *otelzap.Logger, pg *bun.DB) error {
	migrator := migrate.NewMigrator(pg, pgmigrations.Migrations)

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

	logger.Info("migrated PostgreSQL database", zap.String("migrations", group.String()))
	return nil
}

func initClickhouse(lc fx.Lifecycle, logger *otelzap.Logger, conf *bunconf.Config, pg *bun.DB, ch *ch.DB) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := ch.Ping(ctx); err != nil {
				return fmt.Errorf("ClickHouse Ping failed: %w", err)
			}

			if err := bunapp.WithGlobalLock(ctx, pg, func() error {
				return runCHMigrations(ctx, logger, conf, ch)
			}); err != nil {
				return err
			}

			if chSchema := conf.CHSchema; chSchema.Cluster != "" {
				if err := validateCHCluster(ctx, conf, ch); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

func runCHMigrations(ctx context.Context, logger *otelzap.Logger, conf *bunconf.Config, ch *ch.DB) error {
	migrator := command.NewCHMigrator(conf, ch, chmigrations.Migrations)

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

	logger.Info("migrated ClickHouse database", zap.String("migrations", group.String()))
	return nil
}

func validateCHCluster(ctx context.Context, conf *bunconf.Config, ch *ch.DB) error {
	n, err := ch.NewSelect().
		TableExpr("system.clusters").
		Where("cluster = ?", conf.CHSchema.Cluster).
		Count(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("can't find %q cluster in system.clusters (try to reset database)",
			conf.CHSchema.Cluster)
	}

	if !conf.CHSchema.Replicated {
		return nil
	}

	n, err = ch.NewSelect().
		TableExpr("system.replicas").
		Where("database = ?", conf.CH.Database).
		Count(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("can't find %q replicas in system.replicas (try to reset database)",
			conf.CH.Database)
	}

	return nil
}

func loadInitialData(lc fx.Lifecycle, conf *bunconf.Config, pg *bun.DB) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			for i := range conf.Auth.Users {
				src := &conf.Auth.Users[i]

				user, err := org.NewUserFromConfig(src)
				if err != nil {
					return err
				}

				if err := user.Validate(); err != nil {
					return err
				}

				if _, err := pg.NewInsert().
					Model(user).
					On("CONFLICT (email) DO UPDATE").
					Set("name = coalesce(EXCLUDED.name, u.name)").
					Set("avatar = EXCLUDED.avatar").
					Set("notify_by_email = EXCLUDED.notify_by_email").
					Set("auth_token = EXCLUDED.auth_token").
					Set("updated_at = now()").
					Returning("*").
					Exec(ctx); err != nil {
					return err
				}
			}

			for i := range conf.Projects {
				src := &conf.Projects[i]

				dest := &org.Project{
					ID:                  src.ID,
					Name:                src.Name,
					Token:               src.Token,
					PinnedAttrs:         src.PinnedAttrs,
					GroupByEnv:          src.GroupByEnv,
					GroupFuncsByService: src.GroupFuncsByService,
					PromCompat:          src.PromCompat,
					ForceSpanName:       src.ForceSpanName,
				}
				if err := dest.Init(); err != nil {
					return err
				}

				if err := createProject(ctx, pg, dest); err != nil {
					return err
				}
			}

			return nil
		},
	})
}

func initOpentelemetry(lc fx.Lifecycle, conf *bunconf.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			project := &conf.Projects[0]

			if conf.UptraceGo.Disabled {
				return nil
			}

			var options []uptracego.Option

			options = append(options,
				uptracego.WithServiceName(conf.Service),
				uptracego.WithDeploymentEnvironment("self-hosted"))

			if conf.UptraceGo.DSN == "" {
				dsn := org.BuildDSN(conf, project.Token)
				options = append(options, uptracego.WithDSN(dsn))
			} else {
				options = append(options, uptracego.WithDSN(conf.UptraceGo.DSN))
			}

			if conf.UptraceGo.TLS != nil {
				tlsConf, err := conf.UptraceGo.TLS.TLSConfig()
				if err != nil {
					return err
				}
				options = append(options, uptracego.WithTLSConfig(tlsConf))
			}

			uptracego.ConfigureOpentelemetry(options...)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if false {
				return uptracego.Shutdown(ctx)
			}
			return nil
		},
	})
}

func createProject(ctx context.Context, pg *bun.DB, project *org.Project) error {
	project.CreatedAt = time.Now()
	project.UpdatedAt = project.CreatedAt

	if _, err := pg.NewInsert().
		Model(project).
		On("CONFLICT (id) DO UPDATE").
		Set("name = EXCLUDED.name").
		Set("token = EXCLUDED.token").
		Set("pinned_attrs = EXCLUDED.pinned_attrs").
		Set("group_by_env = EXCLUDED.group_by_env").
		Set("group_funcs_by_service = EXCLUDED.group_funcs_by_service").
		Set("prom_compat = EXCLUDED.prom_compat").
		Set("force_span_name = EXCLUDED.force_span_name").
		Set("updated_at = EXCLUDED.updated_at").
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	if !project.UpdatedAt.Equal(project.CreatedAt) {
		return nil
	}

	monitor := &org.ErrorMonitor{
		BaseMonitor: &org.BaseMonitor{
			ProjectID:             project.ID,
			Name:                  "Notify on all errors",
			State:                 org.MonitorActive,
			NotifyEveryoneByEmail: true,

			Type: org.MonitorError,
		},
		Params: org.ErrorMonitorParams{
			NotifyOnNewErrors:       true,
			NotifyOnRecurringErrors: true,
			Matchers:                make([]org.AttrMatcher, 0),
		},
	}
	if _, err := pg.NewInsert().
		Model(monitor).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func handleStaticFiles(conf *bunconf.Config, routerGroup *bunrouter.Group, fsys fs.FS) {
	fsys = newVueFS(fsys, conf.Site.URL.Path)
	httpFS := http.FS(fsys)
	fileServer := http.FileServer(httpFS)

	routerGroup.GET("/*path", func(w http.ResponseWriter, req bunrouter.Request) error {
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

func syncDashboards(lc fx.Lifecycle, logger *otelzap.Logger, pg *bun.DB) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			projects, err := org.SelectProjects(ctx, pg)
			if err != nil {
				return err
			}

			dashSyncer := metrics.NewDashSyncer(logger, pg)
			for _, project := range projects {
				if err := dashSyncer.CreateDashboardsHandler(ctx, project.ID); err != nil {
					return err
				}
			}

			return nil
		},
	})
}

func genSampleTrace() {
	ctx := context.Background()

	tracer := otel.Tracer("github.com/uptrace/uptrace")

	ctx, main := tracer.Start(ctx, "sample-trace-by-uptrace")
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

func newVueFS(fsys fs.FS, publicPath string) *vueFS {
	return &vueFS{
		fs:         fsys,
		publicPath: publicPath,
		prefix:     strings.TrimPrefix(publicPath, "/"),
	}
}

type vueFS struct {
	fs         fs.FS
	publicPath string
	prefix     string
}

var _ fs.FS = (*vueFS)(nil)

func (v *vueFS) Open(name string) (fs.File, error) {
	if v.prefix != "" {
		name = strings.TrimPrefix(name, v.prefix)
	}

	switch filepath.Ext(name) {
	case "", ".html", ".js", ".css":
	default:
		return v.fs.Open(name)
	}

	f, err := v.fs.Open(name)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	data = bytes.ReplaceAll(data, []byte("/UPTRACE_PLACEHOLDER/"), []byte(v.publicPath))

	return &vueFile{
		f:  f,
		rd: bytes.NewReader(data),
	}, nil
}

type vueFile struct {
	f  fs.File
	rd *bytes.Reader
}

func (f *vueFile) Read(b []byte) (int, error) {
	return f.rd.Read(b)
}

func (f *vueFile) Stat() (fs.FileInfo, error) {
	info, err := f.f.Stat()
	if err != nil {
		return nil, err
	}
	return &vueFileInfo{
		FileInfo: info,
		size:     f.rd.Size(),
	}, nil
}

func (f *vueFile) Close() error {
	return f.f.Close()
}

type vueFileInfo struct {
	fs.FileInfo
	size int64
}

func (f *vueFileInfo) Size() int64 {
	return f.size
}
