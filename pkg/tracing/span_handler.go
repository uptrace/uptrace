package tracing

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/attrkey"
	"github.com/uptrace/uptrace/pkg/tracing/upql"
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

	if isAggAttr(f.SortBy) {
		f.SortBy = attrkey.SpanDuration
		f.SortDesc = true
	}

	q := buildSpanIndexQuery(h.App, f, f.Duration().Minutes()).
		ColumnExpr("_id").
		ColumnExpr("_trace_id").
		WithQuery(func(q *ch.SelectQuery) *ch.SelectQuery {
			if f.SortBy == "" {
				return q
			}
			return q.OrderExpr(string(CHAttrExpr(f.SortBy)) + " " + f.SortDir())
		}).
		Limit(10).
		Offset(f.Pager.GetOffset())

	spans := make([]*Span, 0)

	count, err := q.ScanAndCount(ctx, &spans)
	if err != nil {
		return err
	}

	var group syncutil.Group

	for i, span := range spans {
		i := i
		span := span

		group.Go(func() error {
			switch err := SelectSpan(ctx, h.App, span); err {
			case nil:
				return nil
			case sql.ErrNoRows:
				spans[i] = nil
				return nil
			default:
				return err
			}
		})
	}

	if err := group.Err(); err != nil {
		return err
	}

	for i := len(spans) - 1; i >= 0; i-- {
		if spans[i] == nil {
			spans = append(spans[:i], spans[i+1:]...)
		}
	}

	return httputil.JSON(w, bunrouter.H{
		"spans": spans,
		"count": count,
		"order": f.OrderByMixin,
	})
}

func (h *SpanHandler) ListGroups(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}

	q := buildSpanIndexQuery(h.App, f, f.Duration().Minutes()).
		Limit(1000)
	groups := make([]map[string]any, 0)

	if len(f.columnMap) > 0 {
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
	}

	columns := f.columns(groups)

	digest := xxhash.New()
	for _, group := range groups {
		for _, col := range columns {
			if col.IsGroup {
				digest.WriteString(fmt.Sprint(group[col.Name]))
			}
		}

		group[attrkey.ItemID] = strconv.FormatUint(digest.Sum64(), 10)
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

	groupPeriod := org.CalcGroupPeriod(f.TimeGTE, f.TimeLT, 300)
	minutes := groupPeriod.Minutes()

	m := make(map[string]interface{})

	subq := buildSpanIndexQuery(h.App, f, f.Duration().Minutes()).
		WithAlias("qsNaN", "quantilesTDigest(0.5, 0.9, 0.99)(_duration)").
		WithAlias("qs", "if(isNaN(qsNaN[1]), [0, 0, 0], qsNaN)").
		ColumnExpr("sum(_count) AS count").
		ColumnExpr("sum(_count) / ? AS rate", minutes).
		ColumnExpr("toStartOfInterval(_time, INTERVAL ? minute) AS time", minutes).
		WithQuery(func(q *ch.SelectQuery) *ch.SelectQuery {
			if isEventSystem(f.System) {
				return q
			}
			return q.ColumnExpr("sumIf(_count, _status_code = 'error') AS errorCount").
				ColumnExpr("sumIf(_count, _status_code = 'error') / ? AS errorRate",
					minutes).
				ColumnExpr("round(qs[1]) AS p50").
				ColumnExpr("round(qs[2]) AS p90").
				ColumnExpr("round(qs[3]) AS p99")
		}).
		WithQuery(f.whereClause).
		GroupExpr("time").
		OrderExpr("time ASC").
		Limit(10000)

	if err := h.CH.NewSelect().
		ColumnExpr("groupArray(count) AS count").
		ColumnExpr("groupArray(rate) AS rate").
		ColumnExpr("groupArray(time) AS time").
		WithQuery(func(q *ch.SelectQuery) *ch.SelectQuery {
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

	bunutil.FillHoles(m, f.TimeGTE, f.TimeLT, groupPeriod)

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

	colName, err := upql.ParseName(f.Column)
	if err != nil {
		return err
	}

	groupPeriod := org.CalcGroupPeriod(f.TimeGTE, f.TimeLT, 300)
	minutes := groupPeriod.Minutes()
	m := make(map[string]interface{})

	subq := buildSpanIndexQuery(h.App, f, minutes)
	subq = upqlColumn(subq, colName, minutes).
		ColumnExpr("toStartOfInterval(_time, toIntervalMinute(?)) AS time", minutes).
		GroupExpr("time").
		OrderExpr("time ASC")

	if err := h.CH.NewSelect().
		ColumnExpr("groupArray(?) AS ?", ch.Ident(f.Column), ch.Ident(f.Column)).
		ColumnExpr("groupArray(time) AS time").
		TableExpr("(?)", subq).
		GroupExpr("tuple()").
		Limit(1000).
		Scan(ctx, &m); err != nil {
		return err
	}

	bunutil.FillHoles(m, f.TimeGTE, f.TimeLT, groupPeriod)

	return httputil.JSON(w, m)
}
