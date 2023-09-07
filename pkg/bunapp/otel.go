package bunapp

import (
	"context"

	"github.com/uptrace/uptrace-go/uptrace"
)

func configureOpentelemetry(app *App) error {
	conf := app.Config()
	project := &conf.Projects[0]

	if conf.UptraceGo.Disabled {
		return nil
	}

	var options []uptrace.Option

	options = append(options, uptrace.WithServiceName(app.conf.Service))

	if conf.UptraceGo.DSN == "" {
		options = append(options, uptrace.WithDSN(app.conf.GRPCDsn(project.ID, project.Token)))
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

	app.OnStopped("uptrace.Shutdown", func(ctx context.Context, _ *App) error {
		if false {
			return uptrace.Shutdown(ctx)
		}
		return nil
	})

	return nil
}
