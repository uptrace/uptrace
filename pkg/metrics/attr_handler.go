package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type AttrFilter struct {
	org.TimeFilter
	App *bunapp.App

	ProjectID uint32
	Metrics   []string
	Attr      string
}

func decodeAttrFilter(app *bunapp.App, req bunrouter.Request) (*AttrFilter, error) {
	ctx := req.Context()

	f := new(AttrFilter)
	f.App = app
	f.ProjectID = org.ProjectFromContext(ctx).ID

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return nil, err
	}

	return f, nil
}

var _ urlstruct.ValuesUnmarshaler = (*QueryFilter)(nil)

func (f *AttrFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if err := f.TimeFilter.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	return nil
}

//------------------------------------------------------------------------------

type AttrHandler struct {
	App *bunapp.App
}

func NewAttrHandler(app *bunapp.App) *AttrHandler {
	return &AttrHandler{
		App: app,
	}
}

func (h *AttrHandler) Keys(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodeAttrFilter(h.App, req)
	if err != nil {
		return err
	}

	if len(f.Metrics) == 0 {
		return errors.New("at least one metric is required")
	}

	attrKeys, err := h.selectAttrKeys(ctx, f)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"attrs": attrKeys,
	})
}

func (h *AttrHandler) selectAttrKeys(ctx context.Context, f *AttrFilter) ([]string, error) {
	keys := make([]string, 0)
	tableName := measureTableForWhere(h.App, &f.TimeFilter)

	if len(f.Metrics) == 0 {
		if err := h.App.CH.NewSelect().
			ColumnExpr("DISTINCT arrayJoin(attr_keys)").
			TableExpr("?", tableName).
			Where("project_id = ?", f.ProjectID).
			Where("time >= ?", f.TimeGTE).
			Where("time < ?", f.TimeLT).
			ScanColumns(ctx, &keys); err != nil {
			return nil, err
		}
		return keys, nil
	}

	q := h.App.CH.NewSelect().
		TableExpr("?", tableName).
		ColumnExpr("DISTINCT metric").
		ColumnExpr("arrayJoin(attr_keys) AS key").
		Where("project_id = ?", f.ProjectID).
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT).
		Where("metric IN (?)", ch.In(f.Metrics))

	if err := h.App.CH.NewSelect().
		ColumnExpr("key").
		TableExpr("(?) AS q", q).
		GroupExpr("key").
		Having("count() = ?", len(f.Metrics)).
		OrderExpr("key").
		ScanColumns(ctx, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

//------------------------------------------------------------------------------

func (h *AttrHandler) Where(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodeAttrFilter(h.App, req)
	if err != nil {
		return err
	}

	if len(f.Metrics) == 0 {
		return errors.New("at least one metric is required")
	}
	if f.Attr == "" {
		return fmt.Errorf("attribute query param is required")
	}

	where, err := h.selectWhereSuggestions(ctx, f)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"suggestions": where,
	})
}

func (h *AttrHandler) selectWhereSuggestions(
	ctx context.Context, f *AttrFilter,
) ([]WhereSuggestion, error) {
	var values []string

	tableName := measureTableForWhere(h.App, &f.TimeFilter)
	if err := f.App.CH.NewSelect().
		TableExpr("?", tableName).
		ColumnExpr("DISTINCT ? AS value", CHColumn(f.Attr)).
		Where("project_id = ?", f.ProjectID).
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT).
		Where("metric IN (?)", ch.In(f.Metrics)).
		Where("has(attr_keys, ?)", f.Attr).
		OrderExpr("value ASC").
		Limit(1000).
		ScanColumns(ctx, &values); err != nil {
		return nil, err
	}

	suggestions := make([]WhereSuggestion, 0, len(values))

	for _, value := range values {
		suggestions = append(suggestions, WhereSuggestion{
			Text:  fmt.Sprintf("%s = %q", f.Attr, value),
			Key:   f.Attr,
			Value: value,
		})
	}

	return suggestions, nil
}
