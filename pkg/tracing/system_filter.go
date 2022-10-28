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

	switch {
	case f.System == "":
		// nothing
	case f.System == SystemAll:
		// nothing
	case f.System == SystemAllEvents:
		q = q.Where("s.system IN ('other-events', 'log', 'exceptions', 'message')")
	case f.System == SystemAllSpans:
		q = q.Where("s.system NOT IN ('other-events', 'log', 'exceptions', 'message')")
	case strings.HasSuffix(f.System, ":all"):
		system := strings.TrimSuffix(f.System, ":all")
		q = q.Where("startsWith(s.system, ?)", system)
	default:
		q = q.Where("s.system = ?", f.System)
	}

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
