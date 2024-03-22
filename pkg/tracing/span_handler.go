package tracing

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
	orderedmap "github.com/wk8/go-ordered-map/v2"
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

	f := &SpanFilter{}
	if err := DecodeSpanFilter(req, f); err != nil {
		return err
	}
	disableColumnsAndGroups(f.QueryParts)

	if f.SortBy != "" && isAggExpr(tql.Attr{Name: f.SortBy}) {
		f.OrderByMixin.Reset()
	}

	q, _ := BuildSpanIndexQuery(h.App.CH, f, f.TimeFilter.Duration())
	q = q.
		ColumnExpr("id, trace_id").
		Apply(func(q *ch.SelectQuery) *ch.SelectQuery {
			if f.SortBy == "" {
				f.SortBy = attrkey.SpanTime
				f.SortDesc = true
			}

			chExpr := appendCHAttr(nil, tql.Attr{Name: f.SortBy})
			order := string(chExpr) + " " + f.SortDir()
			return q.OrderExpr(order)
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
		"query": map[string]any{
			"parts": f.QueryParts,
		},
	})
}

func (h *SpanHandler) ListGroups(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := &SpanFilter{}
	if err := DecodeSpanFilter(req, f); err != nil {
		return err
	}

	q, columnMap := BuildSpanIndexQuery(h.App.CH, f, f.TimeFilter.Duration())

	if _, ok := columnMap.Load(f.SortBy); !ok {
		f.OrderByMixin.Reset()

		for pair := columnMap.Oldest(); pair != nil; pair = pair.Next() {
			col := pair.Value
			if !col.IsGroup {
				f.SortBy = pair.Key
				break
			}
		}

		if f.SortBy == "" {
			if pair := columnMap.Oldest(); pair != nil {
				f.SortBy = pair.Key
			}
		}
	}

	groups := make([]map[string]any, 0)
	if columnMap.Len() > 0 {
		if err := q.Apply(f.CHOrder).Limit(1000).Scan(ctx, &groups); err != nil {
			return err
		}
	}

	var grouping []string
	for pair := columnMap.Oldest(); pair != nil; pair = pair.Next() {
		col := pair.Value
		if col.IsGroup {
			grouping = append(grouping, col.Name)
		}
	}

	digest := xxhash.New()
	for _, group := range groups {
		digest.Reset()
		id, name, query := itemIDName(group, digest, grouping)
		group["_id"] = strconv.FormatUint(id, 10)
		group["_name"] = name
		group["_query"] = query
	}

	return httputil.JSON(w, bunrouter.H{
		"groups": groups,
		"order":  f.OrderByMixin,
		"query": map[string]any{
			"parts": f.QueryParts,
		},
		"columns": columnList(columnMap),
	})
}

func (h *SpanHandler) Percentiles(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := &SpanFilter{}
	if err := DecodeSpanFilter(req, f); err != nil {
		return err
	}

	groupingInterval := f.GroupingInterval()
	minutes := groupingInterval.Minutes()

	m := make(map[string]interface{})

	subq, _ := BuildSpanIndexQuery(h.App.CH, f, f.TimeFilter.Duration())
	subq = subq.
		ColumnExpr("sum(s.count) AS count").
		ColumnExpr("sum(s.count) / ? AS rate", minutes).
		ColumnExpr("toStartOfInterval(s.time, INTERVAL ? minute) AS time_", minutes).
		Apply(func(q *ch.SelectQuery) *ch.SelectQuery {
			if f.isEventSystem() {
				return q
			}
			return q.
				WithAlias("qsNaN", "quantilesTDigest(0.5, 0.9, 0.99)(s.duration)").
				WithAlias("qs", "if(isNaN(qsNaN[1]), [0, 0, 0], qsNaN)").
				ColumnExpr("sumIf(s.count, s.status_code = 'error') AS errorCount").
				ColumnExpr("sumIf(s.count, s.status_code = 'error') / ? AS errorRate",
					minutes).
				ColumnExpr("round(qs[1]) AS p50").
				ColumnExpr("round(qs[2]) AS p90").
				ColumnExpr("round(qs[3]) AS p99").
				ColumnExpr("max(duration) AS max")
		}).
		Apply(f.whereClause).
		GroupExpr("time_").
		OrderExpr("time_ ASC").
		Limit(10000)

	if err := h.CH.NewSelect().
		ColumnExpr("groupArray(count) AS count").
		ColumnExpr("groupArray(rate) AS rate").
		ColumnExpr("groupArray(time_) AS time").
		Apply(func(q *ch.SelectQuery) *ch.SelectQuery {
			if f.isEventSystem() {
				return q
			}
			return q.ColumnExpr("groupArray(errorCount) AS errorCount").
				ColumnExpr("groupArray(errorRate) AS errorRate").
				ColumnExpr("groupArray(p50) AS p50").
				ColumnExpr("groupArray(p90) AS p90").
				ColumnExpr("groupArray(p99) AS p99").
				ColumnExpr("groupArray(max) AS max")
		}).
		TableExpr("(?)", subq).
		GroupExpr("tuple()").
		Limit(1000).
		Scan(ctx, &m); err != nil {
		return err
	}

	bunutil.FillHoles(m, f.TimeGTE, f.TimeLT, groupingInterval)

	return httputil.JSON(w, map[string]any{
		"stats": m,
	})
}

