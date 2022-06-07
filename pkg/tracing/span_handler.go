package tracing

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/tracing/xattr"
	"github.com/uptrace/uptrace/pkg/uql"
	"go4.org/syncutil"
)

type SpanHandler struct {
	*bunapp.App
}

func NewSpanHandler(app *bunapp.App) *SpanHandler {
	return &SpanHandler{
		App: app,
	}
}

func (h *SpanHandler) ListSpans(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}
	disableColumnsAndGroups(f.parts)

	q := buildSpanIndexQuery(f, f.Duration().Minutes()).
		ColumnExpr("`span.id`").
		ColumnExpr("`span.trace_id`").
		Apply(func(q *ch.SelectQuery) *ch.SelectQuery {
			if f.SortBy == "" {
				return q
			}
			return q.OrderExpr(string(chColumn(f.SortBy)) + " " + f.SortDir)
		}).
		Limit(10).
		Offset(f.Pager.GetOffset())

	spans := make([]*Span, 0)

	count, err := q.ScanAndCount(ctx, &spans)
	if err != nil {
		return err
	}

	var group syncutil.Group

	for _, span := range spans {
		span := span
		group.Go(func() error {
			return SelectSpan(ctx, h.App, span)
		})
	}

	if err := group.Err(); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"spans": spans,
		"count": count,
		"order": f.OrderByMixin,
	})
}

func (h *SpanHandler) ListGroups(w http.ResponseWriter, req bunrouter.Request) error {
	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}

	ctx := req.Context()
	groups := make([]map[string]any, 0)

	q := buildSpanIndexQuery(f, f.Duration().Minutes()).
		Limit(1000)

	if err := q.Scan(ctx, &groups); err != nil {
		if cherr, ok := err.(*ch.Error); ok {
			w.WriteHeader(http.StatusBadRequest)
			return httputil.JSON(w, bunrouter.H{
				"query":   q.String(),
				"code":    "invalid_query",
				"message": cherr.Error(),
			})
		}
		return err
	}

	columns := f.columns(groups)

	digest := xxhash.New()
	for _, group := range groups {
		for _, col := range columns {
			if col.IsGroup {
				digest.WriteString(fmt.Sprint(group[col.Name]))
			}
		}

		group[xattr.ItemID] = strconv.FormatUint(digest.Sum64(), 10)
	}

	return httputil.JSON(w, bunrouter.H{
		"groups":     groups,
		"queryParts": f.parts,
		"columns":    columns,
	})
}

func (h *SpanHandler) Percentiles(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}

	groupPeriod := calcGroupPeriod(&f.TimeFilter, 300)
	minutes := groupPeriod.Minutes()

	m := make(map[string]interface{})

	subq := h.CH().NewSelect().
		Model((*SpanIndex)(nil)).
		WithAlias("qsNaN", "quantilesTDigest(0.5, 0.9, 0.99)(`span.duration`)").
		WithAlias("qs", "if(isNaN(qsNaN[1]), [0, 0, 0], qsNaN)").
		ColumnExpr("sum(`span.count`) AS count").
		ColumnExpr("sum(`span.count`) / ? AS rate", minutes).
		ColumnExpr("toStartOfInterval(`span.time`, INTERVAL ? minute) AS time", minutes).
		Apply(func(q *ch.SelectQuery) *ch.SelectQuery {
			if isEventSystem(f.System) {
				return q
			}
			return q.ColumnExpr("sumIf(`span.count`, `span.status_code` = 'error') AS errorCount").
				ColumnExpr("sumIf(`span.count`, `span.status_code` = 'error') / ? AS errorRate",
					minutes).
				ColumnExpr("round(qs[1]) AS p50").
				ColumnExpr("round(qs[2]) AS p90").
				ColumnExpr("round(qs[3]) AS p99")
		}).
		Apply(f.whereClause).
		GroupExpr("time").
		OrderExpr("time ASC").
		Limit(10000)

	if err := h.CH().NewSelect().
		ColumnExpr("groupArray(count) AS count").
		ColumnExpr("groupArray(rate) AS rate").
		ColumnExpr("groupArray(time) AS time").
		Apply(func(q *ch.SelectQuery) *ch.SelectQuery {
			if isEventSystem(f.System) {
				return q
			}
			return q.ColumnExpr("groupArray(errorCount) AS errorCount").
				ColumnExpr("groupArray(errorRate) AS errorRate").
				ColumnExpr("groupArray(p50) AS p50").
				ColumnExpr("groupArray(p90) AS p90").
				ColumnExpr("groupArray(p99) AS p99")
		}).
		TableExpr("(?)", subq).
		GroupExpr("tuple()").
		Limit(1000).
		Scan(ctx, &m); err != nil {
		return err
	}

	fillHoles(m, f.TimeGTE, f.TimeLT, groupPeriod)

	return httputil.JSON(w, m)
}

func (h *SpanHandler) Stats(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}
	disableColumnsAndGroups(f.parts)

	if f.Column == "" {
		return errors.New("'column' query param is required")
	}

	colName, err := uql.ParseName(f.Column)
	if err != nil {
		return err
	}

	groupPeriod := calcGroupPeriod(&f.TimeFilter, 300)
	minutes := groupPeriod.Minutes()
	m := make(map[string]interface{})

	subq := buildSpanIndexQuery(f, minutes)
	subq = uqlColumn(subq, colName, minutes).
		ColumnExpr("toStartOfInterval(`span.time`, toIntervalMinute(?)) AS time", minutes).
		GroupExpr("time").
		OrderExpr("time ASC")

	if err := h.CH().NewSelect().
		ColumnExpr("groupArray(?) AS ?", ch.Ident(f.Column), ch.Ident(f.Column)).
		ColumnExpr("groupArray(s.time) AS time").
		TableExpr("(?) AS s", subq).
		GroupExpr("tuple()").
		Limit(1000).
		Scan(ctx, &m); err != nil {
		return err
	}

	fillHoles(m, f.TimeGTE, f.TimeLT, groupPeriod)

	return httputil.JSON(w, m)
}
