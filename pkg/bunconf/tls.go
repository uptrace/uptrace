package bunconf

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultMinTLSVersion = tls.VersionTLS12
	defaultMaxTLSVersion = 0 // causes to use the default MaxVersion from "crypto/tls"
)

type TLS struct {
	CAFile     string `yaml:"ca_file"`
	CertFile   string `yaml:"cert_file"`
	KeyFile    string `yaml:"key_file"`
	MinVersion string `yaml:"min_version"`
	MaxVersion string `yaml:"max_version"`
}

func (c *TLS) tlsConfig() (*tls.Config, error) {
	var certPool *x509.CertPool
	if len(c.CAFile) != 0 {
		var err error
		certPool, err = c.readCert(c.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA %s: %w", c.CAFile, err)
		}
	}

	minTLS, err := tlsVersion(c.MinVersion, defaultMinTLSVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid TLS min_version: %w", err)
	}
	maxTLS, err := tlsVersion(c.MaxVersion, defaultMaxTLSVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid TLS max_version: %w", err)
	}

	if c.CertFile == "" && c.KeyFile == "" {
		return &tls.Config{
			RootCAs:              certPool,
			GetCertificate:       nil,
			GetClientCertificate: nil,
			MinVersion:           minTLS,
			MaxVersion:           maxTLS,
		}, nil
	}

	if c.CertFile == "" {
		return nil, fmt.Errorf("cert_file is required")
	}
	if c.KeyFile == "" {
		return nil, fmt.Errorf("key_file is required")
	}

	cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		RootCAs: certPool,
		GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return &cert, nil
		},
		GetClientCertificate: func(cri *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return &cert, nil
		},
		MinVersion: minTLS,
		MaxVersion: maxTLS,
	}, nil
}

func (c *TLS) readCert(caPath string) (*x509.CertPool, error) {
	caPEM, err := os.ReadFile(filepath.Clean(caPath))
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPEM) {
		return nil, err
	}
	return certPool, nil
}

var tlsVersions = map[string]uint16{
	"1.0": tls.VersionTLS10,
	"1.1": tls.VersionTLS11,
	"1.2": tls.VersionTLS12,
	"1.3": tls.VersionTLS13,
}

func tlsVersion(v string, defaultVersion uint16) (uint16, error) {
	if v == "" {
		return defaultVersion, nil
	}
	val, ok := tlsVersions[v]
	if !ok {
		return 0, fmt.Errorf("unsupported TLS version: %q", v)
	}
	return val, nil
}

//------------------------------------------------------------------------------

type TLSClient struct {
	TLS `yaml:",inline"`

	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	ServerName         string `yaml:"server_name_override"`
}

func (c *TLSClient) TLSConfig() (*tls.Config, error) {
	tlsConf, err := c.tlsConfig()
	if err != nil {
		return nil, err
	}

	tlsConf.ServerName = c.ServerName
	tlsConf.InsecureSkipVerify = c.InsecureSkipVerify

	return tlsConf, nil
}

type TLSServer struct {
	TLS `yaml:",inline"`

	ClientCAFile string `yaml:"client_ca_file"`
}

func (c *TLSServer) TLSConfig() (*tls.Config, error) {
	tlsConf, err := c.tlsConfig()
	if err != nil {
		return nil, err
	}

	if c.ClientCAFile != "" {
		certPool, err := c.readCert(c.ClientCAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client CA %s: %w", c.ClientCAFile, err)
		}
		tlsConf.ClientCAs = certPool
		tlsConf.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConf, nil
}
