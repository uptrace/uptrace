package metrics

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	collectormetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
)

const (
	protobufContentType = "application/x-protobuf"
	jsonContentType     = "application/json"
)

var jsonMarshaler = &jsonpb.Marshaler{}

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
	router := app.Router()

	router.WithGroup("", func(g *bunrouter.Group) {
		promHandler := NewPromHandler(app)

		g = g.Use(promHandler.CheckProjectAccess, promErrorHandler)

		g.GET("/api/v1/labels", promHandler.Labels)
		g.POST("/api/v1/labels", promHandler.Labels)
		g.GET("/api/v1/label/:label/values", promHandler.LabelValues)
		g.POST("/api/v1/query_range", promHandler.QueryRange)
		g.GET("/api/v1/query_range", promHandler.QueryRange)
		g.POST("/api/v1/query", promHandler.QueryInstant)
		g.GET("/api/v1/query", promHandler.QueryInstant)
		g.GET("/api/v1/metadata", promHandler.Metadata)
		g.GET("/api/v1/series", promHandler.Series)
		g.POST("/api/v1/series", promHandler.Series)
	})
}

func promErrorHandler(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		err := next(w, req)
		if err == nil {
			return nil
		}
		switch err := err.(type) {
		case *promError:
			return err
		default:
			return newPromError(err)
		}
	}
}

//------------------------------------------------------------------------------

type promError struct {
	Wrapped error `json:"error"`
}

var _ httperror.Error = (*promError)(nil)

func newPromError(err error) *promError {
	return &promError{
		Wrapped: err,
	}
}

func (e *promError) Error() string {
	if e.Wrapped == nil {
		return ""
	}
	return e.Wrapped.Error()
}

func (e *promError) Unwrap() error {
	return e.Wrapped
}

func (e *promError) HTTPStatusCode() int {
	return http.StatusBadRequest
}

func (e *promError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"status":    "error",
		"errorType": "error",
		"error":     e.Error(),
	})
}
