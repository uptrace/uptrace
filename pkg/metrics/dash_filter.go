package metrics

import (
	"context"
	"net/url"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/pkg/urlstruct"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
)

type DashFilter struct {
	ProjectID uint32 `urlstruct:"-"`

	org.OrderByMixin

	Q      string
	Pinned bool
}

func DecodeDashFilter(req bunrouter.Request, f *DashFilter) error {
	project := org.ProjectFromContext(req.Context())
	f.ProjectID = project.ID

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return err
	}

	return nil
}

var _ urlstruct.ValuesUnmarshaler = (*DashFilter)(nil)

func (f *DashFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if err := f.OrderByMixin.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	return nil
}

func (f *DashFilter) PGOrder(q *bun.SelectQuery) *bun.SelectQuery {
	switch f.SortBy {
	case "updated_at":
		return q.OrderExpr("? ? NULLS LAST", bun.Ident(f.SortBy), bun.Safe(f.SortDir()))
	default:
		return q.OrderExpr("pinned DESC, lower(name) ?", bun.Safe(f.SortDir()))
	}
}

func (f *DashFilter) WhereClause(q *bun.SelectQuery) *bun.SelectQuery {
	q = q.Where("project_id = ?", f.ProjectID)

	if f.Q != "" {
		q = q.Where("word_similarity(?, name) >= 0.3", f.Q)
	}
	if f.Pinned {
		q = q.Where("pinned")
	}

	return q
}
