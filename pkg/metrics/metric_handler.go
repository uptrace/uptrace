package metrics

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/chschema"

	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type MetricFilter struct {
	org.TimeFilter

	ProjectID uint32

	AttrKey         []string
	Instrument      []string
	OtelLibraryName []string
	SearchInput     string

	Query string
}

func DecodeMetricFilter(req bunrouter.Request, f *MetricFilter) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)
	f.ProjectID = project.ID

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return err
	}

	return nil
}

var _ urlstruct.ValuesUnmarshaler = (*MetricFilter)(nil)

func (f *MetricFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	return nil
}

//-----------------------------------------------------------------------------------------

type MetricHandler struct {
	*bunapp.App
}

func NewMetricHandler(app *bunapp.App) *MetricHandler {
	return &MetricHandler{
		App: app,
	}
}

func (h *MetricHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	if err := h.createSystemMetrics(ctx, project.ID); err != nil {
		return err
	}

	f := new(MetricFilter)
	now := time.Now()
	f.TimeGTE = now.Add(-24 * time.Hour)
	f.TimeLT = now

	if err := DecodeMetricFilter(req, f); err != nil {
		return err
	}

	metrics, err := selectMetrics(ctx, h.App, f)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"metrics": metrics,
	})
}

func (h *MetricHandler) createSystemMetrics(ctx context.Context, projectID uint32) error {
	achievements, err := org.SelectAchievements(ctx, h.App, 0, projectID)
	if err != nil {
		return err
	}

	var tracingConfigured bool
	for _, achievement := range achievements {
		switch achievement.Name {
		case org.AchievConfigureTracing:
			tracingConfigured = true
		}
	}

	var metrics []Metric

	if tracingConfigured {
		metrics = append(metrics,
			Metric{
				ProjectID:       projectID,
				Name:            uptraceTracingSpans,
				Description:     "Number of spans and their duration (excluding events and logs)",
				Instrument:      InstrumentHistogram,
				Unit:            bunconv.UnitMicroseconds,
				AttrKeys:        []string{attrkey.ServiceName, attrkey.HostName},
				OtelLibraryName: uptraceLibraryName,
			},
			Metric{
				ProjectID:       projectID,
				Name:            uptraceTracingEvents,
				Description:     "Number of events (excluding spans and logs)",
				Instrument:      InstrumentCounter,
				AttrKeys:        []string{attrkey.ServiceName, attrkey.HostName},
				OtelLibraryName: uptraceLibraryName,
			},
			Metric{
				ProjectID:       projectID,
				Name:            uptraceTracingLogs,
				Description:     "Number of logs (excluding spans and events)",
				Instrument:      InstrumentCounter,
				AttrKeys:        []string{attrkey.ServiceName, attrkey.HostName},
				OtelLibraryName: uptraceLibraryName,
			},

			Metric{
				ProjectID:   projectID,
				Name:        uptraceServiceGraphClientDuration,
				Description: "Requests duration between two nodes as seen from the client",
				Instrument:  InstrumentSummary,
				Unit:        bunconv.UnitMicroseconds,
				AttrKeys: []string{
					"type",
					"client",
					"server",
					attrkey.DeploymentEnvironment,
					attrkey.ServiceNamespace,
				},
				OtelLibraryName: uptraceLibraryName,
			},
			Metric{
				ProjectID:   projectID,
				Name:        uptraceServiceGraphServerDuration,
				Description: "Requests duration between two nodes as seen from the server",
				Instrument:  InstrumentSummary,
				Unit:        bunconv.UnitMicroseconds,
				AttrKeys: []string{
					"type",
					"client",
					"server",
					attrkey.DeploymentEnvironment,
					attrkey.ServiceNamespace,
				},
				OtelLibraryName: uptraceLibraryName,
			},
			Metric{
				ProjectID:   projectID,
				Name:        uptraceServiceGraphFailedRequests,
				Description: "Total count of failed requests between two nodes",
				Instrument:  InstrumentCounter,
				AttrKeys: []string{
					"type",
					"client",
					"server",
					attrkey.DeploymentEnvironment,
					attrkey.ServiceNamespace,
				},
				OtelLibraryName: uptraceLibraryName,
			},
		)
	}

	return UpsertMetrics(ctx, h.App, metrics)
}

