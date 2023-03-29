package tracing

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/upql"
	"golang.org/x/exp/slices"

	"github.com/uptrace/uptrace/pkg/bunapp"
)

type AttrHandler struct {
	*bunapp.App
}

func NewAttrHandler(app *bunapp.App) *AttrHandler {
	return &AttrHandler{
		App: app,
	}
}

var spanKeys = []string{
	attrkey.SpanSystem,
	attrkey.SpanKind,
	attrkey.SpanName,
	attrkey.SpanEventName,
	attrkey.SpanStatusCode,
	attrkey.SpanStatusMessage,
}

type AttrKeyItem struct {
	Value  string `json:"value"`
	Pinned bool   `json:"pinned"`
}

func (h *AttrHandler) AttrKeys(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}
	disableColumnsAndGroups(f.parts)

	attrKeys, err := h.selectAttrKeys(ctx, f)
	if err != nil {
		return err
	}
	attrKeys = append(attrKeys, spanKeys...)

	pinnedAttrMap, err := org.SelectPinnedFacetMap(ctx, h.App, user.ID)
	if err != nil {
		return err
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

func (h *AttrHandler) selectAttrKeys(ctx context.Context, f *SpanFilter) ([]string, error) {
	keys := make([]string, 0)
	if err := buildSpanIndexQuery(h.App, f, 0).
		ColumnExpr("groupUniqArrayArray(1000)(s.all_keys)").
		Scan(ctx, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

type AttrValueItem struct {
	Value string `json:"value"`
	Count uint64 `json:"count"`
	Hint  string `json:"hint"`
}

func (h *AttrHandler) AttrValues(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}
	disableColumnsAndGroups(f.parts)

	if f.AttrKey == "" {
		return fmt.Errorf(`"attr_key" query param is required`)
	}
	f.AttrKey = attrkey.Clean(f.AttrKey)

	colName, err := upql.ParseName(f.AttrKey)
	if err != nil {
		return err
	}

	for _, part := range f.parts {
		ast, ok := part.AST.(*upql.Where)
		if !ok {
			continue
		}

		for i := len(ast.Conds) - 1; i >= 0; i-- {
			cond := &ast.Conds[i]
			if cond.Left == colName {
				ast.Conds = append(ast.Conds[:i], ast.Conds[i+1:]...)
			}
		}
	}

	q := buildSpanIndexQuery(h.App, f, 0)
	q = upqlColumn(q, colName, 0).Group(f.AttrKey).
		ColumnExpr("count() AS count")
	if !strings.HasPrefix(f.AttrKey, "span.") {
		q = q.Where("has(s.all_keys, ?)", f.AttrKey)
	}
	if f.SearchInput != "" {
		q = q.Where("? like ?", CHAttrExpr(f.AttrKey), "%"+f.SearchInput+"%")
	}

	var rows []map[string]interface{}

	if err := q.Scan(ctx, &rows); err != nil {
		return err
	}

	items := make([]*AttrValueItem, len(rows))

	for i, item := range rows {
		items[i] = &AttrValueItem{
			Value: asString(item[f.AttrKey]),
			Count: item["count"].(uint64),
		}
	}

	return httputil.JSON(w, bunrouter.H{
		"items": items,
	})
}
