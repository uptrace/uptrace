package tracing

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
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

	slices.SortFunc(items, func(a, b *AttrKeyItem) int {
		return org.CompareAttrs(a.Value, b.Value)
	})

	return httputil.JSON(w, bunrouter.H{
		"items": items,
	})
}

func (h *AttrHandler) selectAttrKeys(ctx context.Context, f *SpanFilter) ([]string, error) {
	keys := make([]string, 0)
	q, _ := buildSpanIndexQuery(h.App, f, 0)
	if err := q.
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

	col, err := tql.ParseColumn(f.AttrKey)
	if err != nil {
		return err
	}

	attr, ok := col.Value.(tql.Attr)
	if !ok {
		return fmt.Errorf("expected an attr, got %T", col.Value)
	}

	for _, part := range f.parts {
		ast, ok := part.AST.(*tql.Where)
		if !ok {
			continue
		}

		for i := len(ast.Filters) - 1; i >= 0; i-- {
			filter := &ast.Filters[i]
			if tql.String(filter.LHS) == attr.Name {
				ast.Filters = append(ast.Filters[:i], ast.Filters[i+1:]...)
			}
		}
	}

	q, _ := buildSpanIndexQuery(h.App, f, 0)
	chExpr := appendCHAttr(nil, attr)

	q = q.ColumnExpr("? AS value", ch.Safe(chExpr)).
		GroupExpr("value").
		ColumnExpr("count() AS count")
	if !strings.HasPrefix(f.AttrKey, ".") {
		q = q.Where("has(s.all_keys, ?)", attr.Name)
	}
	if f.SearchInput != "" {
		q = q.Where("? like ?", chExpr, "%"+f.SearchInput+"%")
	}

	var rows []map[string]interface{}

	if err := q.Scan(ctx, &rows); err != nil {
		return err
	}

	items := make([]*AttrValueItem, len(rows))

	for i, item := range rows {
		items[i] = &AttrValueItem{
			Value: asString(item["value"]),
			Count: item["count"].(uint64),
		}
	}

	return httputil.JSON(w, bunrouter.H{
		"items": items,
	})
}
