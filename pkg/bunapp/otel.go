package bunapp

import (
	"context"

	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
)

var Tracer = otel.Tracer("github.com/uptrace/uptrace")

func setupOpentelemetry(app *App) {
	project := &app.Config().Projects[0]

	uptrace.ConfigureOpentelemetry(
		uptrace.WithMetricsEnabled(false),
		uptrace.WithDSN(app.cfg.GRPCDsn(project)),
		uptrace.WithServiceName(app.cfg.Service),
	)

	app.OnStopped("uptrace.Shutdown", func(ctx context.Context, _ *App) error {
		return uptrace.Shutdown(ctx)
	})
}
