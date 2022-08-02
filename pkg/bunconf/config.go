package bunconf

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

func ReadConfig(configFile, service string) (*Config, error) {
	configFile, err := filepath.Abs(configFile)
	if err != nil {
		return nil, err
	}

	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	conf := new(Config)

	configStr := expandEnv(string(configBytes))
	if err := yaml.Unmarshal([]byte(configStr), conf); err != nil {
		return nil, err
	}

	conf.BaseDir = filepath.Dir(configFile)
	conf.FileName = filepath.Base(configFile)
	conf.Service = service

	if err := validateConfig(conf); err != nil {
		return nil, fmt.Errorf("config is invalid: %w", err)
	}

	return conf, nil
}

func expandEnv(s string) string {
	return os.Expand(s, func(str string) string {
		if str == "$" { // escaping
			return "$"
		}
		return os.Getenv(str)
	})
}

func validateConfig(conf *Config) error {
	if err := validateProjects(conf.Projects); err != nil {
		return err
	}

	if _, err := url.Parse(conf.Site.Addr); err != nil {
		return fmt.Errorf(`can't parse "site.addr": %w`, err)
	}

	httpHost, httpPort, err := net.SplitHostPort(conf.Listen.HTTP)
	if err != nil {
		return fmt.Errorf(`can't parse "listen.http": %w`, err)
	}
	conf.Listen.HTTPHost = httpHost
	conf.Listen.HTTPPort = httpPort

	grpcHost, grpcPort, err := net.SplitHostPort(conf.Listen.GRPC)
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

	if conf.Prometheus.Config == "" {
		return fmt.Errorf(`"prometheus.config" must contain a valid path`)
	}

	if _, err := url.Parse(conf.Prometheus.ExternalURL); err != nil {
		return err
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
	BaseDir  string `yaml:"-"`
	FileName string `yaml:"-"`
	Service  string `yaml:"-"`

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
		TTL         string `yaml:"ttl"`
		Compression string `yaml:"compression"`
		Cluster     string `yaml:"cluster"`
		Replicated  bool   `yaml:"replicated"`
		Distributed bool   `yaml:"distributed"`
	} `yaml:"ch_schema"`

	Spans struct {
		BufferSize int `yaml:"buffer_size"`
		BatchSize  int `yaml:"batch_size"`
	} `yaml:"spans"`

	Metrics struct {
		BufferSize int `yaml:"buffer_size"`
		BatchSize  int `yaml:"batch_size"`
	} `yaml:"metrics"`

	Users    []User    `yaml:"users"`
	Projects []Project `yaml:"projects"`

	Prometheus struct {
		Config      string `yaml:"config"`
		ExternalURL string `yaml:"externalUrl"`
	} `yaml:"prometheus"`

	Loki struct {
		Addr string `yaml:"addr"`
	} `yaml:"loki"`
}

type User struct {
	ID       uint64 `yaml:"id" json:"id"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"-"`
}

type Project struct {
	ID    uint32 `yaml:"id" json:"id"`
	Name  string `yaml:"name" json:"name"`
	Token string `yaml:"token" json:"token"`
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
