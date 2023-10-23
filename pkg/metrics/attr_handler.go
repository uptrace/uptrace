package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/urlstruct"
	"golang.org/x/exp/slices"
)

type AttrFilter struct {
	org.TimeFilter
	App *bunapp.App

	ProjectID uint32
	Metric    []string

	AttrKey     string
	SearchInput string
}

func DecodeAttrFilter(app *bunapp.App, req bunrouter.Request, f *AttrFilter) error {
	ctx := req.Context()
	f.ProjectID = org.ProjectFromContext(ctx).ID

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return err
	}

	return nil
}

var _ urlstruct.ValuesUnmarshaler = (*QueryFilter)(nil)

func (f *AttrFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if err := f.TimeFilter.UnmarshalValues(ctx, values); err != nil {
		return err
	}

	seen := make(map[string]bool, len(f.Metric))
	for i := len(f.Metric) - 1; i >= 0; i-- {
		metric := f.Metric[i]
		if seen[metric] {
			f.Metric = append(f.Metric[:i], f.Metric[i+1:]...)
		} else {
			seen[metric] = true
		}
	}

	return nil
}

//------------------------------------------------------------------------------

type AttrHandler struct {
	*bunapp.App
}

func NewAttrHandler(app *bunapp.App) *AttrHandler {
	return &AttrHandler{
		App: app,
	}
}

type AttrKeyItem struct {
	Value  string `json:"value"`
	Count  uint64 `json:"count"`
	Pinned bool   `json:"pinned"`
}

func (h *AttrHandler) AttrKeys(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	f := new(AttrFilter)
	f.TimeLT = time.Now()
	f.TimeGTE = time.Now().Add(-24 * time.Hour)

	if err := DecodeAttrFilter(h.App, req, f); err != nil {
		return err
	}

	if len(f.Metric) == 0 {
		items := make([]AttrKeyItem, 0)

		subq := h.PG.NewSelect().
			Model((*Metric)(nil)).
			ColumnExpr("name AS metric").
			ColumnExpr("UNNEST(attr_keys) AS value").
			Where("project_id = ?", f.ProjectID)

		if err := h.PG.NewSelect().
			ColumnExpr("value").
			ColumnExpr("count(DISTINCT metric) AS count").
			TableExpr("(?) AS items", subq).
			GroupExpr("value").
			OrderExpr("count DESC").
			Limit(10000).
			Scan(ctx, &items); err != nil {
			return err
		}

		return httputil.JSON(w, bunrouter.H{
			"items": items,
		})
	}

	attrKeys, err := h.selectAttrKeys(ctx, f)
	if err != nil {
		return err
	}

	var pinnedAttrMap map[string]bool
	if user != nil {
		pinnedAttrMap, err = org.SelectPinnedFacetMap(ctx, h.App, user.ID)
		if err != nil {
			return err
		}
	}

	items := make([]*AttrKeyItem, len(attrKeys))

	for i, attrKey := range attrKeys {
		items[i] = &AttrKeyItem{
			Value:  attrKey,
			Pinned: pinnedAttrMap[attrKey],
		}
	}

	slices.SortFunc(items, func(a, b *AttrKeyItem) bool {
		return org.CoreAttrLess(a.Value, b.Value)
	})

	return httputil.JSON(w, bunrouter.H{
		"items": items,
	})
}

func (h *AttrHandler) selectAttrKeys(ctx context.Context, f *AttrFilter) ([]string, error) {
	var keys []string

	if err := h.PG.NewSelect().
		Model((*Metric)(nil)).
		ColumnExpr("UNNEST(array_intersect_agg(attr_keys))").
		Where("project_id = ?", f.ProjectID).
		Where("name IN (?)", bun.In(f.Metric)).
		Scan(ctx, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

func (h *AttrHandler) AttrValues(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := new(AttrFilter)
	if err := DecodeAttrFilter(h.App, req, f); err != nil {
		return err
	}

	if len(f.Metric) == 0 {
		return errors.New(`"metric" query param is required`)
	}
	if f.AttrKey == "" {
		return fmt.Errorf(`"attr_key" query param is required`)
	}

	items, hasMore, err := h.selectAttrValues(ctx, f)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"items":   items,
		"hasMore": hasMore,
	})
}

type AttrValueItem struct {
	Value string `json:"value"`
}

func (h *AttrHandler) selectAttrValues(
	ctx context.Context, f *AttrFilter,
) (any, bool, error) {
	const limit = 1000

	tableName := datapointTableForWhere(h.App, &f.TimeFilter)

	items := make([]AttrValueItem, 0)

	if err := h.CH.NewSelect().
		ColumnExpr("DISTINCT string_values[indexOf(string_keys, ?)] AS value", f.AttrKey).
		TableExpr("?", tableName).
		Where("project_id = ?", f.ProjectID).
		Where("metric IN ?", ch.In(f.Metric)).
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT).
		Where("has(string_keys, ?)", f.AttrKey).
		Apply(func(q *ch.SelectQuery) *ch.SelectQuery {
			if f.SearchInput != "" {
				q = q.Where("string_values[indexOf(string_keys, ?)] like ?",
					f.AttrKey, "%"+f.SearchInput+"%")
			}
			return q
		}).
		OrderExpr("value ASC").
		Limit(limit).
		Scan(ctx, &items); err != nil {
		return nil, false, err
	}

	hasMore := f.SearchInput != "" || len(items) == limit
	return items, hasMore, nil
}
