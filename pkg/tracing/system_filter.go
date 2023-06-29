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

	if f.GroupID != 0 {
		q = q.Where("s.group_id = ?", f.GroupID)
	}

	return q.WhereGroup(" AND ", func(q *ch.SelectQuery) *ch.SelectQuery {
		for _, system := range f.System {
			switch system {
			case "", SystemAll:
				// nothing
			case SystemSpansAll:
				q = q.WhereOr("s.type NOT IN ?", ch.In(LogAndEventTypes))
			case SystemEventsAll:
				q = q.WhereOr("s.type IN ?", ch.In(EventTypes))
			default:
				systemType, systemName := SplitTypeSystem(system)
				if systemName == SystemAll || systemName == systemType {
					q = q.WhereOr("s.type = ?", systemType)
				} else {
					q = q.WhereOr("s.type = ? AND s.system = ?", systemType, systemName)
				}
			}
		}
		return q
	})
}

func (f *SystemFilter) isEventSystem() bool {
	for _, system := range f.System {
		if !isEventSystem(system) {
			return false
		}
	}
	return true
}

func SplitTypeSystem(s string) (string, string) {
	if i := strings.IndexByte(s, ':'); i >= 0 {
		if s[i+1:] == SystemAll {
			return s[:i], SystemAll
		}
		return s[:i], s
	}
	return s, s
}
