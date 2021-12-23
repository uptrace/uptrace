package bunapp

import (
	"context"

	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
)

var Tracer = otel.Tracer("github.com/uptrace/uptrace")

func setupOpentelemetry(app *App) {
	uptrace.ConfigureOpentelemetry(
		uptrace.WithMetricsEnabled(false),
		uptrace.WithDSN(app.cfg.OTLPGrpc()),
		uptrace.WithServiceName(app.cfg.Service),
	)

	app.OnStopped("uptrace.Shutdown", func(ctx context.Context, _ *App) error {
		return uptrace.Shutdown(ctx)
	})
}
