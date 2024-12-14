package metrics

import (
	"context"
	"net/http"

	"github.com/vmihailenco/taskq/v4"
	"go.opentelemetry.io/otel/metric"
	collectormetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/bunotel"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/run"
)

const (
	protobufContentType  = "application/protobuf"
	xprotobufContentType = "application/x-protobuf"
	jsonContentType      = "application/json"
)

var datapointCounter, _ = bunotel.Meter.Int64Counter(
	"uptrace.projects.datapoints",
	metric.WithDescription("Number of processed datapoints"),
)

var Module = fx.Module("metrics",
	fx.Provide(
		fx.Private,

		NewMiddleware,
		NewDatapointProcessor,
		NewMetricsServiceServer,

		NewMetricHandler,
		NewAttrHandler,
		NewQueryHandler,
		NewDashHandler,
		NewGridItemHandler,
		NewGridRowHandler,
		NewKinesisHandler,
		NewPrometheusHandler,
	),
	fx.Invoke(
		registerMetricHandler,
		registerAttrHandler,
		registerQueryHandler,
		registerDashHandler,
		registerGridItemHandler,
		registerGridRowHandler,
		registerKinesisHandler,
		registerPrometheusHandler,

		initOTLP,
		initTasks,
		runSpanMetrics,
		runDatapointProcessor,
	),
)

type OTLPParams struct {
	fx.In

	GRPC          *grpc.Server
	MetricsServer *MetricsServiceServer
}

func initOTLP(p OTLPParams, router bunapp.RouterParams) {
	collectormetricspb.RegisterMetricsServiceServer(p.GRPC, p.MetricsServer)

	router.Router.POST("/v1/metrics", p.MetricsServer.ExportHTTP)
}

func runDatapointProcessor(group *run.Group, dp *DatapointProcessor) {
	group.Add("metrics.DatapointProcessor.Run", func() error {
		dp.Run()
		return nil
	})
	group.OnStop(func(context.Context, error) error {
		dp.Stop()
		return nil
	})
}

func runSpanMetrics(lc fx.Lifecycle, conf *bunconf.Config, pg *bun.DB, ch *ch.DB) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		return initSpanMetrics(ctx, conf, pg, ch)
	}))
}

//------------------------------------------------------------------------------

type Middleware struct {
	*org.Middleware
}

func NewMiddleware(logger *otelzap.Logger, conf *bunconf.Config, pg *bun.DB) *Middleware {
	return &Middleware{
		Middleware: org.NewMiddleware(logger, conf, pg),
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

		dash, err := SelectDashboard(ctx, m.PG, dashID)
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

func initTasks(logger *otelzap.Logger, pg *bun.DB) {
	dashSyncer := NewDashSyncer(logger, pg)
	createDashboardsTask = bunapp.RegisterTaskHandler("create-dashboards", dashSyncer.CreateDashboardsHandler)
}
