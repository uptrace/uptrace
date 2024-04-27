package tracing

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
)

type GroupHandler struct {
	*bunapp.App
}

func NewGroupHandler(app *bunapp.App) *GroupHandler {
	return &GroupHandler{
		App: app,
	}
}

func (h *GroupHandler) ShowSummary(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f := &SpanFilter{}
	if err := DecodeSpanFilter(req, f); err != nil {
		return err
	}

	if f.GroupID == 0 {
		return errors.New("group_id is required")
	}

	parts := []string{
		attrkey.SpanCountSum,
		attrkey.SpanCountPerMin,
	}
	if !f.isEventSystem() {
		parts = append(parts,
			attrkey.SpanErrorCountSum,
			fmt.Sprintf("{p50, p90, p99, max}(%s)", attrkey.SpanDuration),
		)
	}
	f.QueryParts = tql.ParseQuery(strings.Join(parts, " | "))
	q, _ := BuildSpanIndexQuery(h.App.CH, f, f.TimeFilter.Duration())

	summary := make(map[string]any)
	if err := q.Apply(f.CHOrder).Scan(ctx, &summary); err != nil {
		return err
	}

	var firstSeenAt, lastSeenAt time.Time
	if err := NewSpanIndexQuery(h.App.CH).
		ColumnExpr("min(time) as first_seen_at").
		ColumnExpr("max(time) as last_seen_at").
		Where("project_id = ?", f.ProjectID).
		Apply(f.systemFilter).
		Where("group_id = ?", f.GroupID).
		Scan(ctx, &firstSeenAt, &lastSeenAt); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"summary":     summary,
		"firstSeenAt": firstSeenAt,
		"lastSeenAt":  lastSeenAt,
	})
}
