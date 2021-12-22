package ch

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/go-clickhouse/ch/chpool"
	"github.com/uptrace/go-clickhouse/ch/internal"
)

const (
	discardUnknownColumnsFlag internal.Flag = 1 << iota
)

type Config struct {
	chpool.Config

	Network  string
	Addr     string
	User     string
	Password string
	Database string

	DialTimeout   time.Duration
	TLSConfig     *tls.Config
	QuerySettings map[string]any

	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
}

func (cfg *Config) netDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   cfg.DialTimeout,
		KeepAlive: 5 * time.Minute,
	}
}

func defaultConfig() *Config {
	var cfg *Config
	poolSize := 2 * runtime.GOMAXPROCS(0)
	cfg = &Config{
		Network:  "tcp",
		Addr:     "localhost:9000",
		User:     "default",
		Database: "default",

		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,

		MaxRetries:      2,
		MinRetryBackoff: 500 * time.Millisecond,
		MaxRetryBackoff: time.Second,

		Config: chpool.Config{
			PoolSize:     poolSize,
			MaxIdleConns: poolSize,
			PoolTimeout:  30 * time.Second,
		},
	}
	return cfg
}

type Option func(db *DB)

func WithDiscardUnknownColumns() Option {
	return func(db *DB) {
		db.flags.Set(discardUnknownColumnsFlag)
	}
}

// WithNetwork configures network type, either tcp or unix.
// Default is tcp.
func WithNetwork(network string) Option {
	return func(db *DB) {
		db.cfg.Network = network
	}
}

// WithAddr configures TCP host:port or Unix socket depending on Network.
func WithAddr(addr string) Option {
	return func(db *DB) {
		db.cfg.Addr = addr
	}
}

// WithTLSConfig configures TLS config for secure connections.
func WithTLSConfig(cfg *tls.Config) Option {
	return func(db *DB) {
		db.cfg.TLSConfig = cfg
	}
}

func WithQuerySettings(params map[string]any) Option {
	return func(db *DB) {
		db.cfg.QuerySettings = params
	}
}

