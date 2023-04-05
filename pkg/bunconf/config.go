package bunconf

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

func ReadConfig(confPath, service string) (*Config, error) {
	confPath, err := filepath.Abs(confPath)
	if err != nil {
		return nil, err
	}

	configBytes, err := os.ReadFile(confPath)
	if err != nil {
		return nil, err
	}

	conf := new(Config)

	configStr := expandEnv(string(configBytes))
	if err := yaml.Unmarshal([]byte(configStr), conf); err != nil {
		return nil, err
	}

	conf.Path = confPath
	conf.Service = service

	if err := validateConfig(conf); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", conf.Path, err)
	}

	return conf, nil
}

var (
	envVarRe        = regexp.MustCompile(`^[A-Z][A-Z0-9_]+$`)
	envVarDefaultRe = regexp.MustCompile(`^([A-Z][A-Z0-9_]+):(.*)$`)
)

func expandEnv(conf string) string {
	return os.Expand(conf, func(envVar string) string {
		if envVar == "$" { // escaping
			return "$"
		}

		if envVarRe.MatchString(envVar) {
			return os.Getenv(envVar)
		}

		if m := envVarDefaultRe.FindStringSubmatch(envVar); m != nil {
			envVar = m[1]
			defValue := m[2]
			if found, ok := os.LookupEnv(envVar); ok {
				return found
			}
			return defValue
		}

		return "$" + envVar
	})
}

func validateConfig(conf *Config) error {
	if err := validateUsers(conf.Auth.Users); err != nil {
		return err
	}
	if err := validateProjects(conf.Projects); err != nil {
		return err
	}

	if err := conf.Listen.GRPC.init(); err != nil {
		return fmt.Errorf("invalid listen.grpc option: %w", err)
	}
	if err := conf.Listen.HTTP.init(); err != nil {
		return fmt.Errorf("invalid listen.grpc option: %w", err)
	}

	if conf.Site.Addr == "" {
		conf.Site.Addr = conf.HTTPEndpoint()
	}
	if _, err := url.Parse(conf.Site.Addr); err != nil {
		return fmt.Errorf("invalid site.addr option: %w", err)
	}
	if conf.Site.Path == "" {
		conf.Site.Path = "/"
	}

	if conf.Spans.BatchSize == 0 {
		conf.Spans.BatchSize = ScaleWithCPU(1000, 32000)
	}
	if conf.Spans.BufferSize == 0 {
		conf.Spans.BufferSize = runtime.GOMAXPROCS(0) * conf.Spans.BatchSize
	}

	if conf.Metrics.BatchSize == 0 {
		conf.Metrics.BatchSize = ScaleWithCPU(1000, 32000)
	}
	if conf.Metrics.BufferSize == 0 {
		conf.Metrics.BufferSize = runtime.GOMAXPROCS(0) * conf.Spans.BatchSize
	}
	if conf.Metrics.CumToDeltaSize == 0 {
		conf.Metrics.CumToDeltaSize = ScaleWithCPU(10000, 500000)
	}

	return nil
}

func validateUsers(users []User) error {
	if len(users) == 0 {
		return nil
	}

	seen := make(map[string]bool, len(users))
	for i := range users {
		user := &users[i]
		if seen[user.Username] {
			return fmt.Errorf("user with username=%q already exists", user.Username)
		}
	}

	return nil
}

func validateProjects(projects []Project) error {
	if len(projects) == 0 {
		return fmt.Errorf("config must contain at least one project")
	}

	seen := make(map[string]bool, len(projects))
	for i := range projects {
		project := &projects[i]
		if seen[project.Token] {
			return fmt.Errorf("project %d has a duplicated token %q", project.ID, project.Token)
		}
	}

	return nil
}

