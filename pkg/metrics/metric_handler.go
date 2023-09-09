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

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type MetricFilter struct {
	org.TimeFilter

	ProjectID uint32

	Instrument  Instrument
	AttrKey     []string
	SearchInput string

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

	if f.Instrument != "" {
		q = q.Where("instrument = ?", f.Instrument)
	}
	if len(f.AttrKey) > 0 {
		q = q.Where("attr_keys @> ?", pgdialect.Array(f.AttrKey))
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

	tableName := measureTableForWhere(app, &f.TimeFilter)
	q := app.CH.NewSelect().
		ColumnExpr("metric AS name").
		ColumnExpr("any(instrument) AS instrument").
		ColumnExpr("any(string_keys) AS attr_keys").
		ColumnExpr("uniqCombined64(attrs_hash) AS num_timeseries").
		TableExpr("?", tableName).
		Where("project_id = ?", f.ProjectID).
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT).
		GroupExpr("metric").
		OrderExpr("metric ASC").
		Limit(limit)

	if f.Instrument != "" {
		q = q.Where("instrument = ?", f.Instrument)
	}
	if len(f.AttrKey) > 0 {
		q = q.Where("hasAll(string_keys, ?)", chschema.Array(f.AttrKey))
	}
	if f.SearchInput != "" {
		values := strings.Split(f.SearchInput, "|")
		q = q.Where("multiSearchAnyCaseInsensitiveUTF8(metric, ?) != 0", ch.Array(values))
	}

	if f.Query != "" {
		query := mql.Parse(f.Query)
		for _, part := range query.Parts {
			if part.Error.Wrapped != nil {
				continue
			}

			switch v := part.AST.(type) {
			case *ast.Where:
				where, err := compileFilters(v.Filters)
				if err != nil {
					return nil, false, err
				}
				if len(where) > 0 {
					q = q.Where(where)
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

	names := make([]string, len(metrics))
	for i, metric := range metrics {
		names[i] = metric.Name
	}

	if err := app.PG.NewSelect().
		With("data", app.PG.NewValues(&metrics).WithOrder()).
		ColumnExpr("m.id, m.unit, m.description").
		Model(&metrics).
		TableExpr("data").
		Where("m.project_id = ?", f.ProjectID).
		Where("m.name = data.name").
		OrderExpr("data._order").
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
