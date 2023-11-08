package bunconf

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/wneessen/go-mail"
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

	conf := defaultConfig()

	configStr := expandEnv(string(configBytes))
	if err := yaml.Unmarshal([]byte(configStr), conf); err != nil {
		return nil, err
	}

	conf.Path = confPath
	conf.Service = service

	fixUpConfig(conf)
	if err := validateConfig(conf); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", conf.Path, err)
	}

	return conf, nil
}

// defaultConfig returns a minimal working Uptrace config.
func defaultConfig() *Config {
	conf := new(Config)

	conf.CH.Addr = "localhost:9000"
	conf.CH.User = "default"
	conf.CH.Database = "uptrace"
	conf.CH.MaxExecutionTime = 30 * time.Second

	conf.PG.Addr = "localhost:5432"
	conf.PG.User = "uptrace"
	conf.PG.Password = "uptrace"
	conf.PG.Database = "uptrace"

	conf.Projects = []Project{
		{
			ID:    1,
			Name:  "Uptrace",
			Token: "project1_secret_token",
			PinnedAttrs: []string{
				attrkey.ServiceName,
				attrkey.HostName,
				attrkey.DeploymentEnvironment,
			},
		},
	}

	conf.CHSchema.Compression = "ZSTD(3)"
	conf.CHSchema.Spans.TTLDelete = "30 DAY"
	conf.CHSchema.Spans.StoragePolicy = "default"
	conf.CHSchema.Metrics.TTLDelete = "90 DAY"
	conf.CHSchema.Metrics.StoragePolicy = "default"

	conf.Listen.Scheme = "http"
	conf.Listen.GRPC.Addr = ":14317"
	conf.Listen.HTTP.Addr = ":14318"

	conf.SMTPMailer.Port = 25
	conf.SMTPMailer.From = "no-reply@localhost"
	conf.SMTPMailer.AuthType = mail.SMTPAuthPlain

	conf.Logging.Level = "INFO"

	return conf
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

func fixUpConfig(conf *Config) {
	for i := range conf.MetricsFromSpans {
		metric := &conf.MetricsFromSpans[i]
		metric.Value = cleanAttrName(metric.Value)
		for i, attr := range metric.Attrs {
			metric.Attrs[i] = cleanAttrName(attr)
		}
		for i, attr := range metric.Annotations {
			metric.Annotations[i] = cleanAttrName(attr)
		}
		metric.Where = cleanAttrName(metric.Where)
	}
}

func cleanAttrName(attrKey string) string {
	if strings.HasPrefix(attrKey, "span.") {
		return strings.TrimPrefix(attrKey, "span")
	}
	return attrKey
}

func validateConfig(conf *Config) error {
	if conf.CHSchema.Replicated && conf.CHSchema.Cluster == "" {
		return errors.New("ch_schema.cluster can't be empty when replicated=true")
	}

	if err := validateUsers(conf.Auth.Users); err != nil {
		return err
	}
	if err := validateProjects(conf.Projects); err != nil {
		return err
	}

	if conf.Listen.TLS == nil {
		if conf.Listen.HTTP.TLS != nil {
			conf.Listen.TLS = conf.Listen.HTTP.TLS
		} else if conf.Listen.GRPC.TLS != nil {
			conf.Listen.TLS = conf.Listen.GRPC.TLS
		}
	}

	if conf.Listen.TLS != nil {
		tlsConf, err := conf.Listen.TLS.TLSConfig()
		if err != nil {
			return err
		}
		if tlsConf != nil {
			conf.Listen.Scheme = "https"
		}
	}

	if err := conf.Listen.GRPC.init(); err != nil {
		return fmt.Errorf("invalid listen.grpc option: %w", err)
	}
	if err := conf.Listen.HTTP.init(); err != nil {
		return fmt.Errorf("invalid listen.grpc option: %w", err)
	}

	if err := conf.initSite(); err != nil {
		return err
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

	if !conf.ServiceGraph.Disabled {
		store := &conf.ServiceGraph.Store
		if store.Size == 0 {
			store.Size = ScaleWithCPU(100000, 10000000)
		}
		if store.TTL == 0 {
			store.TTL = 5 * time.Second
		}
	}

	return nil
}

func (conf *Config) initSite() error {
	if conf.Site.Addr == "" {
		conf.Site.Addr = fmt.Sprintf("%s://%s:%s",
			conf.Listen.Scheme, conf.Listen.HTTP.Host, conf.Listen.HTTP.Port)
	}

	siteURL, err := url.Parse(conf.Site.Addr)
	if err != nil {
		return fmt.Errorf("invalid site.addr option: %w", err)
	}
	conf.Site.URL = siteURL

	conf.Site.Host, _, err = net.SplitHostPort(conf.Site.URL.Host)
	if err != nil {
		conf.Site.Host = conf.Site.URL.Host
	}

	if !strings.HasSuffix(conf.Site.URL.Path, "/") {
		conf.Site.URL.Path += "/"
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
		if seen[user.Email] {
			return fmt.Errorf("user with username=%q already exists", user.Email)
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
	CH CHConfig  `yaml:"ch"`
	PG BunConfig `yaml:"pg"`

	Projects []Project `yaml:"projects"`

	Auth struct {
		Users      []User                `yaml:"users" json:"users"`
		Cloudflare []*CloudflareProvider `yaml:"cloudflare" json:"cloudflare"`
		OIDC       []*OIDCProvider       `yaml:"oidc" json:"oidc"`
	} `yaml:"auth" json:"auth"`

	MetricsFromSpans []SpanMetric `yaml:"metrics_from_spans"`

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
	} `yaml:"ch_schema"`

	Listen struct {
		HTTP Listen `yaml:"http"`
		GRPC Listen `yaml:"grpc"`

		TLS    *TLSServer `yaml:"tls"`
		Scheme string     `yaml:"-"`
	} `yaml:"listen"`

	Site struct {
		Addr string `yaml:"addr"`

		URL  *url.URL `yaml:"-"`
		Host string   `yaml:"-"`
	} `yaml:"site"`

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

	ServiceGraph struct {
		Disabled bool `yaml:"disabled"`
		Store    struct {
			Size int           `yaml:"size"`
			TTL  time.Duration `yaml:"ttl"`
		} `yaml:"store"`
	}

	SMTPMailer struct {
		Enabled  bool              `yaml:"enabled"`
		Host     string            `yaml:"host"`
		Port     int               `yaml:"port"`
		AuthType mail.SMTPAuthType `yaml:"auth_type"`
		Username string            `yaml:"username"`
		Password string            `yaml:"password"`

		From string `yaml:"from"`
	} `yaml:"smtp_mailer"`

	UptraceGo struct {
		Disabled bool       `yaml:"disabled"`
		DSN      string     `yaml:"dsn"`
		TLS      *TLSClient `yaml:"tls"`
	} `yaml:"uptrace_go"`

	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`

	SecretKey string `yaml:"secret_key"`
	Debug     bool   `yaml:"debug"`

	Path    string `yaml:"-"`
	Service string `yaml:"-"`
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

func (m *SpanMetric) ViewName() string {
	return "metrics_" + strings.ReplaceAll(m.Name, ".", "_") + "_mv"
}

type Listen struct {
	Addr string `yaml:"addr"`
	Host string `yaml:"-"`
	Port string `yaml:"-"`

	// DEPRECATED
	TLS *TLSServer `yaml:"tls"`
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

	return nil
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
