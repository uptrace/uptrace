package uptracebundle

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/uptrace/uptrace-go/uptrace"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/urfave/cli/v2"
)

func StartCLI(c *cli.Context) (context.Context, *bunapp.App, error) {
	return Start(c.Context, c.String("config"), c.Command.Name)
}

func Start(ctx context.Context, confPath, service string) (context.Context, *bunapp.App, error) {
	if confPath == "" {
		var err error
		confPath, err = findConfigPath()
		if err != nil {
			return nil, nil, err
		}
	}

	conf, err := bunconf.ReadConfig(confPath, service)
	if err != nil {
		return nil, nil, err
	}
	return StartConfig(ctx, conf)
}

func findConfigPath() (string, error) {
	files := []string{
		"uptrace.yml",
		"config/uptrace.yml",
		"/etc/uptrace/uptrace.yml",
	}
	for _, confPath := range files {
		if _, err := os.Stat(confPath); err == nil {
			return confPath, nil
		}
	}
	return "", fmt.Errorf("can't find uptrace.yml in usual locations")
}

func StartConfig(ctx context.Context, conf *bunconf.Config) (context.Context, *bunapp.App, error) {
	rand.Seed(time.Now().UnixNano())

	app, err := bunapp.New(ctx, conf)
	if err != nil {
		return ctx, nil, err
	}

	switch conf.Service {
	case "serve":
		if err := configureOpentelemetry(app); err != nil {
			return nil, nil, err
		}
	}

	return app.Context(), app, nil
}

func configureOpentelemetry(app *bunapp.App) error {
	conf := app.Config()
	project := &conf.Projects[0]

	if conf.UptraceGo.Disabled {
		return nil
	}

	var options []uptrace.Option

	options = append(options,
		uptrace.WithServiceName(conf.Service),
		uptrace.WithDeploymentEnvironment("self-hosted"))

	if conf.UptraceGo.DSN == "" {
		dsn := org.BuildDSN(conf, project.Token)
		options = append(options, uptrace.WithDSN(dsn))
	} else {
		options = append(options, uptrace.WithDSN(conf.UptraceGo.DSN))
	}

	if conf.UptraceGo.TLS != nil {
		tlsConf, err := conf.UptraceGo.TLS.TLSConfig()
		if err != nil {
			return err
		}
		options = append(options, uptrace.WithTLSConfig(tlsConf))
	}

	uptrace.ConfigureOpentelemetry(options...)

	app.OnStopped("uptrace.Shutdown", func(ctx context.Context, _ *bunapp.App) error {
		if false {
			return uptrace.Shutdown(ctx)
		}
		return nil
	})

	return nil
}
