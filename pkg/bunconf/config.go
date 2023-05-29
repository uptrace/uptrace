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
	conf.Auth.Users = []User{
		{
			Name:          "John Doe",
			Email:         "uptrace@localhost",
			Password:      "uptrace",
			NotifyByEmail: true,
		},
	}

	conf.CHSchema.Compression = "ZSTD(3)"
	conf.CHSchema.Spans.TTLDelete = "30 DAY"
	conf.CHSchema.Spans.StoragePolicy = "default"
	conf.CHSchema.Metrics.TTLDelete = "90 DAY"
	conf.CHSchema.Metrics.StoragePolicy = "default"

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

func validateConfig(conf *Config) error {
	if conf.CHSchema.Cluster != "" {
		if !conf.CHSchema.Replicated {
			conf.CHSchema.Replicated = true
		}
	}

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

	return nil
}

func (conf *Config) initSite() error {
	if conf.Site.Addr == "" {
		conf.Site.Addr = fmt.Sprintf("%s://%s:%s",
			conf.Listen.HTTP.Scheme, conf.Listen.HTTP.Host, conf.Listen.HTTP.Port)
	}

	siteURL, err := url.Parse(conf.Site.Addr)
	if err != nil {
		return fmt.Errorf("invalid site.addr option: %w", err)
	}
	conf.Site.URL = siteURL

	conf.Site.Host, conf.Site.Port, err = net.SplitHostPort(conf.Site.URL.Host)
	if err != nil {
		return err
	}

	if conf.Site.Path != "" {
		conf.Site.URL.Path = conf.Site.Path
	} else {
		conf.Site.URL.Path = "/"
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

		Tables struct {
			SpansData  CHTableOverride `yaml:"spans_data"`
			SpansIndex CHTableOverride `yaml:"spans_index"`
		}
	} `yaml:"ch_schema"`

	Listen struct {
		HTTP Listen `yaml:"http"`
		GRPC Listen `yaml:"grpc"`
	} `yaml:"listen"`

	Site struct {
		Addr string `yaml:"addr"`
		Path string `yaml:"path"` // DEPRECATED

		URL  *url.URL `yaml:"-"`
		Host string   `yaml:"-"`
		Port string   `yaml:"-"`
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

	SMTPMailer struct {
		Enabled  bool              `json:"enabled"`
		Host     string            `yaml:"host"`
		Port     int               `yaml:"port"`
		AuthType mail.SMTPAuthType `yaml:"auth_type"`
		Username string            `yaml:"username"`
		Password string            `yaml:"password"`

		From string `yaml:"from"`
	} `yaml:"smtp_mailer"`

	UptraceGo struct {
		DSN string     `yaml:"dsn"`
		TLS *TLSClient `yaml:"tls"`
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
	return fmt.Sprintf("%s://%s:%s",
		c.Listen.GRPC.Scheme,
		c.Site.Host,
		c.Listen.GRPC.Port)
}

func (c *Config) HTTPEndpoint() string {
	return fmt.Sprintf("%s://%s:%s",
		c.Listen.HTTP.Scheme,
		c.Site.Host,
		c.Listen.HTTP.Port)
}

func (c *Config) GRPCDsn(projectID uint32, projectToken string) string {
	return fmt.Sprintf("%s://%s@%s:%s/%d",
		c.Listen.GRPC.Scheme,
		projectToken,
		c.Site.Host,
		c.Listen.GRPC.Port,
		projectID)
}

func (c *Config) HTTPDsn(projectID uint32, projectToken string) string {
	return fmt.Sprintf("%s://%s@%s:%s/%d",
		c.Listen.HTTP.Scheme,
		projectToken,
		c.Site.Host,
		c.Listen.HTTP.Port,
		projectID)
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
