package bunapp

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/zapr"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
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
	"github.com/vmihailenco/taskq/pgq/v4"
	"github.com/vmihailenco/taskq/v4"
	"github.com/zyedidia/generic/cache"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httperror"
)

type appCtxKey struct{}

func AppFromContext(ctx context.Context) *App {
	return ctx.Value(appCtxKey{}).(*App)
}

func ContextWithApp(ctx context.Context, app *App) context.Context {
	ctx = context.WithValue(ctx, appCtxKey{}, app)
	return ctx
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

	router        *bunrouter.Router
	routerGroup   *bunrouter.Group
	publicAPIV1   *bunrouter.Group
	internalAPIV1 *bunrouter.Group

	grpcServer *grpc.Server

	PG *bun.DB
	CH *ch.DB

	QueueFactory taskq.Factory
	MainQueue    taskq.Queue

	HTTPClient *http.Client
}

func New(ctx context.Context, conf *bunconf.Config) (*App, error) {
	app := &App{
		startTime: time.Now(),
		conf:      conf,

		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	app.undoneCtx = ContextWithApp(ctx, app)
	app.ctx, app.ctxCancel = context.WithCancel(app.undoneCtx)

	app.initZap()
	app.initRouter()
	if err := app.initGRPC(); err != nil {
		return nil, err
	}
	app.PG = app.newPG()
	app.CH = app.newCH()
	app.initTaskq()

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
	zapConf := zap.NewProductionConfig()
	zapConf.Encoding = "console"
	zapConf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := zapConf.Build()
	if err != nil {
		panic(err)
	}

	level := zap.InfoLevel
	if app.conf.Logging.Level != "" {
		level, err = zapcore.ParseLevel(app.conf.Logging.Level)
		if err != nil {
			panic(err)
		}
	}

	app.Logger = otelzap.New(logger, otelzap.WithMinLevel(level))
	zap.ReplaceGlobals(logger)
	otelzap.ReplaceGlobals(app.Logger)

	zaprLogger := zapr.NewLogger(logger)
	taskq.SetLogger(zaprLogger)

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
	if app.conf.Site.URL.Path != "/" {
		app.routerGroup = app.router.NewGroup(app.conf.Site.URL.Path)
	} else {
		app.routerGroup = app.router.NewGroup("")
	}

	app.internalAPIV1 = app.routerGroup.NewGroup("/internal/v1")
	app.publicAPIV1 = app.routerGroup.NewGroup("/api/v1")
}

func (app *App) newRouter(opts ...bunrouter.Option) *bunrouter.Router {
	opts = append(opts,
		bunrouter.WithMiddleware(reqlog.NewMiddleware(
			reqlog.WithVerbose(app.Debug()),
			reqlog.FromEnv("HTTPDEBUG", "DEBUG"),
		)),
	)

	opts = append(opts,
		bunrouter.WithMiddleware(app.httpErrorHandler),
		bunrouter.WithMiddleware(bunrouterotel.NewMiddleware(
			bunrouterotel.WithClientIP(),
		)),
	)

	router := bunrouter.New(opts...)
	return router
}

func (app *App) httpErrorHandler(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		err := next(w, req)
		if err == nil {
			return nil
		}

		ctx := req.Context()
		httpErr := httperror.From(err)
		statusCode := httpErr.HTTPStatusCode()

		data := map[string]any{
			"statusCode": statusCode,
			"error":      httpErr,
		}

		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			if statusCode >= 400 {
				trace.SpanFromContext(ctx).RecordError(err)
			}

			traceID := span.SpanContext().TraceID()
			data["traceId"] = traceID
		}

		w.WriteHeader(statusCode)
		_ = bunrouter.JSON(w, data)

		return err
	}
}

func (app *App) Router() *bunrouter.Router {
	return app.router
}

func (app *App) RouterGroup() *bunrouter.Group {
	return app.routerGroup
}

func (app *App) PublicAPIV1() *bunrouter.Group {
	return app.publicAPIV1
}

