package tracing

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/fx"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/pkg/clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
)

type GroupHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	CH     *ch.DB
}

type GroupHandler struct {
	*GroupHandlerParams
}

func NewGroupHandler(p GroupHandlerParams) *GroupHandler {
	return &GroupHandler{&p}
}

func registerGroupHandler(h *GroupHandler, p bunapp.RouterParams, m *org.Middleware) {
	p.RouterInternalV1.Use(m.UserAndProject).
		WithGroup("/tracing/:project_id/groups/:group_id", func(g *bunrouter.Group) {
			g.GET("", h.ShowSummary)
		})
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
	q, _ := BuildSpanIndexQuery(h.CH, f, f.TimeFilter.Duration())

	summary := make(map[string]any)
	if err := q.Apply(f.CHOrder).Scan(ctx, &summary); err != nil {
		return err
	}

	var firstSeenAt, lastSeenAt time.Time
	if err := NewSpanIndexQuery(h.CH).
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
