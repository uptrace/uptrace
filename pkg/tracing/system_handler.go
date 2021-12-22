package tracing

import (
	"context"
	"net/http"
	"net/url"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type SystemFilter struct {
	*bunapp.App `urlstruct:"-"`

	TimeFilter
}

func DecodeSystemFilter(app *bunapp.App, req bunrouter.Request) (*SystemFilter, error) {
	f := &SystemFilter{App: app}
	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return nil, err
	}
	return f, nil
}

var _ urlstruct.ValuesUnmarshaler = (*SystemFilter)(nil)

func (f *SystemFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if err := f.TimeFilter.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	return nil
}

//------------------------------------------------------------------------------

type SystemHandler struct {
	*bunapp.App
}

func NewSystemHandler(app *bunapp.App) *SystemHandler {
	return &SystemHandler{
		App: app,
	}
}

func (h *SystemHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSystemFilter(h.App, req)
	if err != nil {
		return err
	}

	tableName := spanSystemTableForWhere(&f.TimeFilter)
	systems := make([]map[string]any, 0)

	if err := h.CH().NewSelect().
		TableExpr(tableName).
		ColumnExpr("system").
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT).
		GroupExpr("system").
		OrderExpr("system ASC").
		Limit(1000).
		Scan(ctx, &systems); err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"systems": systems,
	})
}
