package bunapp

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/pprof"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/extra/bunotel"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/bunrouterotel"
	"github.com/uptrace/bunrouter/extra/reqlog"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chdebug"
	"github.com/uptrace/go-clickhouse/chotel"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "modernc.org/sqlite"
)

type appCtxKey struct{}

func AppFromContext(ctx context.Context) *App {
	return ctx.Value(appCtxKey{}).(*App)
}

func ContextWithApp(ctx context.Context, app *App) context.Context {
	ctx = context.WithValue(ctx, appCtxKey{}, app)
	return ctx
}

func StartCLI(c *cli.Context) (context.Context, *App, error) {
	return Start(c.Context, c.String("config"), c.Command.Name)
}

func Start(ctx context.Context, confPath, service string) (context.Context, *App, error) {
	if confPath == "" {
		var err error
		confPath, err = findConfigPath()
		if err != nil {
			return nil, nil, err
		}
	}

	conf, err := bunconf.ReadConfig(confPath, service)
	if err != nil {
		return nil, nil, err
	}
	return StartConfig(ctx, conf)
}

func findConfigPath() (string, error) {
	files := []string{
		"uptrace.yml",
		"config/uptrace.yml",
		"/etc/uptrace/uptrace.yml",
	}
	for _, confPath := range files {
		if _, err := os.Stat(confPath); err == nil {
			return confPath, nil
		}
	}
	return "", fmt.Errorf("can't find uptrace.yml in usual locations")
}

func StartConfig(ctx context.Context, conf *bunconf.Config) (context.Context, *App, error) {
	rand.Seed(time.Now().UnixNano())

	app, err := New(ctx, conf)
	if err != nil {
		return ctx, nil, err
	}

	return app.Context(), app, nil
}

type App struct {
	undoneCtx context.Context
	ctx       context.Context
	ctxCancel func()

	wg      sync.WaitGroup
	stopped uint32

	startTime time.Time
	conf      *bunconf.Config

	onStop    appHooks
	onStopped appHooks

	Logger *otelzap.Logger

	router   *bunrouter.Router
	apiGroup *bunrouter.Group

	grpcServer *grpc.Server

	DB *bun.DB
	CH *ch.DB

	Notifier *Notifier
}

func New(ctx context.Context, conf *bunconf.Config) (*App, error) {
	app := &App{
		startTime: time.Now(),
		conf:      conf,
	}

	app.undoneCtx = ContextWithApp(ctx, app)
	app.ctx, app.ctxCancel = context.WithCancel(app.undoneCtx)

	app.initZap()
	app.initRouter()
	if err := app.initGRPC(); err != nil {
		return nil, err
	}
	app.DB = app.newDB()
	app.CH = app.newCH()
	app.Notifier = NewNotifier(conf.AlertmanagerClient.URLs)

	switch conf.Service {
	case "serve":
		if err := configureOpentelemetry(app); err != nil {
			return nil, err
		}
	}

	return app, nil
}

func (app *App) Stop() {
	if !atomic.CompareAndSwapUint32(&app.stopped, 0, 1) {
		return
	}
	app.ctxCancel()
	_ = app.onStop.Run(app.undoneCtx, app)
	_ = app.onStopped.Run(app.undoneCtx, app)
}

func (app *App) OnStop(name string, fn HookFunc) {
	app.onStop.Add(newHook(name, fn))
}

func (app *App) OnStopped(name string, fn HookFunc) {
	app.onStopped.Add(newHook(name, fn))
}

func (app *App) Debugging() bool {
	return app.conf.Debug
}

func (app *App) Context() context.Context {
	return app.ctx
}

func (app *App) Done() <-chan struct{} {
	return app.ctx.Done()
}

func (app *App) WaitGroup() *sync.WaitGroup {
	return &app.wg
}

func (app *App) Config() *bunconf.Config {
	return app.conf
}

//------------------------------------------------------------------------------

func (app *App) initZap() {
	zapConf := zap.NewProductionConfig()
	zapConf.Encoding = "console"
	zapConf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := zapConf.Build()
	if err != nil {
		panic(err)
	}

	app.Logger = otelzap.New(logger, otelzap.WithMinLevel(zap.InfoLevel))
	zap.ReplaceGlobals(logger)
	otelzap.ReplaceGlobals(app.Logger)

	app.OnStopped("zap", func(ctx context.Context, _ *App) error {
		_ = app.Logger.Sync()
		return nil
	})
}

func (app *App) Zap(ctx context.Context) otelzap.LoggerWithCtx {
	return app.Logger.Ctx(ctx)
}

//------------------------------------------------------------------------------

func (app *App) initRouter() {
	app.router = app.newRouter()
	app.apiGroup = app.router.NewGroup("/api/v1")
}

