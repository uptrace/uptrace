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

	"github.com/cespare/xxhash/v2"
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

	if conf.Spans.BatchSize == 0 {
		conf.Spans.BatchSize = ScaleWithCPU(1000, 32000)
	}
	if conf.Spans.BufferSize == 0 {
		conf.Spans.BufferSize = 2 * runtime.GOMAXPROCS(0) * conf.Spans.BatchSize
	}

	if conf.Metrics.BatchSize == 0 {
		conf.Metrics.BatchSize = ScaleWithCPU(1000, 32000)
	}
	if conf.Metrics.BufferSize == 0 {
		conf.Metrics.BufferSize = 2 * runtime.GOMAXPROCS(0) * conf.Spans.BatchSize
	}

	if conf.DB.DSN == "" {
		return fmt.Errorf(`db.dsn option can not be empty`)
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
	} `yaml:"site"`

	Listen struct {
		HTTP Listen `yaml:"http"`
		GRPC Listen `yaml:"grpc"`
	} `yaml:"listen"`

	DB BunConfig `yaml:"db"`
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

		BufferSize int `yaml:"buffer_size"`
		BatchSize  int `yaml:"batch_size"`
	} `yaml:"metrics"`

	MetricsFromSpans []SpanMetric `yaml:"metrics_from_spans"`

	Auth struct {
		Users      []User                `yaml:"users" json:"users"`
		Cloudflare []*CloudflareProvider `yaml:"cloudflare" json:"cloudflare"`
		OIDC       []*OIDCProvider       `yaml:"oidc" json:"oidc"`
	} `yaml:"auth" json:"auth"`

	Projects []Project `yaml:"projects"`

	Alerting struct {
		Rules []AlertRule `yaml:"rules"`

		CreateAlertsFromSpans struct {
			Enabled bool              `yaml:"enabled"`
			Labels  map[string]string `yaml:"labels"`
		} `yaml:"create_alerts_from_spans"`
	} `yaml:"alerting"`

	AlertmanagerClient struct {
		URLs []string `yaml:"urls"`
	} `yaml:"alertmanager_client"`

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

type User struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"-"`
}

func (u *User) Hash() uint64 {
	return xxhash.Sum64(append([]byte(u.Username), []byte(u.Password)...))
}

type CloudflareProvider struct {
	TeamURL  string `yaml:"team_url" json:"team_url"`
	Audience string `yaml:"audience" json:"audience"`
}

type OIDCProvider struct {
	ID           string   `yaml:"id" json:"id"`
	DisplayName  string   `yaml:"display_name" json:"display_name"`
	IssuerURL    string   `yaml:"issuer_url" json:"issuer_url"`
	ClientID     string   `yaml:"client_id" json:"client_id"`
	ClientSecret string   `yaml:"client_secret" json:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url" json:"redirect_url"`
	Scopes       []string `yaml:"scopes" json:"scopes"`
	Claim        string   `yaml:"claim" json:"claim"`
}

type Project struct {
	ID                  uint32   `yaml:"id" json:"id"`
	Name                string   `yaml:"name" json:"name"`
	Token               string   `yaml:"token" json:"token"`
	PinnedAttrs         []string `yaml:"pinned_attrs" json:"pinnedAttrs"`
	GroupByEnv          bool     `yaml:"group_by_env" json:"groupByEnv"`
	GroupFuncsByService bool     `yaml:"group_funcs_by_service" json:"groupFuncsByService"`
}

func (c *Config) GRPCEndpoint() string {
	return fmt.Sprintf("%s://%s:%s", c.Listen.GRPC.Scheme, c.Listen.GRPC.Host, c.Listen.GRPC.Port)
}

func (c *Config) HTTPEndpoint() string {
	return fmt.Sprintf("%s://%s:%s", c.Listen.HTTP.Scheme, c.Listen.HTTP.Host, c.Listen.HTTP.Port)
}

func (c *Config) GRPCDsn(project *Project) string {
	return fmt.Sprintf("%s://%s@%s:%s/%d",
		c.Listen.GRPC.Scheme, project.Token, c.Listen.GRPC.Host, c.Listen.GRPC.Port, project.ID)
}

func (c *Config) HTTPDsn(project *Project) string {
	return fmt.Sprintf("%s://%s@%s:%s/%d",
		c.Listen.HTTP.Scheme, project.Token, c.Listen.HTTP.Host, c.Listen.HTTP.Port, project.ID)
}

func (c *Config) SitePath(sitePath string) string {
	u, err := url.Parse(c.Site.Addr)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(u.Path, sitePath)
	return u.String()
}

type BunConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
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
