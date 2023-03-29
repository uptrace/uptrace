package tracing

import (
	"context"
	"net/http"

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

	systems, err := h.selectSystems(ctx, f)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"systems":   systems,
		"hasNoData": len(systems) == 0 && f.Query == "",
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
		ColumnExpr("sumIf(s.count, s.status_code = 'error') / sum(s.count) AS errorPct").
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
