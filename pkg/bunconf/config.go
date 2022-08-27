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

	if _, err := url.Parse(conf.Site.Addr); err != nil {
		return fmt.Errorf(`can't parse "site.addr": %w`, err)
	}

	httpHost, httpPort, err := splitHostPort(conf.Listen.HTTP)
	if err != nil {
		return fmt.Errorf(`can't parse "listen.http": %w`, err)
	}
	conf.Listen.HTTPHost = httpHost
	conf.Listen.HTTPPort = httpPort

	grpcHost, grpcPort, err := splitHostPort(conf.Listen.GRPC)
	if err != nil {
		return fmt.Errorf(`can't parse "listen.grpc": %w`, err)
	}
	conf.Listen.GRPCHost = grpcHost
	conf.Listen.GRPCPort = grpcPort

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

	if conf.DB.DSN == "" {
		return fmt.Errorf(`db.dsn option is required`)
	}

	return nil
}

func splitHostPort(addr string) (string, string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", "", err
	}
	if host == "" {
		host = "localhost"
	}
	return host, port, nil
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
		HTTP string `yaml:"http"`
		GRPC string `yaml:"grpc"`

		// Calculated.

		HTTPHost string `yaml:"-"`
		HTTPPort string `yaml:"-"`

		GRPCHost string `yaml:"-"`
		GRPCPort string `yaml:"-"`
	} `yaml:"listen"`

	DB BunConfig `yaml:"db"`
	CH CHConfig  `yaml:"ch"`

	CHSchema struct {
		Compression string `yaml:"compression"`
		Cluster     string `yaml:"cluster"`
		Replicated  bool   `yaml:"replicated"`
		Distributed bool   `yaml:"distributed"`
	} `yaml:"ch_schema"`

	Spans struct {
		TTL string `yaml:"ttl"`

		BufferSize int `yaml:"buffer_size"`
		BatchSize  int `yaml:"batch_size"`
	} `yaml:"spans"`

	Metrics struct {
		TTL string `yaml:"ttl"`

		DropAttrs []string `yaml:"drop_attrs"`

		BufferSize int `yaml:"buffer_size"`
		BatchSize  int `yaml:"batch_size"`
	} `yaml:"metrics"`

	Users    []User    `yaml:"users"`
	Projects []Project `yaml:"projects"`

	Alerting struct {
		Rules []AlertRule `yaml:"rules"`

		AlertsFromErrors struct {
			Enabled bool              `yaml:"enabled"`
			Labels  map[string]string `yaml:"labels"`
		} `yaml:"alerts_from_errors"`
	} `yaml:"alerting"`

	Alertmanager struct {
		URLs []string `yaml:"urls"`
	} `yaml:"alertmanager"`

	Dashboards []string `yaml:"dashboards"`
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

func (c *Config) Scheme() string {
	return "http"
}

func (c *Config) GRPCEndpoint(project *Project) string {
	return fmt.Sprintf("%s://%s:%s", c.Scheme(), c.Listen.GRPCHost, c.Listen.GRPCPort)
}

func (c *Config) HTTPEndpoint(project *Project) string {
	return fmt.Sprintf("%s://%s:%s", c.Scheme(), c.Listen.HTTPHost, c.Listen.HTTPPort)
}

func (c *Config) GRPCDsn(project *Project) string {
	return fmt.Sprintf("%s://%s@%s:%s/%d",
		c.Scheme(), project.Token, c.Listen.GRPCHost, c.Listen.GRPCPort, project.ID)
}

func (c *Config) HTTPDsn(project *Project) string {
	return fmt.Sprintf("%s://%s@%s:%s/%d",
		c.Scheme(), project.Token, c.Listen.HTTPHost, c.Listen.HTTPPort, project.ID)
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
