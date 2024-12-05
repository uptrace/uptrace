package tracing

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
	"golang.org/x/exp/slices"

	"github.com/uptrace/uptrace/pkg/bunapp"
)

type AttrHandler struct {
	logger *otelzap.Logger
	pg     *bun.DB
	ch     *ch.DB
}

func NewAttrHandler(logger *otelzap.Logger, pg *bun.DB, ch *ch.DB) *AttrHandler {
	return &AttrHandler{
		logger: logger,
		pg:     pg,
		ch:     ch,
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

	f := &SpanFilter{}
	if err := DecodeSpanFilter(req, f); err != nil {
		return err
	}
	disableColumnsAndGroups(f.QueryParts)

	fakeApp := &bunapp.App{CH: h.ch}
	attrKeys, err := SelectAttrKeys(ctx, fakeApp, f)
	if err != nil {
		return err
	}
	attrKeys = append(attrKeys, spanKeys...)

	pinnedAttrMap, err := org.SelectPinnedFacetMap(ctx, fakeApp, user.ID)
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

func SelectAttrKeys(ctx context.Context, app *bunapp.App, f *SpanFilter) ([]string, error) {
	keys := make([]string, 0)
	q, _ := BuildSpanIndexQuery(app.CH, f, 0)
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
	attrKey := req.Param("attr")

	f := &SpanFilter{}
	if err := DecodeSpanFilter(req, f); err != nil {
		return err
	}

	fakeApp := &bunapp.App{PG: h.pg}
	items, hasMore, err := SelectAttrValues(ctx, fakeApp, f, attrKey)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"items":   items,
		"hasMore": hasMore,
	})
}

func SelectAttrValues(
	ctx context.Context, app *bunapp.App, f *SpanFilter, attrKey string,
) ([]*AttrValueItem, bool, error) {
	const limit = 1000

	disableColumnsAndGroups(f.QueryParts)

	col, err := tql.ParseColumn(attrKey)
	if err != nil {
		return nil, false, err
	}

	attr, ok := col.Value.(tql.Attr)
	if !ok {
		return nil, false, fmt.Errorf("expected an attr, got %T", col.Value)
	}

	for _, part := range f.QueryParts {
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

	q, _ := BuildSpanIndexQuery(app.CH, f, 0)
	chExpr := appendCHAttr(nil, attr)

	q = q.ColumnExpr("? AS value", ch.Safe(chExpr)).
		GroupExpr("value").
		ColumnExpr("count() AS count")
	if !strings.HasPrefix(attrKey, "_") {
		q = q.Where("has(s.all_keys, ?)", attr.Name)
	}
	if f.SearchInput != "" {
		q = q.Where("? like ?", chExpr, "%"+f.SearchInput+"%")
	}

	items := make([]*AttrValueItem, 0)

	if err := q.Limit(limit).Scan(ctx, &items); err != nil {
		return nil, false, err
	}

	hasMore := f.SearchInput != "" || len(items) == limit
	return items, hasMore, nil
}
