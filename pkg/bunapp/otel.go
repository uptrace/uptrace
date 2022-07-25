package bunapp

import (
	"context"

	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/global"
)

var (
	Meter  = global.Meter("github.com/uptrace/uptrace")
	Tracer = otel.Tracer("github.com/uptrace/uptrace")
)

func setupOpentelemetry(app *App) {
	project := &app.Config().Projects[0]

	uptrace.ConfigureOpentelemetry(
		uptrace.WithDSN(app.cfg.GRPCDsn(project)),
		uptrace.WithServiceName(app.cfg.Service),
	)

	app.OnStopped("uptrace.Shutdown", func(ctx context.Context, _ *App) error {
		return uptrace.Shutdown(ctx)
	})
}
