package tracing

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/fx"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
)

type SystemHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	CH     *ch.DB
}

type SystemHandler struct {
	*SystemHandlerParams
}

func NewSystemHandler(p SystemHandlerParams) *SystemHandler {
	return &SystemHandler{&p}
}

func registerSystemHandler(h *SystemHandler, p bunapp.RouterParams, m *org.Middleware) {
	p.RouterInternalV1.
		Use(m.UserAndProject).
		WithGroup("/tracing/:project_id", func(g *bunrouter.Group) {
			g.GET("/systems", h.ListSystems)
		})
}

func (h *SystemHandler) ListSystems(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := &SpanFilter{}
	if err := DecodeSpanFilter(req, f); err != nil {
		return err
	}
	f.GroupID = 0

	systems, err := h.selectSystems(ctx, f)
	if err != nil {
		return err
	}

	if len(systems) > 0 || f.Query != "" {
		return httputil.JSON(w, bunrouter.H{
			"systems": systems,
		})
	}

	dataHint, err := h.selectDataHint(ctx, f)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"systems":  systems,
		"dataHint": dataHint,
	})
}

func (h *SystemHandler) selectSystems(
	ctx context.Context, f *SpanFilter,
) ([]map[string]any, error) {
	systems := make([]map[string]any, 0)
	tmp := make([]map[string]any, 0)

	for _, table := range []string{"spans_index", "logs_index", "events_index"} {
		query := h.CH.NewSelect().
			TableExpr("? AS s", ch.Name(table)).
			ColumnExpr("s.project_id AS projectId").
			ColumnExpr("s.system").
			ColumnExpr("sum(s.count) AS count").
			ColumnExpr("sum(s.count) / ? AS rate", f.TimeFilter.Duration().Minutes()).
			ColumnExpr("uniqCombined64(s.group_id) AS groupCount").
			GroupExpr("project_id, system").
			OrderExpr("system ASC").
			Limit(1000)

		if table == "spans_index" {
			query.
				ColumnExpr("sumIf(s.count, s.status_code = 'error') AS errorCount").
				ColumnExpr("sumIf(s.count, s.status_code = 'error') / sum(s.count) AS errorRate")
		} else {
			query.
				ColumnExpr("0 AS errorCount").
				ColumnExpr("0 AS errorRate")
		}

		if err := query.Scan(ctx, &tmp); err != nil {
			return nil, err
		}

		systems = append(systems, tmp...)
		clear(tmp)
		tmp = tmp[:0]
	}

	return systems, nil
}

func (h *SystemHandler) selectDataHint(
	ctx context.Context, f *SpanFilter,
) (map[string]time.Time, error) {
	var before, after time.Time

	if err := NewSpanIndexQuery(h.CH).
		ColumnExpr("max(time)").
		Where("s.project_id = ?", f.ProjectID).
		Where("s.time < ?", f.TimeGTE).
		Scan(ctx, &before); err != nil {
		return nil, err
	}

	if err := NewSpanIndexQuery(h.CH).
		ColumnExpr("min(time)").
		Where("s.project_id = ?", f.ProjectID).
		Where("s.time >= ?", f.TimeLT).
		Scan(ctx, &after); err != nil {
		return nil, err
	}

	dataHint := make(map[string]time.Time)

	if !before.IsZero() {
		dataHint["before"] = before
	}
	if !after.IsZero() {
		dataHint["after"] = after
	}

	return dataHint, nil
}
