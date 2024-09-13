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
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/klauspost/compress/gzhttp"
	_ "github.com/mostynb/go-grpc-compression/snappy"
	_ "github.com/mostynb/go-grpc-compression/zstd"
	"github.com/rs/cors"
	"github.com/uptrace/bun/migrate"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace"
	"github.com/uptrace/uptrace/cmd/uptrace/command"
	"github.com/uptrace/uptrace/pkg"
	"github.com/uptrace/uptrace/pkg/alerting"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunapp/chmigrations"
	"github.com/uptrace/uptrace/pkg/bunapp/pgmigrations"
	"github.com/uptrace/uptrace/pkg/grafana"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/run"
	"github.com/uptrace/uptrace/pkg/uptracebundle"
	"github.com/vmihailenco/taskq/extra/oteltaskq/v4"
	"golang.org/x/net/http2"

	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/tracing"
	"github.com/urfave/cli/v2"
	"github.com/wneessen/go-mail"
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
			emailTestCommand,
			command.NewCHCommand(chmigrations.Migrations),
			command.NewBunCommand(pgmigrations.Migrations),
			command.NewTemplateCommand(),
			command.NewCHSchemaCommand(),
			command.NewConfigCommand(),
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

// UPTRACE_CONFIG=config/uptrace.yml go run cmd/uptrace/main.go email-test --to uptrace@localhost
var emailTestCommand = &cli.Command{
	Name:  "email-test",
	Usage: "send test email",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "to",
			Usage:    "recipient email address",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		_, app, err := uptracebundle.StartCLI(c)
		if err != nil {
			return fmt.Errorf("failed to start app: %w", err)
		}
		defer app.Stop()

		client, err := app.InitMailer()
		if err != nil {
			return fmt.Errorf("failed to initialize mailer: %w", err)
		}

		recipient := c.String("to")

		msg := mail.NewMsg()
		msg.AddTo(recipient)
		msg.SetBodyString(mail.TypeTextPlain, "This is a test email")

		err = client.Send(msg)
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}

		fmt.Println("Test email sent successfully to", recipient)
		return nil
	},
}

