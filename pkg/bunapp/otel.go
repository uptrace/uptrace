package bunapp

import (
	"context"

	"github.com/uptrace/uptrace-go/uptrace"
)

func configureOpentelemetry(app *App) error {
	conf := app.Config()
	project := &conf.Projects[0]

	var options []uptrace.Option
	options = append(options,
		uptrace.WithDSN(app.conf.GRPCDsn(project)),
		uptrace.WithServiceName(app.conf.Service),
	)
	uptrace.ConfigureOpentelemetry(options...)

	app.OnStopped("uptrace.Shutdown", func(ctx context.Context, _ *App) error {
		if false {
			return uptrace.Shutdown(ctx)
		}
		return nil
	})

	return nil
}
