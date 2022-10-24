package tracing

import (
	"context"
	"net/url"
	"strings"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type SystemFilter struct {
	org.TimeFilter

	ProjectID uint32
	System    string
	GroupID   uint64
	Envs      []string
	Services  []string

	prefixEnabled bool
}

func DecodeSystemFilter(app *bunapp.App, req bunrouter.Request) (*SystemFilter, error) {
	project := org.ProjectFromContext(req.Context())

	f := &SystemFilter{
		ProjectID: project.ID,
	}

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return nil, err
	}

	return f, nil
}

var _ urlstruct.ValuesUnmarshaler = (*SpanFilter)(nil)

func (f *SystemFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if err := f.TimeFilter.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	return nil
}

func (f *SystemFilter) whereClause(q *ch.SelectQuery) *ch.SelectQuery {
	q = q.Where("? = ?", f.prefix("project_id"), f.ProjectID).
		Where("? >= ?", f.prefix("time"), f.TimeGTE).
		Where("? < ?", f.prefix("time"), f.TimeLT)

	switch {
	case f.System == "":
		// nothing
	case f.System == SystemAllEvents:
		q = q.Where("?", f.prefix("is_event"))
	case f.System == SystemAllSpans:
		q = q.Where("NOT ?", f.prefix("is_event"))
	case f.System == SystemAll:
		q = q.Where("? != ?", f.prefix("system"), SystemInternalSpan)
	case strings.HasSuffix(f.System, ":all"):
		system := strings.TrimSuffix(f.System, ":all")
		q = q.Where("startsWith(?, ?)", f.prefix("system"), system)
	default:
		q = q.Where("? = ?", f.prefix("system"), f.System)
	}

	if f.GroupID != 0 {
		q = q.Where("? = ?", f.prefix("group_id"), f.GroupID)
	}

	if len(f.Envs) > 0 {
		q = q.Where("? IN (?)", f.prefix("deployment_environment"), ch.In(f.Envs))
	}
	if len(f.Services) > 0 {
		q = q.Where("? IN (?)", f.prefix("service"), ch.In(f.Services))
	}

	return q
}

func (f *SystemFilter) prefix(name string) ch.Ident {
	if f.prefixEnabled {
		return ch.Ident("_" + name)
	}
	return ch.Ident(name)
}
