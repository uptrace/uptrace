package org

import (
	"fmt"
	"net/url"

	"github.com/uptrace/uptrace/pkg/bunconf"
)

type DSN struct {
	original string

	Scheme string
	Host   string
	Token  string
}

func (dsn *DSN) String() string {
	return dsn.original
}

func ParseDSN(dsnStr string) (*DSN, error) {
	u, err := url.Parse(dsnStr)
	if err != nil {
		return nil, fmt.Errorf("can't parse DSN=%q: %s", dsnStr, err)
	}

	dsn := DSN{
		original: dsnStr,
	}

	dsn.Scheme = u.Scheme
	if dsn.Scheme == "" {
		return nil, fmt.Errorf("DSN=%q does not have a scheme", dsnStr)
	}

	dsn.Host = u.Host
	if dsn.Host == "" {
		return nil, fmt.Errorf("DSN=%q does not have a host", dsnStr)
	}
	if dsn.Host == "api.uptrace.dev" {
		dsn.Host = "uptrace.dev"
	}

	if u.User != nil {
		dsn.Token = u.User.Username()
	}
	if dsn.Token == "" {
		return nil, fmt.Errorf("DSN=%q does not have a token", dsnStr)
	}

	return &dsn, nil
}

func BuildDSN(conf *bunconf.Config, token string) string {
	return fmt.Sprintf("%s://%s@%s:%s?grpc=%s",
		conf.Listen.Scheme,
		token,
		conf.Site.Host,
		conf.Listen.HTTP.Port,
		conf.Listen.GRPC.Port)
}
