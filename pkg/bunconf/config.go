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

var envVarRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]+$`)

func expandEnv(conf string) string {
	return os.Expand(conf, func(envVar string) string {
		if envVar == "$" { // escaping
			return "$"
		}
		if !envVarRe.MatchString(envVar) {
			return "$" + envVar
		}
		return os.Getenv(envVar)
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

	Users    []User    `yaml:"users"`
	Projects []Project `yaml:"projects"`

	Alerting struct {
		Rules []AlertRule `yaml:"rules"`

		CreateAlertsFromSpans struct {
			Enabled bool              `yaml:"enabled"`
			Labels  map[string]string `yaml:"labels"`
		} `yaml:"create_alerts_from_spans"`
	} `yaml:"alerting"`

	Alertmanager struct {
		URLs []string `yaml:"urls"`
	} `yaml:"alertmanager"`
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
	ID       uint64 `yaml:"id" json:"id"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"-"`
}

type Project struct {
	ID          uint32   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	Token       string   `yaml:"token" json:"token"`
	PinnedAttrs []string `yaml:"pinned_attrs" json:"pinnedAttrs"`
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
	DSN string `yaml:"dsn"`
}

type CHConfig struct {
	DSN string `yaml:"dsn"`

	Addr     string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
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
