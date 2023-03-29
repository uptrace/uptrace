package metrics

import (
	"context"
	"fmt"
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

var measureCounter, _ = bunotel.Meter.Int64Counter(
	"uptrace.projects.measures",
	instrument.WithDescription("Number of processed measures"),
)

func Init(ctx context.Context, app *bunapp.App) {
	initOTLP(ctx, app)
	initRoutes(ctx, app)
	if err := initSpanMetrics(ctx, app); err != nil {
		app.Logger.Error("initSpanMetrics failed", zap.Error(err))
	}
}

func initOTLP(ctx context.Context, app *bunapp.App) {
	metricsService := NewMetricsServiceServer(app)
	collectormetricspb.RegisterMetricsServiceServer(app.GRPCServer(), metricsService)

	router := app.Router()
	router.POST("/v1/metrics", metricsService.ExportHTTP)
}

func initRoutes(ctx context.Context, app *bunapp.App) {
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

			g = g.NewGroup("/:dash_id").Use(middleware.Dashboard)

			g.GET("", dashHandler.Show)
			g.GET("/yaml", dashHandler.ShowYAML)
			g.POST("", dashHandler.Clone)
			g.PUT("", dashHandler.Update)
			g.PUT("/table", dashHandler.UpdateTable)
			g.PUT("/yaml", dashHandler.FromYAML)
			g.DELETE("", dashHandler.Delete)
			g.PUT("/pinned", dashHandler.Pin)
			g.PUT("/unpinned", dashHandler.Unpin)
		})

	api.
		Use(middleware.UserAndProject).
		Use(middleware.Dashboard).
		WithGroup("/metrics/:project_id/dashboards/:dash_id/grid", func(g *bunrouter.Group) {
			handler := NewGridColumnHandler(app)

			g.PUT("", handler.UpdateOrder)
			g.POST("", handler.Create)

			g = g.Use(middleware.GridColumn)

			g.PUT("/:col_id", handler.Update)
			g.DELETE("/:col_id", handler.Delete)
		})

	api.
		Use(middleware.UserAndProject).
		Use(middleware.Dashboard).
		WithGroup("/metrics/:project_id/dashboards/:dash_id/gauges", func(g *bunrouter.Group) {
			handler := NewDashGaugeHandler(app)

			g.GET("", handler.List)
			g.POST("", handler.Create)
			g.PUT("", handler.UpdateOrder)

			g = g.Use(middleware.DashGauge)

			g.PUT("/:gauge_id", handler.Update)
			g.DELETE("/:gauge_id", handler.Delete)
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

type gridColumnCtxKey struct{}

func (m *Middleware) GridColumn(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()
		dash := dashFromContext(ctx)

		colID, err := req.Params().Uint64("col_id")
		if err != nil {
			return err
		}

		col, err := SelectGridColumn(ctx, m.App, colID)
		if err != nil {
			return err
		}

		if col.Base().DashID != dash.ID {
			return org.ErrAccessDenied
		}

		ctx = context.WithValue(ctx, gridColumnCtxKey{}, col)
		return next(w, req.WithContext(ctx))
	}
}

func gridColumnFromContext(ctx context.Context) GridColumn {
	return ctx.Value(gridColumnCtxKey{}).(GridColumn)
}

func chartGridColumnFromContext(ctx context.Context) (*ChartGridColumn, error) {
	anyGridCol := ctx.Value(gridColumnCtxKey{})
	gridCol, ok := anyGridCol.(*ChartGridColumn)
	if !ok {
		return nil, fmt.Errorf("expected *ChartGridColumn, got %T", anyGridCol)
	}
	return gridCol, nil
}

func tableGridColumnFromContext(ctx context.Context) (*TableGridColumn, error) {
	anyGridCol := ctx.Value(gridColumnCtxKey{})
	gridCol, ok := anyGridCol.(*TableGridColumn)
	if !ok {
		return nil, fmt.Errorf("expected *TableGridColumn, got %T", anyGridCol)
	}
	return gridCol, nil
}

func heatmapGridColumnFromContext(ctx context.Context) (*HeatmapGridColumn, error) {
	anyGridCol := ctx.Value(gridColumnCtxKey{})
	gridCol, ok := anyGridCol.(*HeatmapGridColumn)
	if !ok {
		return nil, fmt.Errorf("expected *HeatmapGridColumn, got %T", anyGridCol)
	}
	return gridCol, nil
}

//------------------------------------------------------------------------------

type dashGaugeCtxKey struct{}

func (m *Middleware) DashGauge(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()
		dash := dashFromContext(ctx)

		gaugeID, err := req.Params().Uint64("gauge_id")
		if err != nil {
			return err
		}

		gauge, err := SelectDashGauge(ctx, m.App, dash.ID, gaugeID)
		if err != nil {
			return err
		}

		if gauge.ProjectID != dash.ProjectID {
			return org.ErrAccessDenied
		}

		ctx = context.WithValue(ctx, dashGaugeCtxKey{}, gauge)
		return next(w, req.WithContext(ctx))
	}
}

func dashGaugeFromContext(ctx context.Context) *DashGauge {
	return ctx.Value(dashGaugeCtxKey{}).(*DashGauge)
}
