package org

import (
	"fmt"
	"net/url"
)

type DSN struct {
	original string

	Scheme string
	Host   string

	ProjectID string
	Token     string
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

	if len(u.Path) > 0 {
		dsn.ProjectID = u.Path[1:]
	}
	if dsn.ProjectID == "" {
		return nil, fmt.Errorf("DSN=%q does not have a project id", dsnStr)
	}

	if u.User != nil {
		dsn.Token = u.User.Username()
	}
	if dsn.Token == "" {
		return nil, fmt.Errorf("DSN=%q does not have a token", dsnStr)
	}

	return &dsn, nil
}
