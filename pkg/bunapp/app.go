package bunapp

import (
	"context"
	"math/rand"
	"net/http"
	"net/http/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/bunrouterotel"
	"github.com/uptrace/bunrouter/extra/reqlog"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chdebug"
	"github.com/uptrace/go-clickhouse/chotel"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
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
		cfg, err := ReadConfig(configFile, service)
		if err == nil {
			return StartConfig(ctx, cfg)
		}
		firstErr = err
	}

	return context.TODO(), nil, firstErr
}

func StartConfig(ctx context.Context, cfg *AppConfig) (context.Context, *App, error) {
	rand.Seed(time.Now().UnixNano())

	app := New(ctx, cfg)
	if err := onStart.Run(ctx, app); err != nil {
		return nil, nil, err
	}

	setupOpentelemetry(app)

	return app.Context(), app, nil
}

type App struct {
	undoneCtx context.Context
	ctx       context.Context
	ctxCancel func()

	wg       sync.WaitGroup
	stopping uint32

	startTime time.Time
	cfg       *AppConfig

	onStop    appHooks
	onStopped appHooks

	logger *otelzap.Logger

	router   *bunrouter.Router
	apiGroup *bunrouter.Group

	grpcServer *grpc.Server

	chdb *ch.DB
}

func New(ctx context.Context, cfg *AppConfig) *App {
	app := &App{
		startTime: time.Now(),
		cfg:       cfg,
	}

	app.undoneCtx = ContextWithApp(ctx, app)
	app.ctx, app.ctxCancel = context.WithCancel(app.undoneCtx)

	app.initZap()
	app.initRouter()
	app.initGRPC()
	app.initCH()

	return app
}

func (app *App) Stop() {
	if !atomic.CompareAndSwapUint32(&app.stopping, 0, 1) {
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
	return app.cfg.Debug
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

func (app *App) Config() *AppConfig {
	return app.cfg
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

func (app *App) ZapLogger() *otelzap.Logger {
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
		statusCode := httpErr.StatusCode()

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

func (app *App) initCH() {
	db := ch.Connect(
		ch.WithDSN(app.cfg.CH.DSN),
		ch.WithQuerySettings(map[string]any{
			"prefer_column_name_to_alias": 1,
		}),
	)

	replicated := ""
	if app.cfg.CHSchema.Replicated {
		replicated = "Replicated"
	}

	compression := app.cfg.CHSchema.Compression
	if compression == "" {
		compression = "Default"
	}

	fmter := db.Formatter().
		WithNamedArg("TTL", ch.Safe(app.cfg.CHSchema.TTL)).
		WithNamedArg("REPLICATED", ch.Safe(replicated)).
		WithNamedArg("CODEC", ch.Safe(compression))
	db = db.WithFormatter(fmter)

	db.AddQueryHook(chdebug.NewQueryHook(
		chdebug.WithVerbose(app.Debug()),
		chdebug.FromEnv("DEBUG"),
	))
	db.AddQueryHook(chotel.NewQueryHook())

	app.chdb = db
}

func (app *App) CH() *ch.DB {
	return app.chdb
}
