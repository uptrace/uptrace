package tracing

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
	"go4.org/syncutil"
)

type LogHandler struct {
	*bunapp.App
}

func NewLogHandler(app *bunapp.App) *LogHandler {
	return &LogHandler{
		App: app,
	}
}

type LogIdentity struct {
	ProjectID uint32
	TraceID   idgen.TraceID
	ID        idgen.SpanID
	Time      time.Time
}

func (h *LogHandler) ListLogs(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := &LogFilter{}
	if err := DecodeLogFilter(req, f); err != nil {
		return err
	}

	disableColumnsAndGroupsforLogs(f.QueryParts)

	if f.SortBy != "" && isAggExpr(tql.Attr{Name: f.SortBy}) {
		f.OrderByMixin.Reset()
	}

	q, _ := BuildLogIndexQuery(h.App.CH, f, f.TimeFilter.Duration())
	q = q.
		ColumnExpr("project_id, trace_id, id").
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

	ids := make([]LogIdentity, 0)
	count, err := q.ScanAndCount(ctx, &ids)
	if err != nil {
		return err
	}

	logs := make([]*Span, len(ids))
	var group syncutil.Group

	for i := range ids {
		id := &ids[i]
		idx := i

		group.Go(func() error {
			log, err := SelectLog(ctx, h.App, id.ProjectID, id.TraceID, id.ID)
			if err != nil {
				if err == sql.ErrNoRows {
					return nil
				}
				return err
			}

			logs[idx] = log
			return nil
		})
	}
	if err := group.Err(); err != nil {
		return err
	}
	for i := len(logs) - 1; i >= 0; i-- {
		if logs[i] == nil {
			logs = append(logs[:i], logs[i+1:]...)
		}
	}

	return httputil.JSON(w, bunrouter.H{
		"logs":  logs,
		"count": count,
		"order": f.OrderByMixin,
		"query": map[string]any{
			"parts": f.QueryParts,
		},
	})
}
