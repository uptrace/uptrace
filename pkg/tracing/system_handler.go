package tracing

import (
	"context"
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
)

type SystemHandler struct {
	*bunapp.App
}

func NewSystemHandler(app *bunapp.App) *SystemHandler {
	return &SystemHandler{
		App: app,
	}
}

func (h *SystemHandler) ListSystems(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
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

	if err := NewSpanIndexQuery(h.App).
		ColumnExpr("s.project_id AS projectId").
		ColumnExpr("s.system").
		ColumnExpr("sum(s.count) AS count").
		ColumnExpr("sumIf(s.count, s.status_code = 'error') AS errorCount").
		ColumnExpr("sum(s.count) / ? AS rate", f.TimeFilter.Duration().Minutes()).
		ColumnExpr("sumIf(s.count, s.status_code = 'error') / sum(s.count) AS errorRate").
		ColumnExpr("uniqCombined64(s.group_id) AS groupCount").
		Apply(f.whereClause).
		Apply(f.spanqlWhere).
		GroupExpr("project_id, system").
		OrderExpr("system ASC").
		Limit(1000).
		Scan(ctx, &systems); err != nil {
		return nil, err
	}

	return systems, nil
}

func (h *SystemHandler) selectDataHint(
	ctx context.Context, f *SpanFilter,
) (map[string]time.Time, error) {
	var before, after time.Time

	if err := NewSpanIndexQuery(h.App).
		ColumnExpr("max(time)").
		Where("s.project_id = ?", f.ProjectID).
		Where("s.time < ?", f.TimeGTE).
		Scan(ctx, &before); err != nil {
		return nil, err
	}

	if err := NewSpanIndexQuery(h.App).
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
