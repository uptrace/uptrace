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
	autoCreateDatabaseFlag
)

type Config struct {
	chpool.Config

	Compression bool

	Addr     string
	User     string
	Password string
	Database string
	Cluster  string

	DialTimeout   time.Duration
	TLSConfig     *tls.Config
	QuerySettings map[string]any

	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
}

func (conf *Config) clone() *Config {
	clone := *conf
	return &clone
}

func (conf *Config) netDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   conf.DialTimeout,
		KeepAlive: 5 * time.Minute,
	}
}

func defaultConfig() *Config {
	var conf *Config
	poolSize := 2 * runtime.GOMAXPROCS(0)
	conf = &Config{
		Config: chpool.Config{
			PoolSize:        poolSize,
			PoolTimeout:     10 * time.Second,
			MaxIdleConns:    poolSize,
			ConnMaxIdleTime: 30 * time.Minute,
		},

		Compression: true,

		Addr:     "localhost:9000",
		User:     "default",
		Database: "default",

		DialTimeout:  5 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 10 * time.Second,

		MaxRetries:      3,
		MinRetryBackoff: time.Millisecond,
		MaxRetryBackoff: 3 * time.Second,
	}
	return conf
}

type Option func(db *DB)

func WithDiscardUnknownColumns() Option {
	return func(db *DB) {
		db.flags.Set(discardUnknownColumnsFlag)
	}
}

// WithCompression enables/disables LZ4 compression.
func WithCompression(enabled bool) Option {
	return func(db *DB) {
		db.conf.Compression = enabled
	}
}

func WithAutoCreateDatabase(enabled bool) Option {
	return func(db *DB) {
		db.flags.Set(autoCreateDatabaseFlag)
	}
}

// WithAddr configures TCP host:port.
func WithAddr(addr string) Option {
	return func(db *DB) {
		db.conf.Addr = addr
	}
}

// WithTLSConfig configures TLS config for secure connections.
func WithTLSConfig(conf *tls.Config) Option {
	return func(db *DB) {
		db.conf.TLSConfig = conf
	}
}

func WithQuerySettings(params map[string]any) Option {
	return func(db *DB) {
		db.conf.QuerySettings = params
	}
}

func WithInsecure(on bool) Option {
	return func(db *DB) {
		if on {
			db.conf.TLSConfig = nil
		} else {
			db.conf.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}
}

func WithUser(user string) Option {
	return func(db *DB) {
		db.conf.User = user
	}
}

func WithPassword(password string) Option {
	return func(db *DB) {
		db.conf.Password = password
	}
}

func WithDatabase(database string) Option {
	return func(db *DB) {
		db.conf.Database = database
	}
}

func WithCluster(cluster string) Option {
	return func(db *DB) {
		db.conf.Cluster = cluster
	}
}

// WithDialTimeout configures dial timeout for establishing new connections.
// Default is 5 seconds.
func WithDialTimeout(timeout time.Duration) Option {
	return func(db *DB) {
		db.conf.DialTimeout = timeout
	}
}

// WithReadTimeout configures timeout for socket reads. If reached, commands will fail
// with a timeout instead of blocking.
func WithReadTimeout(timeout time.Duration) Option {
	return func(db *DB) {
		db.conf.ReadTimeout = timeout
	}
}

// WithWriteTimeout configures timeout for socket writes. If reached, commands will fail
// with a timeout instead of blocking.
func WithWriteTimeout(timeout time.Duration) Option {
	return func(db *DB) {
		db.conf.WriteTimeout = timeout
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(db *DB) {
		db.conf.DialTimeout = timeout
		db.conf.ReadTimeout = timeout
		db.conf.WriteTimeout = timeout
	}
}

// WithMaxRetries configures maximum number of retries before giving up.
// Default is to retry query 2 times.
func WithMaxRetries(maxRetries int) Option {
	return func(db *DB) {
		db.conf.MaxRetries = maxRetries
	}
}

// WithMinRetryBackoff configures minimum backoff between each retry.
// Default is 250 milliseconds; -1 disables backoff.
func WithMinRetryBackoff(backoff time.Duration) Option {
	return func(db *DB) {
		db.conf.MinRetryBackoff = backoff
	}
}

// WithMaxRetryBackoff configures maximum backoff between each retry.
// Default is 4 seconds; -1 disables backoff.
func WithMaxRetryBackoff(backoff time.Duration) Option {
	return func(db *DB) {
		db.conf.MaxRetryBackoff = backoff
	}
}

// WithPoolSize configures maximum number of socket connections.
// Default is 2 connections per every CPU as reported by runtime.NumCPU.
func WithPoolSize(poolSize int) Option {
	return func(db *DB) {
		db.conf.PoolSize = poolSize
		db.conf.MaxIdleConns = poolSize
	}
}

// WithConnMaxLifetime sets the maximum amount of time a connection may be reused.
// Expired connections may be closed lazily before reuse.

// If d <= 0, connections are not closed due to a connection's age.
func WithConnMaxLifetime(d time.Duration) Option {
	return func(db *DB) {
		db.conf.ConnMaxLifetime = d
	}
}

// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle.
// Expired connections may be closed lazily before reuse.
//
// If d <= 0, connections are not closed due to a connection's idle time.
//
// ClickHouse closes idle connections after 1 hour (see idle_connection_timeout).
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(db *DB) {
		db.conf.ConnMaxIdleTime = d
	}
}

// WithPoolTimeout configures time for which client waits for free connection if all
// connections are busy before opening a new connection.
// Default is 10 seconds.
func WithPoolTimeout(timeout time.Duration) Option {
	return func(db *DB) {
		db.conf.PoolTimeout = timeout
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
		// ok
	default:
		return nil, errors.New("ch: unknown scheme: " + u.Scheme)
	}

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
			params[k] = parseSettingValue(v)
		}
		opts = append(opts, WithQuerySettings(params))
	}

	return opts, nil
}

func parseSettingValue(s string) any {
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	return s
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