var serveCommand = &cli.Command{
	Name:  "serve",
	Usage: "run HTTP and gRPC APIs",
	Action: func(c *cli.Context) error {
		ctx, app, err := uptracebundle.StartCLI(c)
		if err != nil {
			return err
		}
		defer app.Stop()

		conf := app.Config()
		logger := app.Logger

		fmt.Printf("read the docs at            https://uptrace.dev/get/\n")
		fmt.Printf("changelog                   https://github.com/uptrace/uptrace/blob/master/CHANGELOG.md\n")
		fmt.Printf("Telegram chat               https://t.me/uptrace\n")
		fmt.Printf("Slack chat                  https://join.slack.com/t/uptracedev/shared_invite/zt-1xr19nhom-cEE3QKSVt172JdQLXgXGvw\n")
		fmt.Println()

		fmt.Printf("Open UI (site.addr)         %s\n", conf.SiteURL("/"))
		fmt.Println()

		app.Logger.Info("starting Uptrace...",
			zap.String("version", pkg.Version()),
			zap.String("config", conf.Path))

		httpLn, err := net.Listen("tcp", conf.Listen.HTTP.Addr)
		if err != nil {
			logger.Error("net.Listen failed (edit listen.http YAML option)",
				zap.Error(err), zap.String("addr", conf.Listen.HTTP.Addr))
			return err
		}

		grpcLn, err := net.Listen("tcp", conf.Listen.GRPC.Addr)
		if err != nil {
			logger.Error("net.Listen failed (edit listen.grpc YAML option)",
				zap.String("addr", conf.Listen.GRPC.Addr),
				zap.Error(err))
			return err
		}

		if err := initPostgres(ctx, app); err != nil {
			return fmt.Errorf("initPostgres failed: %w", err)
		}
		if err := initClickhouse(ctx, app); err != nil {
			return fmt.Errorf("initClickhouse failed: %w", err)
		}
		if err := loadInitialData(ctx, app); err != nil {
			return fmt.Errorf("loadInitialData failed: %w", err)
		}

		org.Init(ctx, app)
		tracing.Init(ctx, app)
		metrics.Init(ctx, app)
		alerting.Init(ctx, app)
		grafana.Init(ctx, app)

		if err := syncDashboards(ctx, app); err != nil {
			app.Zap(ctx).Error("syncDashboards failed", zap.Error(err))
		}

		var group run.Group

		{
			handleStaticFiles(app, uptrace.DistFS())
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
				ReadTimeout:  10 * time.Second,
				WriteTimeout: time.Minute,
				IdleTimeout:  3 * time.Minute,
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

		{
			man := alerting.NewManager(app, &alerting.ManagerConfig{
				Logger: app.Logger.Logger,
			})

			group.Add(func() error {
				man.Run()
				return nil
			}, func(err error) {
				man.Stop()
			})
		}

		{
			group.Add(func() error {
				consumer := app.MainQueue.Consumer()
				consumer.AddHook(oteltaskq.NewHook())

				if err := consumer.Start(ctx); err != nil {
					return err
				}

				<-app.Done()
				return nil
			}, func(err error) {
				app.MainQueue.Close()
			})
		}

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

		runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return group.Run(runCtx)
	},
}

func initPostgres(ctx context.Context, app *bunapp.App) error {
	if err := app.PG.Ping(); err != nil {
		return fmt.Errorf("PostgreSQL Ping failed: %w", err)
	}

	if err := app.WithGlobalLock(ctx, func() error {
		return runPGMigrations(ctx, app)
	}); err != nil {
		return err
	}

	return nil
}

func runPGMigrations(ctx context.Context, app *bunapp.App) error {
	migrator := migrate.NewMigrator(app.PG, pgmigrations.Migrations)

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

	app.Logger.Info("migrated PostgreSQL database", zap.String("migrations", group.String()))
	return nil
}

func initClickhouse(ctx context.Context, app *bunapp.App) error {
	if err := app.CH.Ping(ctx); err != nil {
		return fmt.Errorf("ClickHouse Ping failed: %w", err)
	}

	if err := app.WithGlobalLock(ctx, func() error {
		return runCHMigrations(ctx, app)
	}); err != nil {
		return err
	}

	if chSchema := app.Config().CHSchema; chSchema.Cluster != "" {
		if err := validateCHCluster(ctx, app); err != nil {
			return err
		}
	}

	return nil
}

func runCHMigrations(ctx context.Context, app *bunapp.App) error {
	migrator := command.NewCHMigrator(app, chmigrations.Migrations)

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

	app.Logger.Info("migrated ClickHouse database", zap.String("migrations", group.String()))
	return nil
}

func validateCHCluster(ctx context.Context, app *bunapp.App) error {
	conf := app.Config()

	n, err := app.CH.NewSelect().
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

	n, err = app.CH.NewSelect().
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

func loadInitialData(ctx context.Context, app *bunapp.App) error {
	conf := app.Config()

	for i := range conf.Auth.Users {
		src := &conf.Auth.Users[i]

		user, err := org.NewUserFromConfig(src)
		if err != nil {
			return err
		}

		if err := user.Validate(); err != nil {
			return err
		}

		if _, err := app.PG.NewInsert().
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

		if err := createProject(ctx, app, dest); err != nil {
			return err
		}
	}

	return nil
}

func createProject(ctx context.Context, app *bunapp.App, project *org.Project) error {
	project.CreatedAt = time.Now()
	project.UpdatedAt = project.CreatedAt

	if _, err := app.PG.NewInsert().
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
	if _, err := app.PG.NewInsert().
		Model(monitor).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func handleStaticFiles(app *bunapp.App, fsys fs.FS) {
	conf := app.Config()

	fsys = newVueFS(fsys, conf.Site.URL.Path)
	httpFS := http.FS(fsys)
	fileServer := http.FileServer(httpFS)

	app.RouterGroup().GET("/*path", func(w http.ResponseWriter, req bunrouter.Request) error {
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

func syncDashboards(ctx context.Context, app *bunapp.App) error {
	projects, err := org.SelectProjects(ctx, app)
	if err != nil {
		return err
	}

	dashSyncer := metrics.NewDashSyncer(app)
	for _, project := range projects {
		if err := dashSyncer.CreateDashboardsHandler(ctx, project.ID); err != nil {
			return err
		}
	}

	return nil
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

//------------------------------------------------------------------------------
