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

	Envs     []string
	Services []string

	System  []string
	GroupID uint64
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
	q = q.Where("s.project_id = ?", f.ProjectID).
		Where("s.time >= ?", f.TimeGTE).
		Where("s.time < ?", f.TimeLT)

	return q.WhereGroup(" AND ", func(q *ch.SelectQuery) *ch.SelectQuery {
		for _, system := range f.System {
			switch {
			case system == "":
				// nothing
			case system == SystemAll:
				// nothing
			case system == SystemEventsAll:
				q = q.WhereOr("s.system IN (?)", ch.In(eventSystems))
			case system == SystemSpansAll:
				q = q.WhereOr("s.system NOT IN (?)", ch.In(eventSystems))
			case strings.HasSuffix(system, ":all"):
				system := strings.TrimSuffix(system, ":all")
				q = q.WhereOr("startsWith(s.system, ?)", system)
			default:
				q = q.WhereOr("s.system = ?", system)
			}
		}
		return q
	})

	if f.GroupID != 0 {
		q = q.Where("s.group_id = ?", f.GroupID)
	}

	if len(f.Envs) > 0 {
		q = q.Where("s.deployment_environment IN (?)", ch.In(f.Envs))
	}
	if len(f.Services) > 0 {
		q = q.Where("s.service IN (?)", ch.In(f.Services))
	}

	return q
}

func (f *SystemFilter) isEventSystem() bool {
	for _, system := range f.System {
		if !isEventSystem(system) {
			return false
		}
	}
	return true
}
