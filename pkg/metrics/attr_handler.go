package metrics

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/fx"
	"golang.org/x/exp/slices"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type AttrFilter struct {
	org.TimeFilter

	ProjectID uint32
	Metric    []string

	AttrKey         []string
	Instrument      []string
	OtelLibraryName []string
	SearchInput     string
}

func DecodeAttrFilter(req bunrouter.Request, f *AttrFilter) error {
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

func (f *AttrFilter) pgWhere(selq *bun.SelectQuery) *bun.SelectQuery {
	selq = selq.Where("project_id = ?", f.ProjectID)

	if len(f.Metric) == 0 {
		selq = selq.Where("updated_at >= ?", f.TimeGTE).
			Where("updated_at < ?", f.TimeLT)
	}
	if len(f.Instrument) > 0 {
		selq = selq.Where("instrument IN (?)", bun.In(f.Instrument))
	}
	if len(f.OtelLibraryName) > 0 {
		selq = selq.Where("otel_library_name IN (?)", bun.In(f.OtelLibraryName))
	}

	return selq
}

//------------------------------------------------------------------------------

type AttrHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	PG     *bun.DB
	CH     *ch.DB
}

type AttrHandler struct {
	*AttrHandlerParams
}

func NewAttrHandler(p AttrHandlerParams) *AttrHandler {
	return &AttrHandler{&p}
}

func registerAttrHandler(h *AttrHandler, p bunapp.RouterParams, m *Middleware) {
	p.RouterInternalV1.
		Use(m.UserAndProject).
		WithGroup("/metrics/:project_id", func(g *bunrouter.Group) {
			g.GET("/attributes", h.AttrKeys)
			g.GET("/attributes/:attr", h.AttrValues)
		})
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

	if err := DecodeAttrFilter(req, f); err != nil {
		return err
	}

	if len(f.Metric) == 0 {
		items := make([]AttrKeyItem, 0)

		subq := h.PG.NewSelect().
			Model((*Metric)(nil)).
			ColumnExpr("name AS metric").
			ColumnExpr("UNNEST(attr_keys) AS value").
			Apply(f.pgWhere)

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
		pinnedAttrMap, err = org.SelectPinnedFacetMap(ctx, h.PG, user.ID)
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

	slices.SortFunc(items, func(a, b *AttrKeyItem) int {
		return org.CompareAttrs(a.Value, b.Value)
	})

	return httputil.JSON(w, bunrouter.H{
		"items": items,
	})
}

func (h *AttrHandler) selectAttrKeys(ctx context.Context, f *AttrFilter) ([]string, error) {
	if len(f.Metric) == 1 {
		switch f.Metric[0] {
		case uptraceTracingSpans, uptraceTracingEvents, uptraceTracingLogs:
			typeFilter, err := newTypeFilter(ctx, f.ProjectID, &f.TimeFilter, f.Metric[0])
			if err != nil {
				return nil, err
			}
			spanFilter := newSpanFilter(typeFilter, "")

			keys, err := tracing.SelectAttrKeys(ctx, h.CH, spanFilter)
			if err != nil {
				return nil, err
			}

			return keys, nil
		}
	}

	keys := make([]string, 0)

	if err := h.PG.NewSelect().
		Model((*Metric)(nil)).
		ColumnExpr("UNNEST(array_intersect_agg(attr_keys))").
		Apply(f.pgWhere).
		Where("name IN (?)", bun.In(f.Metric)).
		Scan(ctx, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

func (h *AttrHandler) AttrValues(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	attrKey := req.Param("attr")

	f := new(AttrFilter)
	if err := DecodeAttrFilter(req, f); err != nil {
		return err
	}

	items, hasMore, err := h.selectAttrValues(ctx, attrKey, f)
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
	Count uint64 `json:"count"`
}

func (h *AttrHandler) selectAttrValues(
	ctx context.Context, attrKey string, f *AttrFilter,
) (any, bool, error) {
	const limit = 1000

	if len(f.Metric) == 1 {
		switch f.Metric[0] {
		case uptraceTracingSpans, uptraceTracingEvents, uptraceTracingLogs:
			typeFilter, err := newTypeFilter(ctx, f.ProjectID, &f.TimeFilter, f.Metric[0])
			if err != nil {
				return nil, false, err
			}

			spanFilter := newSpanFilter(typeFilter, "")
			return tracing.SelectAttrValues(ctx, h.CH, spanFilter, attrKey)
		}
	}

	tableName := DatapointTableForWhere(&f.TimeFilter)
	items := make([]AttrValueItem, 0)

	if err := h.CH.NewSelect().
		ColumnExpr("? AS value", chAttrExpr(attrKey)).
		ColumnExpr("count(DISTINCT metric) AS count").
		TableExpr("? AS d", ch.Name(tableName)).
		Where("d.project_id = ?", f.ProjectID).
		Where("d.time >= ?", f.TimeGTE).
		Where("d.time < ?", f.TimeLT).
		Apply(func(q *ch.SelectQuery) *ch.SelectQuery {
			if len(f.Metric) > 0 {
				q = q.Where("d.metric IN ?", ch.In(f.Metric))
			}
			if !isMandatoryAttr(attrKey) {
				q = q.Where("has(d.string_keys, ?)", attrKey)
			}

			for _, attrKey := range f.AttrKey {
				if isMandatoryAttr(attrKey) {
					continue
				}
				q = q.Where("has(d.string_keys, ?)", attrKey)
			}
			if len(f.Instrument) > 0 {
				q = q.Where("d.instrument IN ?", ch.In(f.Instrument))
			}
			if len(f.OtelLibraryName) > 0 {
				q = q.Where("d.otel_library_name IN ?", ch.In(f.OtelLibraryName))
			}

			if f.SearchInput != "" {
				q = q.Where("? like ?", chAttrExpr(attrKey), "%"+f.SearchInput+"%")
			}

			return q
		}).
		GroupExpr("value").
		OrderExpr("value ASC").
		Limit(limit).
		Scan(ctx, &items); err != nil {
		return nil, false, err
	}

	hasMore := f.SearchInput != "" || len(items) == limit
	return items, hasMore, nil
}

func chAttrExpr(attrKey string) ch.Safe {
	switch attrKey {
	case attrkey.MetricInstrument:
		return "d.instrument"
	}
	return ch.Safe(appendCHAttrExpr(nil, attrKey))
}

func appendCHAttrExpr(b []byte, attrKey string) []byte {
	return chschema.AppendQuery(b, "d.string_values[indexOf(string_keys, ?)]", attrKey)
}

func isMandatoryAttr(attrKey string) bool {
	switch attrKey {
	case attrkey.MetricInstrument, attrkey.OtelLibraryName:
		return true
	default:
		return false
	}
}
