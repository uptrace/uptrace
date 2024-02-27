package metrics

import (
	"context"
	"net/http"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/vmihailenco/taskq/v4"
	"go.opentelemetry.io/otel/metric"
	collectormetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"go.uber.org/zap"
)

const (
	numMetricLimit = 6
	numEntryLimit  = 20
)

const (
	protobufContentType  = "application/protobuf"
	xprotobufContentType = "application/x-protobuf"
	jsonContentType      = "application/json"
)

var jsonMarshaler = &jsonpb.Marshaler{}

var datapointCounter, _ = bunotel.Meter.Int64Counter(
	"uptrace.projects.datapoints",
	metric.WithDescription("Number of processed datapoints"),
)

func Init(ctx context.Context, app *bunapp.App) {
	mp := NewDatapointProcessor(app)

	initTasks(ctx, app)
	initOTLP(ctx, app, mp)
	initRoutes(ctx, app, mp)
	if err := initSpanMetrics(ctx, app); err != nil {
		app.Logger.Error("initSpanMetrics failed", zap.Error(err))
	}
}

func initOTLP(ctx context.Context, app *bunapp.App, mp *DatapointProcessor) {
	metricsService := NewMetricsServiceServer(app, mp)
	collectormetricspb.RegisterMetricsServiceServer(app.GRPCServer(), metricsService)

	router := app.Router()
	router.POST("/v1/metrics", metricsService.ExportHTTP)
}

func initRoutes(ctx context.Context, app *bunapp.App, mp *DatapointProcessor) {
	router := app.Router()
	middleware := NewMiddleware(app)
	api := app.APIGroup()

	api.
		Use(middleware.UserAndProject).
		WithGroup("/metrics/:project_id", func(g *bunrouter.Group) {
			metricHandler := NewMetricHandler(app)

			g.GET("", metricHandler.List)
			g.GET("/describe", metricHandler.Describe)
			g.GET("/stats", metricHandler.Stats)
		})

	api.
		Use(middleware.UserAndProject).
		WithGroup("/metrics/:project_id", func(g *bunrouter.Group) {
			attrHandler := NewAttrHandler(app)

			g.GET("/attr-keys", attrHandler.AttrKeys)
			g.GET("/attr-values", attrHandler.AttrValues)
		})

	api.
		Use(middleware.UserAndProject).
		WithGroup("/metrics/:project_id", func(g *bunrouter.Group) {
			queryHandler := NewQueryHandler(app)

			g.GET("/table", queryHandler.Table)
			g.GET("/timeseries", queryHandler.Timeseries)
			g.GET("/gauge", queryHandler.Gauge)
			g.GET("/heatmap", queryHandler.Heatmap)
		})

	api.
		Use(middleware.UserAndProject).
		WithGroup("/metrics/:project_id/dashboards", func(g *bunrouter.Group) {
			dashHandler := NewDashHandler(app)

			g.POST("", dashHandler.Create)
			g.GET("", dashHandler.List)
			g.POST("/yaml", dashHandler.CreateFromYAML)

			g = g.NewGroup("/:dash_id").Use(middleware.Dashboard)

			g.GET("", dashHandler.Show)
			g.GET("/yaml", dashHandler.ShowYAML)
			g.POST("", dashHandler.Clone)
			g.PUT("", dashHandler.Update)
			g.PUT("/yaml", dashHandler.UpdateYAML)
			g.PUT("/table", dashHandler.UpdateTable)
			g.PUT("/grid", dashHandler.UpdateGrid)
			g.PUT("/reset", dashHandler.Reset)
			g.DELETE("", dashHandler.Delete)
			g.PUT("/pinned", dashHandler.Pin)
			g.PUT("/unpinned", dashHandler.Unpin)
		})

	api.
		Use(middleware.UserAndProject).
		Use(middleware.Dashboard).
		WithGroup("/metrics/:project_id/dashboards/:dash_id/grid", func(g *bunrouter.Group) {
			handler := NewGridItemHandler(app)

			g.POST("", handler.Create)
			g.PUT("/layout", handler.UpdateLayout)

			g = g.Use(handler.GridItemMiddleware)

			g.PUT("/:row_id", handler.Update)
			g.DELETE("/:row_id", handler.Delete)
		})

	api.
		Use(middleware.UserAndProject).
		Use(middleware.Dashboard).
		WithGroup("/metrics/:project_id/dashboards/:dash_id/rows", func(g *bunrouter.Group) {
			handler := NewGridRowHandler(app)

			g.POST("", handler.Create)

			g = g.Use(handler.GridRowMiddleware)

			g.GET("/:row_id", handler.Show)
			g.PUT("/:row_id", handler.Update)
			g.PUT("/:row_id/up", handler.MoveUp)
			g.PUT("/:row_id/down", handler.MoveDown)
			g.DELETE("/:row_id", handler.Delete)
		})

	router.WithGroup("/api/v1/cloudwatch", func(g *bunrouter.Group) {
		handler := NewKinesisHandler(app, mp)

		g.POST("/metrics", handler.Metrics)
	})

	router.WithGroup("/api/v1/prometheus", func(g *bunrouter.Group) {
		handler := NewPrometheusHandler(app, mp)

		g.POST("/write", handler.Write)
		g.POST("/read", handler.Read)
	})
}

//------------------------------------------------------------------------------

type Middleware struct {
	*org.Middleware
	App *bunapp.App
}

func NewMiddleware(app *bunapp.App) Middleware {
	return Middleware{
		Middleware: org.NewMiddleware(app),
		App:        app,
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

		dash, err := SelectDashboard(ctx, m.App, dashID)
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

var createDashboardsTask *taskq.Task

func initTasks(ctx context.Context, app *bunapp.App) {
	dashSyncer := NewDashSyncer(app)
	createDashboardsTask = app.RegisterTask("create-dashboards", &taskq.TaskConfig{
		Handler: dashSyncer.CreateDashboardsHandler,
	})
}
