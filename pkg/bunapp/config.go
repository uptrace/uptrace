package bunapp

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

func ReadConfig(configFile, service string) (*AppConfig, error) {
	configFile, err := filepath.Abs(configFile)
	if err != nil {
		return nil, err
	}

	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	cfg := new(AppConfig)

	configStr := expandEnv(string(configBytes))
	if err := yaml.Unmarshal([]byte(configStr), cfg); err != nil {
		return nil, err
	}

	cfg.Filepath = configFile
	cfg.Service = service

	if err := validateProjects(cfg.Projects); err != nil {
		return nil, err
	}

	httpHost, httpPort, err := net.SplitHostPort(cfg.Listen.HTTP)
	if err != nil {
		return nil, fmt.Errorf("can't parse option listen.http addr: %w", err)
	}
	if httpHost == "" {
		httpHost = cfg.Site.Host
	}
	cfg.Listen.HTTPHost = httpHost
	cfg.Listen.HTTPPort = httpPort

	grpcHost, grpcPort, err := net.SplitHostPort(cfg.Listen.GRPC)
	if err != nil {
		return nil, fmt.Errorf("can't parse option listen.grpc addr: %w", err)
	}
	if grpcHost == "" {
		grpcHost = cfg.Site.Host
	}
	cfg.Listen.GRPCHost = grpcHost
	cfg.Listen.GRPCPort = grpcPort

	if cfg.Spans.BatchSize == 0 {
		cfg.Spans.BatchSize = scaleWithCPU(1000, 32000)
	}
	if cfg.Spans.BufferSize == 0 {
		cfg.Spans.BufferSize = runtime.GOMAXPROCS(0) * cfg.Spans.BatchSize
	}

	if cfg.Metrics.BatchSize == 0 {
		cfg.Metrics.BatchSize = scaleWithCPU(1000, 32000)
	}
	if cfg.Metrics.BufferSize == 0 {
		cfg.Metrics.BufferSize = runtime.GOMAXPROCS(0) * cfg.Spans.BatchSize
	}

	return cfg, nil
}

func expandEnv(s string) string {
	return os.Expand(s, func(str string) string {
		if str == "$" { // escaping
			return "$"
		}
		return os.Getenv(str)
	})
}

func validateProjects(projects []Project) error {
	if len(projects) == 0 {
		return fmt.Errorf("config must contain at least one project")
	}

	seen := make(map[string]bool, len(projects))
	for i := range projects {
		project := &projects[i]
		if seen[project.Token] {
			return fmt.Errorf("project %d has a duplicate token %q", project.ID, project.Token)
		}
	}

	return nil
}

type AppConfig struct {
	Filepath string `yaml:"-"`
	Service  string `yaml:"service"`

	Debug     bool   `yaml:"debug"`
	SecretKey string `yaml:"secret_key"`

	Site struct {
		Scheme string `yaml:"scheme"`
		Host   string `yaml:"host"`
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
		Replicated  bool   `yaml:"replicated"`
		Cluster     string `yaml:"cluster"`
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

func (c *AppConfig) SiteAddr() string {
	return fmt.Sprintf("%s://%s:%s/", c.Site.Scheme, c.Listen.HTTPHost, c.Listen.HTTPPort)
}

func (c *AppConfig) GRPCEndpoint(project *Project) string {
	return fmt.Sprintf("%s://%s:%s", c.Site.Scheme, c.Listen.GRPCHost, c.Listen.GRPCPort)
}

func (c *AppConfig) HTTPEndpoint(project *Project) string {
	return fmt.Sprintf("%s://%s:%s", c.Site.Scheme, c.Listen.HTTPHost, c.Listen.HTTPPort)
}

func (c *AppConfig) GRPCDsn(project *Project) string {
	return fmt.Sprintf("%s://%s@%s:%s/%d",
		c.Site.Scheme, project.Token, c.Listen.GRPCHost, c.Listen.GRPCPort, project.ID)
}

func (c *AppConfig) HTTPDsn(project *Project) string {
	return fmt.Sprintf("%s://%s@%s:%s/%d",
		c.Site.Scheme, project.Token, c.Listen.HTTPHost, c.Listen.HTTPPort, project.ID)
}

type BunConfig struct {
	DSN string `yaml:"dsn"`
}

type CHConfig struct {
	DSN string `yaml:"dsn"`
}

func scaleWithCPU(min, max int) int {
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
