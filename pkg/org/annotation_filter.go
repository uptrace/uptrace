package org

import (
	"context"
	"net/url"
	"time"

	"github.com/codemodus/kace"
	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/pkg/urlstruct"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

type AnnotationFilter struct {
	ProjectID uint32 `urlstruct:"-"`
	TimeFilter
	urlstruct.Pager
	OrderByMixin
}

func decodeAnnotationFilter(req bunrouter.Request, f *AnnotationFilter) error {
	project := ProjectFromContext(req.Context())
	f.ProjectID = project.ID

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return err
	}

	if f.SortBy != "" {
		f.SortBy = kace.Snake(f.SortBy)
	}

	return nil
}

var _ urlstruct.ValuesUnmarshaler = (*AnnotationFilter)(nil)

func (f *AnnotationFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if f.TimeGTE.IsZero() && f.TimeLT.IsZero() {
		f.TimeLT = time.Now()
		f.TimeGTE = f.TimeLT.Add(-30 * 24 * time.Hour)
	}

	if err := f.Pager.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	if err := f.TimeFilter.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	if err := f.OrderByMixin.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	return nil
}

func (f *AnnotationFilter) WhereClause(q *bun.SelectQuery) *bun.SelectQuery {
	return q.Where("project_id = ?", f.ProjectID).
		Where("created_at >= ?", f.TimeGTE).
		Where("created_at < ?", f.TimeLT)
}
