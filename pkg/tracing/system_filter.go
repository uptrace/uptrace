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
}

func DecodeSystemFilter(app *bunapp.App, req bunrouter.Request) (*SystemFilter, error) {
	f := new(SystemFilter)

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return nil, err
	}

	f.ProjectID = org.ProjectFromContext(req.Context()).ID

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
	q = q.Where("project_id = ?", f.ProjectID).
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT)

	switch {
	case f.System == "":
		// nothing
	case f.System == SystemAllEvents:
		q = q.Where("is_event")
	case f.System == SystemAllSpans:
		q = q.Where("NOT is_event")
	case f.System == SystemAll:
		q = q.Where("system != ?", SystemInternalSpan)
	case strings.HasSuffix(f.System, ":all"):
		system := strings.TrimSuffix(f.System, ":all")
		q = q.Where("startsWith(system, ?)", system)
	default:
		q = q.Where("system = ?", f.System)
	}

	if f.GroupID != 0 {
		q = q.Where("group_id = ?", f.GroupID)
	}

	return q
}