type Config struct {
	Path    string `yaml:"-"`
	Service string `yaml:"-"`

	Debug     bool   `yaml:"debug"`
	SecretKey string `yaml:"secret_key"`

	Logs struct {
		Level string `yaml:"level"`
	} `yaml:"logs"`

	Site struct {
		Addr string `yaml:"addr"`
		Path string `yaml:"path"`
	} `yaml:"site"`

	Listen struct {
		HTTP Listen `yaml:"http"`
		GRPC Listen `yaml:"grpc"`
	} `yaml:"listen"`

	PG BunConfig `yaml:"pg"`
	CH CHConfig  `yaml:"ch"`

	CHSchema struct {
		Compression string `yaml:"compression"`
		Replicated  bool   `yaml:"replicated"`
		Cluster     string `yaml:"cluster"`

		Spans struct {
			StoragePolicy string `yaml:"storage_policy"`
			TTLDelete     string `yaml:"ttl_delete"`
		} `yaml:"spans"`

		Metrics struct {
			StoragePolicy string `yaml:"storage_policy"`
			TTLDelete     string `yaml:"ttl_delete"`
		} `yaml:"metrics"`

		Tables struct {
			SpansData  CHTableOverride `yaml:"spans_data"`
			SpansIndex CHTableOverride `yaml:"spans_index"`
		}
	} `yaml:"ch_schema"`

	Spans struct {
		BufferSize int `yaml:"buffer_size"`
		BatchSize  int `yaml:"batch_size"`
	} `yaml:"spans"`

	Metrics struct {
		DropAttrs []string `yaml:"drop_attrs"`

		BufferSize     int `yaml:"buffer_size"`
		BatchSize      int `yaml:"batch_size"`
		CumToDeltaSize int `yaml:"cum_to_delta_size"`
	} `yaml:"metrics"`

	MetricsFromSpans []SpanMetric `yaml:"metrics_from_spans"`

	Auth struct {
		Users      []User                `yaml:"users" json:"users"`
		Cloudflare []*CloudflareProvider `yaml:"cloudflare" json:"cloudflare"`
		OIDC       []*OIDCProvider       `yaml:"oidc" json:"oidc"`
	} `yaml:"auth" json:"auth"`

	Projects []Project `yaml:"projects"`

	UptraceGo struct {
		DSN string     `yaml:"dsn"`
		TLS *TLSClient `yaml:"tls"`
	} `yaml:"uptrace_go"`
}

type SpanMetric struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Instrument  string   `yaml:"instrument"`
	Unit        string   `yaml:"unit"`
	Value       string   `yaml:"value"`
	Attrs       []string `yaml:"attrs"`
	Annotations []string `yaml:"annotations"`
	Where       string   `yaml:"where"`
}

type Listen struct {
	Addr string     `yaml:"addr"`
	TLS  *TLSServer `yaml:"tls"`

	Scheme string `yaml:"-"`
	Host   string `yaml:"-"`
	Port   string `yaml:"-"`
}

func (l *Listen) init() error {
	host, port, err := net.SplitHostPort(l.Addr)
	if err != nil {
		return err
	}

	l.Host = host
	if l.Host == "" {
		l.Host = "localhost"
	}
	l.Port = port

	l.Scheme = "http"
	if l.TLS != nil {
		tlsConf, err := l.TLS.TLSConfig()
		if err != nil {
			return err
		}
		if tlsConf != nil {
			l.Scheme = "https"
		}
	}

	return nil
}

type CHTableOverride struct {
	TTL string `yaml:"ttl"`
}

func (c *Config) GRPCEndpoint() string {
	return fmt.Sprintf("%s://%s:%s", c.Listen.GRPC.Scheme, c.Listen.GRPC.Host, c.Listen.GRPC.Port)
}

func (c *Config) HTTPEndpoint() string {
	return fmt.Sprintf("%s://%s:%s", c.Listen.HTTP.Scheme, c.Listen.HTTP.Host, c.Listen.HTTP.Port)
}

func (c *Config) GRPCDsn(projectID uint32, projectToken string) string {
	return fmt.Sprintf("%s://%s@%s:%s/%d",
		c.Listen.GRPC.Scheme, projectToken, c.Listen.GRPC.Host, c.Listen.GRPC.Port, projectID)
}

func (c *Config) HTTPDsn(projectID uint32, projectToken string) string {
	return fmt.Sprintf("%s://%s@%s:%s/%d",
		c.Listen.HTTP.Scheme, projectToken, c.Listen.HTTP.Host, c.Listen.HTTP.Port, projectID)
}

func (c *Config) SiteURL(sitePath string, args ...any) string {
	u, err := url.Parse(c.Site.Addr)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(u.Path, fmt.Sprintf(sitePath, args...))
	return u.String()
}

type BunConfig struct {
	DSN string `yaml:"dsn"`

	Addr     string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`

	TLS *TLSClient `yaml:"tls"`

	ConnParams map[string]any `yaml:"conn_params"`
}

type CHConfig struct {
	DSN string `yaml:"dsn"`

	Addr          string         `yaml:"addr"`
	User          string         `yaml:"user"`
	Password      string         `yaml:"password"`
	Database      string         `yaml:"database"`
	QuerySettings map[string]any `yaml:"query_settings"`

	TLS *TLSClient `yaml:"tls"`

	MaxExecutionTime time.Duration `yaml:"max_execution_time"`
}

func ScaleWithCPU(min, max int) int {
	if min == 0 {
		panic("min == 0")
	}
	if max == 0 {
		panic("max == 0")
	}

	n := runtime.GOMAXPROCS(0) * min
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
