package tracing

import (
	"context"
	"net/http"
	"net/url"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
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

func (f *SystemFilter) whereClause(q *ch.SelectQuery) *ch.SelectQuery {
	return q.Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT)
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
		Apply(f.whereClause).
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

func (h *SystemHandler) Stats(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSystemFilter(h.App, req)
	if err != nil {
		return err
	}

	tableName, groupPeriod := spanSystemTableForGroup(&f.TimeFilter)

	subq := h.CH().NewSelect().
		WithAlias("tdigest", "quantilesTDigestMergeState(0.5, 0.9, 0.99)(s.tdigest)").
		WithAlias("qsNaN", "finalizeAggregation(tdigest)").
		WithAlias("qs", "if(isNaN(qsNaN[1]), [0, 0, 0], qsNaN)").
		ColumnExpr("system").
		ColumnExpr("sum(count) AS stats__count").
		ColumnExpr("stats__count / ? AS stats__rate", groupPeriod.Minutes()).
		ColumnExpr("sum(error_count) AS stats__errorCount").
		ColumnExpr("stats__errorCount / stats__count AS stats__errorPct").
		ColumnExpr("tdigest").
		ColumnExpr("qs[1] AS stats__p50").
		ColumnExpr("qs[2] AS stats__p90").
		ColumnExpr("qs[3] AS stats__p99").
		ColumnExpr("toStartOfInterval(time, INTERVAL ? minute) AS time", groupPeriod.Minutes()).
		TableExpr(tableName).
		Where("s.system != ?", internalSpanType).
		Apply(f.whereClause).
		GroupExpr("system, time").
		OrderExpr("system ASC, time ASC").
		Limit(10000)

	systems := make([]map[string]any, 0)

	if err := h.CH().NewSelect().
		WithAlias("qsNaN", "quantilesTDigestMerge(0.5, 0.9, 0.99)(tdigest)").
		WithAlias("qs", "if(isNaN(qsNaN[1]), [0, 0, 0], qsNaN)").
		ColumnExpr("system").
		ColumnExpr("sum(s.stats__count) AS count").
		ColumnExpr("count / ? AS rate", f.Duration().Minutes()).
		ColumnExpr("sum(s.stats__errorCount) AS errorCount").
		ColumnExpr("errorCount / count AS errorPct").
		ColumnExpr("qs[1] AS p50").
		ColumnExpr("qs[2] AS p90").
		ColumnExpr("qs[3] AS p99").
		ColumnExpr("groupArray(stats__count) AS stats__count").
		ColumnExpr("groupArray(stats__rate) AS stats__rate").
		ColumnExpr("groupArray(stats__errorCount) AS stats__errorCount").
		ColumnExpr("groupArray(stats__errorPct) AS stats__errorPct").
		ColumnExpr("groupArray(stats__p50) AS stats__p50").
		ColumnExpr("groupArray(stats__p90) AS stats__p90").
		ColumnExpr("groupArray(stats__p99) AS stats__p99").
		ColumnExpr("groupArray(time) AS stats__time").
		TableExpr("(?) AS s", subq).
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
