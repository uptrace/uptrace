package bunapp

import (
	"context"

	"github.com/uptrace/uptrace-go/uptrace"
)

func setupOpentelemetry(app *App) {
	project := &app.Config().Projects[0]

	uptrace.ConfigureOpentelemetry(
		uptrace.WithDSN(app.conf.GRPCDsn(project)),
		uptrace.WithServiceName(app.conf.Service),
	)

	app.OnStopped("uptrace.Shutdown", func(ctx context.Context, _ *App) error {
		if false {
			return uptrace.Shutdown(ctx)
		}
		return nil
	})
}