func (app *App) newRouter(opts ...bunrouter.Option) *bunrouter.Router {
	opts = append(opts,
		bunrouter.WithMiddleware(reqlog.NewMiddleware(
			reqlog.WithVerbose(app.Debugging()),
			reqlog.FromEnv("BRDEBUG", "DEBUG"),
		)),
	)

	opts = append(opts,
		bunrouter.WithMiddleware(app.httpErrorHandler),
		bunrouter.WithMiddleware(bunrouterotel.NewMiddleware(
			bunrouterotel.WithClientIP(),
		)),
	)

	router := bunrouter.New(opts...)

	if app.Debugging() {
		adapter := bunrouter.HTTPHandlerFunc

		router.GET("/debug/pprof/", adapter(pprof.Index))
		router.GET("/debug/pprof/cmdline", adapter(pprof.Cmdline))
		router.GET("/debug/pprof/profile", adapter(pprof.Profile))
		router.GET("/debug/pprof/symbol", adapter(pprof.Symbol))
		router.GET("/debug/pprof/:name", func(w http.ResponseWriter, req bunrouter.Request) error {
			h := pprof.Handler(req.Param("name"))
			h.ServeHTTP(w, req.Request)
			return nil
		})
	}

	return router
}

func (app *App) httpErrorHandler(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		err := next(w, req)
		if err == nil {
			return nil
		}

		httpErr := httperror.From(err)
		statusCode := httpErr.HTTPStatusCode()

		if statusCode >= 400 {
			trace.SpanFromContext(req.Context()).RecordError(err)
		}

		w.WriteHeader(statusCode)
		_ = bunrouter.JSON(w, httpErr)

		return err
	}
}

func (app *App) Router() *bunrouter.Router {
	return app.router
}

func (app *App) APIGroup() *bunrouter.Group {
	return app.apiGroup
}

func (app *App) HTTPHandler() http.Handler {
	return app.router
}

//------------------------------------------------------------------------------

func (app *App) initGRPC() error {
	var opts []grpc.ServerOption

	opts = append(opts,
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.ReadBufferSize(512<<10),
		grpc.MaxRecvMsgSize(32<<20),
	)

	if app.conf.Listen.GRPC.TLS != nil {
		tlsConf, err := app.conf.Listen.GRPC.TLS.TLSConfig()
		if err != nil {
			return err
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConf)))
	}

	app.grpcServer = grpc.NewServer(opts...)

	return nil
}

func (app *App) GRPCServer() *grpc.Server {
	return app.grpcServer
}

//------------------------------------------------------------------------------

func (app *App) newDB() *bun.DB {
	db := app.newBunDB()
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithEnabled(app.Debugging()),
		bundebug.FromEnv("BUNDEBUG", "DEBUG"),
	))
	db.AddQueryHook(bunotel.NewQueryHook())
	return db
}

func (app *App) newBunDB() *bun.DB {
	switch driverName := app.conf.DB.Driver; driverName {
	case "", "sqlite":
		sqldb, err := sql.Open("sqlite", app.conf.DB.DSN)
		if err != nil {
			panic(err)
		}

		db := bun.NewDB(sqldb, sqlitedialect.New())
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		return db
	case "postgres":
		sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(app.conf.DB.DSN)))
		return bun.NewDB(sqldb, pgdialect.New())
	default:
		panic(fmt.Errorf("unsupported database/sql driver: %q", driverName))
	}
}

//------------------------------------------------------------------------------

func (app *App) newCH() *ch.DB {
	conf := app.conf.CH

	settings := conf.QuerySettings
	if settings == nil {
		settings = make(map[string]any)
	}
	settings["prefer_column_name_to_alias"] = 1
	if seconds := int(conf.MaxExecutionTime.Seconds()); seconds > 0 {
		settings["max_execution_time"] = seconds
	}

	opts := []ch.Option{
		ch.WithQuerySettings(settings),
		ch.WithAutoCreateDatabase(true),
	}

	if conf.DSN != "" {
		opts = append(opts, ch.WithDSN(conf.DSN))
	}
	if conf.Addr != "" {
		opts = append(opts, ch.WithAddr(conf.Addr))
	}
	if conf.User != "" {
		opts = append(opts, ch.WithUser(conf.User))
	}
	if conf.Password != "" {
		opts = append(opts, ch.WithPassword(conf.Password))
	}
	if conf.Database != "" {
		opts = append(opts, ch.WithDatabase(conf.Database))
	}
	if conf.TLS != nil {
		tlsConf, err := conf.TLS.TLSConfig()
		if err != nil {
			panic(fmt.Errorf("ch.tls option failed: %w", err))
		}
		opts = append(opts, ch.WithTLSConfig(tlsConf))
	}

	if conf.MaxExecutionTime != 0 {
		opts = append(opts, ch.WithReadTimeout(conf.MaxExecutionTime+5*time.Second))
	}

	db := ch.Connect(opts...)
	fmter := db.Formatter().WithNamedArg("DB", ch.Safe(db.Config().Database))
	db = db.WithFormatter(fmter)

	db.AddQueryHook(chdebug.NewQueryHook(
		chdebug.WithVerbose(app.Debugging()),
		chdebug.FromEnv("CHDEBUG", "DEBUG"),
	))
	db.AddQueryHook(chotel.NewQueryHook())

	return db
}

func (app *App) DistTable(tableName string) ch.Ident {
	if app.conf.CHSchema.Cluster != "" {
		return ch.Ident(tableName + "_dist")
	}
	return ch.Ident(tableName)
}