func (h *SpanHandler) GroupStats(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := &SpanFilter{}
	if err := DecodeSpanFilter(req, f); err != nil {
		return err
	}
	disableColumnsAndGroups(f.QueryParts)

	if len(f.Column) == 0 {
		return errors.New(`"column" query param is required`)
	}
	f.Pager.Limit = 1000

	groupingInterval := f.GroupingInterval()

	subq, _ := BuildSpanIndexQuery(h.App.CH, f, groupingInterval)
	subq = subq.
		ColumnExpr("toStartOfInterval(time, toIntervalMinute(?)) AS time_", groupingInterval.Minutes()).
		GroupExpr("time_").
		OrderExpr("time_ ASC")

	for _, colName := range f.Column {
		col, err := tql.ParseColumn(colName)
		if err != nil {
			return err
		}
		chExpr, err := appendCHColumn(nil, col, groupingInterval)
		if err != nil {
			return err
		}
		subq = subq.ColumnExpr(string(chExpr))
	}

	item := make(map[string]interface{})

	if err := h.CH.NewSelect().
		Apply(func(q *ch.SelectQuery) *ch.SelectQuery {
			for _, colName := range f.Column {
				q = q.ColumnExpr("groupArray(?) AS ?", ch.Name(colName), ch.Name(colName))
			}
			return q
		}).
		ColumnExpr("groupArray(time_) AS _time").
		TableExpr("(?)", subq).
		GroupExpr("tuple()").
		Limit(1000).
		Scan(ctx, &item); err != nil {
		return err
	}

	bunutil.FillHoles(item, f.TimeGTE, f.TimeLT, groupingInterval)

	return httputil.JSON(w, item)
}

//------------------------------------------------------------------------------

func (h *SpanHandler) Timeseries(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := &SpanFilter{}
	if err := DecodeSpanFilter(req, f); err != nil {
		return err
	}

	groupingInterval := f.GroupingInterval()
	subq, columnMap := BuildSpanIndexQuery(h.App.CH, f, groupingInterval)

	var numAgg int
	for _, colName := range columnNames(columnMap) {
		col := columnMap.Value(colName)
		if !col.IsGroup && col.IsNum {
			if numAgg >= 4 {
				columnMap.Delete(colName)
			}
			numAgg++
		}
	}

	subq = subq.
		ColumnExpr(
			"toStartOfInterval(s.time, INTERVAL ? minute) AS `_time`",
			groupingInterval.Minutes(),
		).
		Group("_time").
		OrderExpr("`_time` ASC")

	q := h.CH.NewSelect().
		ColumnExpr("groupArray(`_time`) AS `_time`").
		TableExpr("(?)", subq).
		Limit(100)

	var grouping []string

	for pair := columnMap.Oldest(); pair != nil; pair = pair.Next() {
		colName := pair.Key
		col := pair.Value

		if col.IsGroup {
			q = q.Column(colName).Group(colName).OrderExpr("? ASC", ch.Name(colName))
			grouping = append(grouping, colName)

			if colName == attrkey.SpanGroupID {
				q = q.ColumnExpr("any(?) AS ?",
					ch.Name(attrkey.DisplayName), ch.Name(attrkey.DisplayName))
			}
		} else if col.IsNum {
			q = q.ColumnExpr("groupArray(?) AS ?", ch.Name(colName), ch.Name(colName))
		} else {
			q = q.ColumnExpr("any(?) AS ?", ch.Name(colName), ch.Name(colName))
		}
	}

	if len(grouping) == 0 {
		q = q.GroupExpr("tuple()")
	}

	groups := make([]map[string]any, 0)

	if err := q.Scan(ctx, &groups); err != nil {
		return err
	}

	var timeCol []time.Time

	digest := xxhash.New()
	for _, group := range groups {
		bunutil.FillHoles(group, f.TimeGTE, f.TimeLT, groupingInterval)

		if timeCol == nil {
			timeCol = group["_time"].([]time.Time)
		}
		delete(group, "_time")

		digest.Reset()
		id, name, query := itemIDName(group, digest, grouping)
		group["_id"] = strconv.FormatUint(id, 10)
		group["_name"] = name
		group["_query"] = query
	}

	return httputil.JSON(w, map[string]any{
		"groups":  groups,
		"time":    timeCol,
		"columns": columnList(columnMap),
		"query": map[string]any{
			"parts": f.QueryParts,
		},
	})
}

func itemIDName(
	item map[string]any, digest *xxhash.Digest, grouping []string,
) (uint64, string, string) {
	var names []string
	var filters []string

	for _, colName := range grouping {
		value := item[colName]
		digest.WriteString(fmt.Sprint(value))

		filters = append(filters, fmt.Sprintf("%s = %s", colName, quote(value)))

		if colName == attrkey.SpanGroupID {
			if s, ok := item[attrkey.DisplayName].(string); ok && s != "" {
				names = append(names, s)
			}
		} else {
			names = append(names, fmt.Sprintf("%s=%s", colName, quote(value)))
		}
	}

	var query string
	if len(filters) > 0 {
		query = "where " + strings.Join(filters, " and ")
	}

	return digest.Sum64(), strings.Join(names, " "), query
}

func quote(v any) string {
	if s, ok := v.(string); ok {
		return strconv.Quote(s)
	}
	return fmt.Sprint(v)
}

func columnNames(m *orderedmap.OrderedMap[string, *ColumnInfo]) []string {
	names := make([]string, 0, m.Len())
	for pair := m.Oldest(); pair != nil; pair = pair.Next() {
		names = append(names, pair.Key)
	}
	return names
}

func columnList(m *orderedmap.OrderedMap[string, *ColumnInfo]) []*ColumnInfo {
	columns := make([]*ColumnInfo, 0, m.Len())
	for pair := m.Oldest(); pair != nil; pair = pair.Next() {
		columns = append(columns, pair.Value)
	}
	return columns
}