func selectMetrics(ctx context.Context, app *bunapp.App, f *MetricFilter) ([]*Metric, error) {
	if f.Query != "" {
		metrics, _, err := selectMetricsFromCH(ctx, app, f)
		return metrics, err
	}

	var metrics []*Metric

	q := app.PG.NewSelect().
		Model(&metrics).
		Where("project_id = ?", f.ProjectID).
		Where("updated_at IS NULL OR updated_at >= ?", time.Now().Add(-24*time.Hour)).
		OrderExpr("name ASC").
		Limit(10000)

	if len(f.AttrKey) > 0 {
		q = q.Where("attr_keys @> ?", pgdialect.Array(f.AttrKey))
	}
	if len(f.Instrument) > 0 {
		q = q.Where("instrument IN (?)", bun.In(f.Instrument))
	}
	if len(f.OtelLibraryName) > 0 {
		q = q.Where("otel_library_name IN (?)", bun.In(f.OtelLibraryName))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return metrics, nil
}

func selectMetricsFromCH(
	ctx context.Context, app *bunapp.App, f *MetricFilter,
) ([]*Metric, bool, error) {
	const limit = 1000

	tableName := DatapointTableForWhere(&f.TimeFilter)
	q := app.CH.NewSelect().
		ColumnExpr("m.metric AS name").
		ColumnExpr("anyLast(m.instrument) AS instrument").
		ColumnExpr("uniqCombined64(m.attrs_hash) AS num_timeseries").
		TableExpr("? AS m", ch.Name(tableName)).
		Where("m.project_id = ?", f.ProjectID).
		Where("m.time >= ?", f.TimeGTE).
		Where("m.time < ?", f.TimeLT).
		GroupExpr("m.metric").
		OrderExpr("name ASC").
		Limit(limit)

	if len(f.AttrKey) > 0 {
		q = q.Where("hasAll(m.string_keys, ?)", chschema.Array(f.AttrKey))
	}
	if len(f.Instrument) > 0 {
		q = q.Where("m.instrument IN ?", ch.In(f.Instrument))
	}
	if len(f.OtelLibraryName) > 0 {
		q = q.Where("m.otel_library_name IN ?", ch.In(f.OtelLibraryName))
	}
	if f.SearchInput != "" {
		values := strings.Split(f.SearchInput, "|")
		q = q.Where("multiSearchAnyCaseInsensitiveUTF8(metric, ?) != 0", ch.Array(values))
	}

	if f.Query != "" {
		query := mql.ParseQuery(f.Query)
		for _, part := range query.Parts {
			if part.Error.Wrapped != nil {
				continue
			}

			switch v := part.AST.(type) {
			case *ast.Where:
				if err := compileFilters(q, InstrumentDeleted, v.Filters); err != nil {
					return nil, false, err
				}
			default:
				return nil, false, fmt.Errorf("unsupported AST type: %T", v)
			}
		}
	}

	metrics := make([]*Metric, 0)
	if err := q.Scan(ctx, &metrics); err != nil {
		return nil, false, err
	}

	if len(metrics) == 0 {
		return metrics, false, nil
	}

	if err := app.PG.NewSelect().
		With("data", app.PG.NewValues(&metrics).Column("name", "num_timeseries").WithOrder()).
		ColumnExpr("m.*, data.num_timeseries").
		Model(&metrics).
		TableExpr("data").
		Where("m.project_id = ?", f.ProjectID).
		Where("m.name = data.name").
		OrderExpr("data._order ASC").
		Scan(ctx); err != nil {
		return nil, false, err
	}

	return metrics, len(metrics) == limit, nil
}

func (h *MetricHandler) Describe(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	metricNames := req.URL.Query()["metric[]"]
	metrics := make([]*Metric, 0, len(metricNames))

	if len(metricNames) == 0 {
		return httputil.JSON(w, bunrouter.H{
			"metrics": metrics,
		})
	}

	if err := h.PG.NewSelect().
		Model(&metrics).
		Where("project_id = ?", project.ID).
		Where("name IN (?)", bun.In(metricNames)).
		Limit(1000).
		Scan(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"metrics": metrics,
	})
}

func (h *MetricHandler) Stats(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := new(MetricFilter)
	if err := DecodeMetricFilter(req, f); err != nil {
		return err
	}

	metrics, hasMore, err := selectMetricsFromCH(ctx, h.App, f)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"metrics": metrics,
		"hasMore": hasMore,
	})
}
