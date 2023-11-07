package tracing

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/uptrace/bunrouter"
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

type GroupTimeOut struct {
	FirstSeenAt time.Time `bun:",scanonly"`
	LastSeenAt  time.Time `bun:",scanonly"`
}

func (h *GroupHandler) ShowSummary(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}

	if f.GroupID == 0 {
		return errors.New("group_id is required")
	}

	parts := []string{
		".count",
		"per_min(.count)",
	}
	if !f.isEventSystem() {
		parts = append(parts,
			".error_count",
			"{p50, p90, p99, max}(.duration)",
		)
	}
	f.parts = tql.ParseQuery(strings.Join(parts, " | "))

	summary := make(map[string]any)
	q, _ := buildSpanIndexQuery(h.App, f, f.TimeFilter.Duration())
	if err = q.Apply(f.CHOrder).Scan(ctx, &summary); err != nil {
		return err
	}

	groupTimeOut := new(GroupTimeOut)
	if err = NewSpanIndexQuery(h.App).
		ColumnExpr("min(time) as first_seen_at").
		ColumnExpr("max(time) as last_seen_at").
		Where("group_id = ?", f.GroupID).
		Scan(ctx, groupTimeOut); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"firstSeenAt": groupTimeOut.FirstSeenAt,
		"lastSeenAt":  groupTimeOut.LastSeenAt,
		"summary":     summary,
	})
}
