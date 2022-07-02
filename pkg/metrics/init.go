package metrics

import (
	"context"

	"github.com/uptrace/uptrace/pkg/bunapp"
	collectormetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
)

func init() {
	bunapp.OnStart("metrics.init", func(ctx context.Context, app *bunapp.App) error {
		initGRPC(ctx, app)
		initRoutes(ctx, app)
		return nil
	})
}

func initGRPC(ctx context.Context, app *bunapp.App) {
	metricsService := NewMetricsServiceServer(app)
	collectormetricspb.RegisterMetricsServiceServer(app.GRPCServer(), metricsService)
}

func initRoutes(ctx context.Context, app *bunapp.App) {
}