func WithInsecure(on bool) Option {
	return func(db *DB) {
		if on {
			db.cfg.TLSConfig = nil
		} else {
			db.cfg.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}
}

func WithUser(user string) Option {
	return func(db *DB) {
		db.cfg.User = user
	}
}

func WithPassword(password string) Option {
	return func(db *DB) {
		db.cfg.Password = password
	}
}

func WithDatabase(database string) Option {
	return func(db *DB) {
		db.cfg.Database = database
	}
}

// WithDialTimeout configures dial timeout for establishing new connections.
// Default is 5 seconds.
func WithDialTimeout(timeout time.Duration) Option {
	return func(db *DB) {
		db.cfg.DialTimeout = timeout
	}
}

// WithReadTimeout configures timeout for socket reads. If reached, commands will fail
// with a timeout instead of blocking.
func WithReadTimeout(timeout time.Duration) Option {
	return func(db *DB) {
		db.cfg.ReadTimeout = timeout
	}
}

// WithWriteTimeout configures timeout for socket writes. If reached, commands will fail
// with a timeout instead of blocking.
func WithWriteTimeout(timeout time.Duration) Option {
	return func(db *DB) {
		db.cfg.WriteTimeout = timeout
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(db *DB) {
		db.cfg.DialTimeout = timeout
		db.cfg.ReadTimeout = timeout
		db.cfg.WriteTimeout = timeout
	}
}

// WithMaxRetries configures maximum number of retries before giving up.
// Default is to retry query 2 times.
func WithMaxRetries(maxRetries int) Option {
	return func(db *DB) {
		db.cfg.MaxRetries = maxRetries
	}
}

// WithMinRetryBackoff configures minimum backoff between each retry.
// Default is 250 milliseconds; -1 disables backoff.
func WithMinRetryBackoff(backoff time.Duration) Option {
	return func(db *DB) {
		db.cfg.MinRetryBackoff = backoff
	}
}

// WithMaxRetryBackoff configures maximum backoff between each retry.
// Default is 4 seconds; -1 disables backoff.
func WithMaxRetryBackoff(backoff time.Duration) Option {
	return func(db *DB) {
		db.cfg.MaxRetryBackoff = backoff
	}
}

// WithPoolSize configures maximum number of socket connections.
// Default is 2 connections per every CPU as reported by runtime.NumCPU.
func WithPoolSize(poolSize int) Option {
	return func(db *DB) {
		db.cfg.PoolSize = poolSize
		db.cfg.MaxIdleConns = poolSize
	}
}

// WithMinIdleConns configures minimum number of idle connections which is useful when establishing
// new connection is slow.
func WithMinIdleConns(minIdleConns int) Option {
	return func(db *DB) {
		db.cfg.MinIdleConns = minIdleConns
	}
}

// WithMaxConnAge configures Connection age at which client retires (closes) the connection.
// It is useful with proxies like HAProxy.
// Default is to not close aged connections.
func WithMaxConnAge(timeout time.Duration) Option {
	return func(db *DB) {
		db.cfg.MaxConnAge = timeout
	}
}

// WithPoolTimeout configures time for which client waits for free connection if all
// connections are busy before returning an error.
// Default is 30 seconds if ReadTimeOut is not defined, otherwise,
// ReadTimeout + 1 second.
func WithPoolTimeout(timeout time.Duration) Option {
	return func(db *DB) {
		db.cfg.PoolTimeout = timeout
	}
}

func WithDSN(dsn string) Option {
	return func(db *DB) {
		opts, err := parseDSN(dsn)
		if err != nil {
			panic(err)
		}
		for _, opt := range opts {
			opt(db)
		}
	}
}

func parseDSN(dsn string) ([]Option, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	q := queryOptions{q: u.Query()}
	var opts []Option

	switch u.Scheme {
	case "ch", "clickhouse":
		if u.Host != "" {
			addr := u.Host
			if !strings.Contains(addr, ":") {
				addr += ":5432"
			}
			opts = append(opts, WithAddr(addr))
		}

		if len(u.Path) > 1 {
			opts = append(opts, WithDatabase(u.Path[1:]))
		}

		if host := q.string("host"); host != "" {
			opts = append(opts, WithAddr(host))
			if host[0] == '/' {
				opts = append(opts, WithNetwork("unix"))
			}
		}
	default:
		return nil, errors.New("ch: unknown scheme: " + u.Scheme)
	}

	if u.User != nil {
		opts = append(opts, WithUser(u.User.Username()))
		if password, ok := u.User.Password(); ok {
			opts = append(opts, WithPassword(password))
		}
	}

	switch sslMode := q.string("sslmode"); sslMode {
	case "verify-ca", "verify-full":
		opts = append(opts, WithTLSConfig(new(tls.Config)))
	case "allow", "prefer", "require", "":
		opts = append(opts, WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	case "disable":
		opts = append(opts, WithInsecure(true))
	default:
		return nil, fmt.Errorf("ch: sslmode '%s' is not supported", sslMode)
	}

	if d := q.duration("timeout"); d != 0 {
		opts = append(opts, WithTimeout(d))
	}
	if d := q.duration("dial_timeout"); d != 0 {
		opts = append(opts, WithDialTimeout(d))
	}
	if d := q.duration("read_timeout"); d != 0 {
		opts = append(opts, WithReadTimeout(d))
	}
	if d := q.duration("write_timeout"); d != 0 {
		opts = append(opts, WithWriteTimeout(d))
	}

	rem, err := q.remaining()
	if err != nil {
		return nil, q.err
	}

	if len(rem) > 0 {
		params := make(map[string]any, len(rem))
		for k, v := range rem {
			params[k] = v
		}
		opts = append(opts, WithQuerySettings(params))
	}

	return opts, nil
}

type queryOptions struct {
	q   url.Values
	err error
}

func (o *queryOptions) string(name string) string {
	vs := o.q[name]
	if len(vs) == 0 {
		return ""
	}
	delete(o.q, name) // enable detection of unknown parameters
	return vs[len(vs)-1]
}

func (o *queryOptions) duration(name string) time.Duration {
	s := o.string(name)
	if s == "" {
		return 0
	}
	// try plain number first
	if i, err := strconv.Atoi(s); err == nil {
		if i <= 0 {
			// disable timeouts
			return -1
		}
		return time.Duration(i) * time.Second
	}
	dur, err := time.ParseDuration(s)
	if err == nil {
		return dur
	}
	if o.err == nil {
		o.err = fmt.Errorf("ch: invalid %s duration: %w", name, err)
	}
	return 0
}

func (o *queryOptions) remaining() (map[string]string, error) {
	if o.err != nil {
		return nil, o.err
	}
	if len(o.q) == 0 {
		return nil, nil
	}
	m := make(map[string]string, len(o.q))
	for k, ss := range o.q {
		m[k] = ss[len(ss)-1]
	}
	return m, nil
}
