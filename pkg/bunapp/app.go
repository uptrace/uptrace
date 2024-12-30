package bunapp

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/zapr"
	"github.com/vmihailenco/taskq/pgq/v4"
	"github.com/vmihailenco/taskq/v4"
	"github.com/wneessen/go-mail"
	"github.com/zyedidia/generic/cache"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/extra/bunotel"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/pkg/clickhouse/ch"
	"github.com/uptrace/pkg/clickhouse/chdebug"
	"github.com/uptrace/pkg/clickhouse/chotel"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/run"
)

func New(configPath string, opts ...fx.Option) (*fx.App, error) {
	conf, err := initConfig(configPath)
	if err != nil {
		return nil, err
	}

	group := run.NewGroup()

	options := []fx.Option{
		fx.Supply(conf),
		fx.Supply(group),

		fx.Provide(initSlog),
		fx.Provide(initZap),
		fx.Provide(newPG),
		fx.Provide(newCH),
		fx.Provide(initRouter),
		fx.Provide(initGRPC),
		fx.Provide(fx.Annotate(initTaskq, fx.As(new(taskq.Queue)))),
		fx.Provide(newHTTPClient),
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: logger}
		}),
	}
	options = append(options, opts...)
	options = append(options, fx.Invoke(runGroup))

	app := fx.New(options...)

	group.Add("app.Done", func() error {
		sig := <-app.Done()
		return run.SignalError{Signal: sig}
	})

	return app, nil
}

type LoggerResults struct {
	fx.Out

	Logger *slog.Logger
	Level  *slog.LevelVar
}

func initSlog(conf *bunconf.Config) LoggerResults {
	var level = new(slog.LevelVar)

	switch strings.ToLower(conf.Logging.Level) {
	case "debug":
		level.Set(slog.LevelDebug)
	case "info":
		level.Set(slog.LevelInfo)
	case "warn":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}

	return LoggerResults{
		Logger: slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: level}),
		),
		Level: level,
	}
}

func initConfig(path string) (*bunconf.Config, error) {
	if path == "" {
		var err error
		path, err = findConfigPath()
		if err != nil {
			return nil, err
		}
	}

	return bunconf.ReadConfig(path, "serve")
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

func runGroup(
	lc fx.Lifecycle, shutdowner fx.Shutdowner, group *run.Group, logger *otelzap.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := group.Run(); err != nil {
					if err := shutdowner.Shutdown(); err != nil {
						logger.Error("shutdowner.Shutdown failed", zap.Error(err))
					}
				}
			}()
			return nil
		},
	})
}

func initZap(conf *bunconf.Config) *otelzap.Logger {
	zapConf := zap.NewProductionConfig()
	zapConf.Encoding = "console"
	zapConf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := zapConf.Build()
	if err != nil {
		panic(err)
	}

	level := zap.InfoLevel
	if conf.Logging.Level != "" {
		level, err = zapcore.ParseLevel(conf.Logging.Level)
		if err != nil {
			panic(err)
		}
	}

	otelzapLogger := otelzap.New(logger, otelzap.WithMinLevel(level))
	zap.ReplaceGlobals(logger)
	otelzap.ReplaceGlobals(otelzapLogger)

	zaprLogger := zapr.NewLogger(logger)
	taskq.SetLogger(zaprLogger)

	return otelzapLogger
}

//------------------------------------------------------------------------------

func initGRPC(conf *bunconf.Config) (*grpc.Server, error) {
	var opts []grpc.ServerOption

	opts = append(opts,
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.MaxRecvMsgSize(32<<20),
		grpc.ReadBufferSize(512<<10),
	)

	if conf.Listen.GRPC.TLS != nil {
		tlsConf, err := conf.Listen.GRPC.TLS.TLSConfig()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConf)))
	}

	return grpc.NewServer(opts...), nil
}

//------------------------------------------------------------------------------

func newPG(cfg *bunconf.Config) *bun.DB {
	conf := cfg.PG

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
		bundebug.WithEnabled(cfg.Debug),
		bundebug.FromEnv("PGDEBUG", "DEBUG"),
	))
	db.AddQueryHook(bunotel.NewQueryHook(bunotel.WithFormattedQueries(true)))

	return db
}

//------------------------------------------------------------------------------

func newCH(conf *bunconf.Config) *ch.DB {
	chConf := conf.CH

	settings := chConf.QuerySettings
	if settings == nil {
		settings = make(map[string]any)
	}
	if seconds := int(chConf.MaxExecutionTime.Seconds()); seconds > 0 {
		settings["max_execution_time"] = seconds
	}

	opts := []ch.Option{
		ch.WithQuerySettings(settings),
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
	if conf.CHSchema.Cluster != "" {
		opts = append(opts, ch.WithCluster(conf.CHSchema.Cluster))
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
		chdebug.WithVerbose(conf.Debug),
		chdebug.FromEnv("CHDEBUG", "DEBUG"),
	))
	db.AddQueryHook(chotel.NewQueryHook())

	return db
}

func initTaskq(pg *bun.DB) taskq.Queue {
	return pgq.NewFactory(pg).RegisterQueue(
		&taskq.QueueConfig{
			Name:    "main",
			Storage: newLocalStorage(),
		},
	)
}

func WithGlobalLock(ctx context.Context, pg *bun.DB, fn func() error) error {
	return pg.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
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

//------------------------------------------------------------------------------

func NewMailer(conf *bunconf.Config) (*mail.Client, error) {
	cfg := conf.SMTPMailer

	if !cfg.Enabled {
		return nil, fmt.Errorf("smtp_mailer is disabled in the config")
	}

	options := []mail.Option{
		mail.WithSMTPAuth(cfg.AuthType),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.Password),
	}

	switch {
	case cfg.TLS == nil:
		options = append(options,
			mail.WithTLSPortPolicy(mail.TLSOpportunistic),
			mail.WithSSLPort(false),
			mail.WithPort(cfg.Port),
		)
	case cfg.TLS.Disabled:
		options = append(options,
			mail.WithTLSPortPolicy(mail.NoTLS),
			mail.WithPort(cfg.Port),
		)
	default:
		options = append(options,
			mail.WithTLSPortPolicy(mail.TLSMandatory),
			mail.WithSSLPort(false),
			mail.WithPort(cfg.Port),
		)

		tlsCfg, err := cfg.TLS.TLSConfig()
		if err != nil {
			return nil, err
		}
		options = append(options, mail.WithTLSConfig(tlsCfg))
	}

	client, err := mail.NewClient(cfg.Host, options...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func RegisterTaskHandler(name string, handler any) *taskq.Task {
	return taskq.RegisterTask(name, &taskq.TaskConfig{
		RetryLimit: 16,
		Handler:    handler,
	})
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
	}
}
