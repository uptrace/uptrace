package metrics

import (
	"context"
	"net/http"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"go.opentelemetry.io/otel/metric/instrument"
	collectormetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"go.uber.org/zap"
)

const (
	numMetricLimit = 6
	numEntryLimit  = 20
)

const (
	protobufContentType = "application/x-protobuf"
	jsonContentType     = "application/json"
)

var jsonMarshaler = &jsonpb.Marshaler{}

var measureCounter, _ = bunotel.Meter.SyncInt64().Counter(
	"uptrace.projects.measures",
	instrument.WithDescription("Number of processed measures"),
)

func Init(ctx context.Context, app *bunapp.App) {
	initGRPC(ctx, app)
	initRoutes(ctx, app)
	if err := initSpanMetrics(ctx, app); err != nil {
		app.Logger.Error("initSpanMetrics failed", zap.Error(err))
	}
}

func initGRPC(ctx context.Context, app *bunapp.App) {
	metricsService := NewMetricsServiceServer(app)
	collectormetricspb.RegisterMetricsServiceServer(app.GRPCServer(), metricsService)
}

func initRoutes(ctx context.Context, app *bunapp.App) {
	middleware := NewMiddleware(app)
	api := app.APIGroup().
		NewGroup("/metrics/:project_id").
		Use(middleware.UserAndProject)

	api.
		WithGroup("", func(g *bunrouter.Group) {
			metricHandler := NewMetricHandler(app)

			g.GET("", metricHandler.List)

			g = g.NewGroup("/:metric_id").Use(middleware.Metric)

			g.GET("", metricHandler.Show)
			g.GET("/attributes", metricHandler.Attributes)
			g.GET("/where", metricHandler.Where)
		})

	api.
		WithGroup("", func(g *bunrouter.Group) {
			attrHandler := NewAttrHandler(app)

			g.GET("/attributes", attrHandler.Keys)
			g.GET("/where", attrHandler.Where)
		})

	api.
		WithGroup("", func(g *bunrouter.Group) {
			queryHandler := NewQueryHandler(app)

			g.GET("/table", queryHandler.Table)
			g.GET("/timeseries", queryHandler.Timeseries)
			g.GET("/gauge", queryHandler.Gauge)
		})

	api.
		WithGroup("/dashboards", func(g *bunrouter.Group) {
			dashHandler := NewDashHandler(app)

			g.POST("", dashHandler.Create)
			g.GET("", dashHandler.List)

			g = g.NewGroup("/:dash_id").Use(middleware.Dashboard)

			g.GET("", dashHandler.Show)
			g.POST("", dashHandler.Clone)
			g.PUT("", dashHandler.Update)
			g.DELETE("", dashHandler.Delete)
		})

	api.
		Use(middleware.Dashboard).
		WithGroup("/dashboards/:dash_id/entries", func(g *bunrouter.Group) {
			dashEntryHandler := NewDashEntryHandler(app)

			g.GET("", dashEntryHandler.List)
			g.POST("", dashEntryHandler.Create)
			g.PUT("", dashEntryHandler.UpdateOrder)
			g.PUT("/:id", dashEntryHandler.Update)
			g.DELETE("/:id", dashEntryHandler.Delete)
		})
}

//------------------------------------------------------------------------------

type Middleware struct {
	*org.Middleware
	app *bunapp.App
}

func NewMiddleware(app *bunapp.App) Middleware {
	return Middleware{
		Middleware: org.NewMiddleware(app),
		app:        app,
	}
}

type dashCtxKey struct{}

func (m Middleware) Dashboard(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()

		dashID, err := req.Params().Uint64("dash_id")
		if err != nil {
			return err
		}

		dash, err := SelectDashboard(ctx, m.app, dashID)
		if err != nil {
			return err
		}

		project := org.ProjectFromContext(ctx)
		if dash.ProjectID != project.ID {
			return org.ErrAccessDenied
		}

		ctx = context.WithValue(ctx, dashCtxKey{}, dash)
		return next(w, req.WithContext(ctx))
	}
}

func dashFromContext(ctx context.Context) *Dashboard {
	return ctx.Value(dashCtxKey{}).(*Dashboard)
}

//------------------------------------------------------------------------------

type metricCtxKey struct{}

func (m Middleware) Metric(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()

		metricID, err := req.Params().Uint64("metric_id")
		if err != nil {
			return err
		}

		metric, err := SelectMetric(ctx, m.app, metricID)
		if err != nil {
			return err
		}

		project := org.ProjectFromContext(ctx)
		if metric.ProjectID != project.ID {
			return org.ErrAccessDenied
		}

		ctx = context.WithValue(ctx, metricCtxKey{}, metric)
		return next(w, req.WithContext(ctx))
	}
}

func metricFromContext(ctx context.Context) *Metric {
	return ctx.Value(metricCtxKey{}).(*Metric)
}
