package alerting

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type AlertFilter struct {
	ProjectID uint32 `urlstruct:"-"`

	org.FacetFilter
	org.OrderByMixin

	Status []string
	Type   []string

	MonitorID uint64
}

func DecodeAlertFilter(req bunrouter.Request, f *AlertFilter) error {
	project := org.ProjectFromContext(req.Context())
	f.ProjectID = project.ID

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return err
	}

	if err := extractParamsFromQuery(f); err != nil {
		return err
	}

	f.Status = f.Attrs[attrkey.AlertStatus]
	delete(f.Attrs, attrkey.AlertStatus)

	f.Type = f.Attrs[attrkey.AlertType]
	delete(f.Attrs, attrkey.AlertType)

	return nil
}

var _ urlstruct.ValuesUnmarshaler = (*AlertFilter)(nil)

func (f *AlertFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if err := f.FacetFilter.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	if err := f.OrderByMixin.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	return nil
}

func (f *AlertFilter) PGOrder(q *bun.SelectQuery) *bun.SelectQuery {
	if f.SortBy == "" {
		return q
	}

	var sortBy string

	switch f.SortBy {
	case "created_at", "createdAt":
		sortBy = "created_at"
	default:
		sortBy = "event.created_at"
	}

	return q.OrderExpr("? ? NULLS LAST", bun.Ident(sortBy), bun.Safe(f.SortDir()))
}

func extractParamsFromQuery(filter *AlertFilter) error {
	parts := strings.Split(filter.Q, " ")

	for i, part := range parts {
		ss := strings.Split(part, ":")

		if len(ss) != 2 {
			continue
		}

		switch ss[0] {
		case "monitor":
			monitorID, err := strconv.ParseUint(ss[1], 10, 64)
			if err != nil {
				return err
			}

			filter.MonitorID = monitorID

			parts = append(parts[:i], parts[i+1:]...)
		}
	}

	filter.Q = strings.Join(parts, " ")

	return nil
}

func (f *AlertFilter) WhereClause(q *bun.SelectQuery) *bun.SelectQuery {
	q = q.Where("a.project_id = ?", f.ProjectID).
		Apply(f.FacetFilter.WhereClause)

	if len(f.Type) > 0 {
		q = q.Where("a.type IN (?)", bun.In(f.Type))
	}
	if f.MonitorID != 0 {
		q = q.Where("a.monitor_id = ?", f.MonitorID)
	}
	if len(f.Status) > 0 {
		q = q.Where("event.status IN (?)", bun.In(f.Status))
	}

	return q
}

func (f *AlertFilter) Clone() *AlertFilter {
	clone := *f
	return &clone
}
