package metrics

import (
	"context"

	"github.com/uptrace/uptrace/pkg/bunapp"
	collectormetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
)

type MetricsServiceServer struct {
	collectormetricspb.UnimplementedMetricsServiceServer

	*bunapp.App
}

func NewMetricsServiceServer(app *bunapp.App) *MetricsServiceServer {
	return &MetricsServiceServer{
		App: app,
	}
}

var _ collectormetricspb.MetricsServiceServer = (*MetricsServiceServer)(nil)

func (s *MetricsServiceServer) Export(
	ctx context.Context, req *collectormetricspb.ExportMetricsServiceRequest,
) (*collectormetricspb.ExportMetricsServiceResponse, error) {
	return &collectormetricspb.ExportMetricsServiceResponse{}, nil
}