func (app *App) InternalAPIV1() *bunrouter.Group {
	return app.internalAPIV1
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
		grpc.MaxRecvMsgSize(32<<20),
		grpc.ReadBufferSize(512<<10),
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

func (app *App) newPG() *bun.DB {
	conf := app.conf.PG

	var options []pgdriver.Option

	if conf.DSN != "" {
		options = append(options, pgdriver.WithDSN(conf.DSN))
	}
	if conf.Addr != "" {
		options = append(options, pgdriver.WithAddr(conf.Addr))
	}
	if conf.User != "" {
		options = append(options, pgdriver.WithUser(conf.User))
	}
	if conf.Password != "" {
		options = append(options, pgdriver.WithPassword(conf.Password))
	}
	if conf.Database != "" {
		options = append(options, pgdriver.WithDatabase(conf.Database))
	}
	if conf.TLS != nil {
		tlsConf, err := conf.TLS.TLSConfig()
		if err != nil {
			panic(fmt.Errorf("pgdriver.tls option failed: %w", err))
		}
		options = append(options, pgdriver.WithTLSConfig(tlsConf))
	} else {
		options = append(options, pgdriver.WithInsecure(true))
	}
	if len(conf.ConnParams) > 0 {
		options = append(options, pgdriver.WithConnParams(conf.ConnParams))
	}

	pgdb := sql.OpenDB(pgdriver.NewConnector(options...))
	db := bun.NewDB(pgdb, pgdialect.New())

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithEnabled(app.Debug()),
		bundebug.FromEnv("PGDEBUG", "DEBUG"),
	))
	db.AddQueryHook(bunotel.NewQueryHook(bunotel.WithFormattedQueries(true)))

	return db
}

//------------------------------------------------------------------------------

func (app *App) newCH() *ch.DB {
	chConf := app.conf.CH

	settings := chConf.QuerySettings
	if settings == nil {
		settings = make(map[string]any)
	}
	if seconds := int(chConf.MaxExecutionTime.Seconds()); seconds > 0 {
		settings["max_execution_time"] = seconds
	}

	opts := []ch.Option{
		ch.WithQuerySettings(settings),
		ch.WithAutoCreateDatabase(true),
	}

	if chConf.DSN != "" {
		opts = append(opts, ch.WithDSN(chConf.DSN))
	}
	if chConf.Addr != "" {
		opts = append(opts, ch.WithAddr(chConf.Addr))
	}
	if chConf.User != "" {
		opts = append(opts, ch.WithUser(chConf.User))
	}
	if chConf.Password != "" {
		opts = append(opts, ch.WithPassword(chConf.Password))
	}
	if chConf.Database != "" {
		opts = append(opts, ch.WithDatabase(chConf.Database))
	}
	if app.conf.CHSchema.Cluster != "" {
		opts = append(opts, ch.WithCluster(app.conf.CHSchema.Cluster))
	}
	if chConf.TLS != nil {
		tlsConf, err := chConf.TLS.TLSConfig()
		if err != nil {
			panic(fmt.Errorf("ch.tls option failed: %w", err))
		}
		opts = append(opts, ch.WithTLSConfig(tlsConf))
	}

	if chConf.MaxExecutionTime != 0 {
		opts = append(opts, ch.WithReadTimeout(chConf.MaxExecutionTime+5*time.Second))
	}

	db := ch.Connect(opts...)
	fmter := db.Formatter().WithNamedArg("DB", ch.Safe(db.Config().Database))
	db = db.WithFormatter(fmter)

	db.AddQueryHook(chdebug.NewQueryHook(
		chdebug.WithVerbose(app.Debug()),
		chdebug.FromEnv("CHDEBUG", "DEBUG"),
	))
	db.AddQueryHook(chotel.NewQueryHook())

	return db
}

func (app *App) initTaskq() {
	app.QueueFactory = pgq.NewFactory(app.PG)
	app.MainQueue = app.RegisterQueue(&taskq.QueueConfig{
		Name:    "main",
		Storage: newLocalStorage(),
	})
}

func (app *App) RegisterQueue(conf *taskq.QueueConfig) taskq.Queue {
	queue := app.QueueFactory.RegisterQueue(conf)
	return queue
}

func (app *App) RegisterTask(name string, conf *taskq.TaskConfig) *taskq.Task {
	if conf.RetryLimit == 0 {
		conf.RetryLimit = 16
	}
	return taskq.RegisterTask(name, conf)
}

func (app *App) SiteURL(path string, args ...any) string {
	return app.conf.SiteURL(path, args...)
}

func (app *App) WithGlobalLock(ctx context.Context, fn func() error) error {
	return app.PG.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(?)", 0); err != nil {
			return err
		}
		return fn()
	})
}

//------------------------------------------------------------------------------

type localStorage struct {
	mu    sync.Mutex
	cache *cache.Cache[string, time.Time]
}

func newLocalStorage() *localStorage {
	return &localStorage{
		cache: cache.New[string, time.Time](10000),
	}
}

func (s *localStorage) Exists(ctx context.Context, key string) bool {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	if tm, ok := s.cache.Get(key); ok && now.Sub(tm) < 24*time.Hour {
		return true
	}

	s.cache.Put(key, now)
	return false
}
