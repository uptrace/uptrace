package bunapp

import (
	"context"
	"math/rand"
	"net/http"
	"net/http/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/prometheus/notifier"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/rules"
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
	"google.golang.org/grpc"
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

func Start(ctx context.Context, configFile, service string) (context.Context, *App, error) {
	files := []string{
		configFile,
		"uptrace.yml",
		"config/uptrace.yml",
		"/etc/uptrace/uptrace.yml",
	}

	var firstErr error

	for _, configFile := range files {
		conf, err := bunconf.ReadConfig(configFile, service)
		if err == nil {
			return StartConfig(ctx, conf)
		}
		firstErr = err
	}

	return context.TODO(), nil, firstErr
}

func StartConfig(ctx context.Context, conf *bunconf.Config) (context.Context, *App, error) {
	rand.Seed(time.Now().UnixNano())

	app := New(ctx, conf)

	switch conf.Service {
	case "serve":
		setupOpentelemetry(app)
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

	logger *otelzap.Logger

	router   *bunrouter.Router
	apiGroup *bunrouter.Group

	grpcServer *grpc.Server

	CH *ch.DB

	QueryEngine     *promql.Engine
	NotifierManager *notifier.Manager
	RuleManager     *rules.Manager
}

func New(ctx context.Context, conf *bunconf.Config) *App {
	app := &App{
		startTime: time.Now(),
		conf:      conf,
	}

	app.undoneCtx = ContextWithApp(ctx, app)
	app.ctx, app.ctxCancel = context.WithCancel(app.undoneCtx)

	app.initZap()
	app.initRouter()
	app.initGRPC()
	app.CH = app.newCH()

	return app
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

func (app *App) Debug() bool {
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
	var zapLogger *zap.Logger
	var err error
	if app.Debug() {
		zapLogger, err = zap.NewDevelopment()
	} else {
		zapLogger, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}

	app.logger = otelzap.New(zapLogger, otelzap.WithMinLevel(zap.InfoLevel))
	otelzap.ReplaceGlobals(app.logger)

	app.OnStopped("zap", func(ctx context.Context, _ *App) error {
		_ = app.logger.Sync()
		return nil
	})
}

func (app *App) Logger() *otelzap.Logger {
	return app.logger
}

func (app *App) Zap(ctx context.Context) otelzap.LoggerWithCtx {
	return app.logger.Ctx(ctx)
}

//------------------------------------------------------------------------------

func (app *App) initRouter() {
	app.router = app.newRouter()

	app.apiGroup = app.router.NewGroup("/api")
}

func (app *App) newRouter(opts ...bunrouter.Option) *bunrouter.Router {
	opts = append(opts,
		bunrouter.WithMiddleware(reqlog.NewMiddleware(
			reqlog.WithVerbose(app.Debug()),
			reqlog.FromEnv("DEBUG"),
		)),
	)

	opts = append(opts,
		bunrouter.WithMiddleware(app.httpErrorHandler),
		bunrouter.WithMiddleware(bunrouterotel.NewMiddleware(
			bunrouterotel.WithClientIP(),
		)),
	)

	router := bunrouter.New(opts...)

	if app.Debug() {
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

func (app *App) initGRPC() {
	app.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.ReadBufferSize(512<<10),
	)
}

func (app *App) GRPCServer() *grpc.Server {
	return app.grpcServer
}

//------------------------------------------------------------------------------

func (app *App) newCH() *ch.DB {
	opts := []ch.Option{
		ch.WithQuerySettings(map[string]any{
			"prefer_column_name_to_alias": 1,
		}),
		ch.WithAutoCreateDatabase(true),
	}

	conf := app.conf.CH
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

	db := ch.Connect(opts...)
	chSchema := app.Config().CHSchema

	compression := chSchema.Compression
	if compression == "" {
		compression = "Default"
	}

	replicated := ""
	if chSchema.Replicated {
		replicated = "Replicated"
	}

	onCluster := ""
	if chSchema.Replicated {
		onCluster = "ON CLUSTER " + chSchema.Cluster
	}

	fmter := db.Formatter().
		WithNamedArg("DB", ch.Safe(db.Config().Database)).
		WithNamedArg("TTL", ch.Safe(app.conf.CHSchema.TTL)).
		WithNamedArg("REPLICATED", ch.Safe(replicated)).
		WithNamedArg("CODEC", ch.Safe(compression)).
		WithNamedArg("CLUSTER", ch.Safe(app.conf.CHSchema.Cluster)).
		WithNamedArg("ON_CLUSTER", ch.Safe(onCluster))

	db = db.WithFormatter(fmter)

	db.AddQueryHook(chdebug.NewQueryHook(
		chdebug.WithVerbose(app.Debug()),
		chdebug.FromEnv("DEBUG"),
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
