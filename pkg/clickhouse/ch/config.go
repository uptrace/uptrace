package ch

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/uptrace/pkg/clickhouse/ch/chpool"
	"github.com/uptrace/pkg/clickhouse/ch/internal"
	"net"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	discardUnknownColumnsFlag internal.Flag = 1 << iota
)

type Config struct {
	chpool.Config
	Network         string
	Addr            string
	User            string
	Password        string
	Database        string
	TLSConfig       *tls.Config
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
	Compression     bool
	IsBackup        bool
	QuerySettings   map[string]any
	Cluster         string
	Replicated      bool
	Distributed     bool
	NamedArgs       map[string]any
}

func (conf *Config) netDialer() *net.Dialer { return &net.Dialer{Timeout: conf.DialTimeout} }
func defaultConfig() *Config {
	var conf *Config
	poolSize := 2 * runtime.GOMAXPROCS(0)
	conf = &Config{Config: chpool.Config{PoolSize: poolSize, PoolTimeout: 10 * time.Second, MaxIdleConns: poolSize, ConnMaxIdleTime: 30 * time.Minute}, Compression: true, Network: "tcp", Addr: "localhost:9000", User: "default", Database: "default", DialTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 5 * time.Second, MaxRetries: 3, MinRetryBackoff: time.Second, MaxRetryBackoff: 3 * time.Second}
	return conf
}

type Option func(db *DB)

func WithDiscardUnknownColumns() Option {
	return func(db *DB) { db.flags.Set(discardUnknownColumnsFlag) }
}
func WithCompression(enabled bool) Option   { return func(db *DB) { db.conf.Compression = enabled } }
func WithNetwork(network string) Option     { return func(db *DB) { db.conf.Network = network } }
func WithAddr(addr string) Option           { return func(db *DB) { db.conf.Addr = addr } }
func WithTLSConfig(conf *tls.Config) Option { return func(db *DB) { db.conf.TLSConfig = conf } }
func WithIsBackup(on bool) Option           { return func(db *DB) { db.conf.IsBackup = on } }
func WithQuerySettings(params map[string]any) Option {
	return func(db *DB) { db.conf.QuerySettings = params }
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
func WithUser(user string) Option         { return func(db *DB) { db.conf.User = user } }
func WithPassword(password string) Option { return func(db *DB) { db.conf.Password = password } }
func WithDatabase(database string) Option { return func(db *DB) { db.conf.Database = database } }
func WithDialTimeout(timeout time.Duration) Option {
	return func(db *DB) { db.conf.DialTimeout = timeout }
}
func WithReadTimeout(timeout time.Duration) Option {
	return func(db *DB) { db.conf.ReadTimeout = timeout }
}
func WithWriteTimeout(timeout time.Duration) Option {
	return func(db *DB) { db.conf.WriteTimeout = timeout }
}
func WithTimeout(timeout time.Duration) Option {
	return func(db *DB) {
		db.conf.DialTimeout = timeout
		db.conf.ReadTimeout = timeout
		db.conf.WriteTimeout = timeout
	}
}
func WithMaxRetries(maxRetries int) Option { return func(db *DB) { db.conf.MaxRetries = maxRetries } }
func WithMinRetryBackoff(backoff time.Duration) Option {
	return func(db *DB) { db.conf.MinRetryBackoff = backoff }
}
func WithMaxRetryBackoff(backoff time.Duration) Option {
	return func(db *DB) { db.conf.MaxRetryBackoff = backoff }
}
func WithPoolSize(poolSize int) Option {
	return func(db *DB) { db.conf.PoolSize = poolSize; db.conf.MaxIdleConns = poolSize }
}
func WithConnMaxLifetime(d time.Duration) Option { return func(db *DB) { db.conf.ConnMaxLifetime = d } }
func WithConnMaxIdleTime(d time.Duration) Option { return func(db *DB) { db.conf.ConnMaxIdleTime = d } }
func WithPoolTimeout(timeout time.Duration) Option {
	return func(db *DB) { db.conf.PoolTimeout = timeout }
}
func WithNamedArg(key string, value any) Option {
	return func(db *DB) {
		if db.conf.NamedArgs == nil {
			db.conf.NamedArgs = make(map[string]any)
		}
		db.conf.NamedArgs[key] = value
	}
}
func WithCluster(cluster string) Option     { return func(db *DB) { db.conf.Cluster = cluster } }
func WithReplicated(replicated bool) Option { return func(db *DB) { db.conf.Replicated = replicated } }
func WithDistributed(distributed bool) Option {
	return func(db *DB) { db.conf.Distributed = distributed }
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
		return nil, errors.New("ch: invalid scheme: " + u.Scheme)
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
	delete(o.q, name)
	return vs[len(vs)-1]
}
func (o *queryOptions) duration(name string) time.Duration {
	s := o.string(name)
	if s == "" {
		return 0
	}
	if i, err := strconv.Atoi(s); err == nil {
		if i <= 0 {
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
