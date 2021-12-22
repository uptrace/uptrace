package bunapp

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func ReadConfig(configFile, service string) (*AppConfig, error) {
	configFile, err := filepath.Abs(configFile)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	cfg := new(AppConfig)
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}

	cfg.Filepath = configFile
	cfg.Service = service

	httpHost, httpPort, err := net.SplitHostPort(cfg.Listen.HTTP)
	if err != nil {
		return nil, fmt.Errorf("can't parse HTTP addr: %w", err)
	}
	if httpHost == "" {
		httpHost = cfg.Site.Host
	}
	cfg.Listen.HTTPHost = httpHost
	cfg.Listen.HTTPPort = httpPort

	grpcHost, grpcPort, err := net.SplitHostPort(cfg.Listen.GRPC)
	if err != nil {
		return nil, fmt.Errorf("can't parse GRPC addr: %w", err)
	}
	if grpcHost == "" {
		grpcHost = cfg.Site.Host
	}
	cfg.Listen.GRPCHost = grpcHost
	cfg.Listen.GRPCPort = grpcPort

	return cfg, nil
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

		HTTPHost string `yaml:"-"`
		HTTPPort string `yaml:"-"`

		GRPCHost string `yaml:"-"`
		GRPCPort string `yaml:"-"`
	} `yaml:"listen"`

	DB BunConfig `yaml:"db"`
	CH CHConfig  `yaml:"ch"`

	Retention struct {
		TTL string `yaml:"ttl"`
	} `yaml:"retention"`
}

func (c *AppConfig) SiteAddr() string {
	return fmt.Sprintf("%s://%s:%s/", c.Site.Scheme, c.Listen.HTTPHost, c.Listen.HTTPPort)
}

func (c *AppConfig) OTLPEndpoint() string {
	return fmt.Sprintf("%s://%s:%s", c.Site.Scheme, c.Listen.GRPCHost, c.Listen.GRPCPort)
}

func (c *AppConfig) UptraceDSN() string {
	return fmt.Sprintf("%s://%s:%s", c.Site.Scheme, c.Listen.GRPCHost, c.Listen.GRPCPort)
}

type BunConfig struct {
	DSN string `yaml:"dsn"`
}

type CHConfig struct {
	DSN string `yaml:"dsn"`
}
