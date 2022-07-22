package metrics

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type PromFilter struct {
	org.TimeFilter
	ProjectID uint32

	// Range query.
	Start time.Time
	End   time.Time

	// Instant query.
	Time time.Time

	// Group interval.
	Step time.Duration

	Label string
	Query string
}

func decodePromFilter(app *bunapp.App, req bunrouter.Request) (*PromFilter, error) {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	f := new(PromFilter)
	f.ProjectID = project.ID

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return nil, err
	}

	return f, nil
}

var _ urlstruct.ValuesUnmarshaler = (*PromFilter)(nil)

func (f *PromFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if f.End.IsZero() {
		f.End = time.Now()
	}
	if f.Start.IsZero() {
		f.Start = f.End.Add(-6 * time.Hour)
	}
	if f.End.Before(f.Start) {
		return errors.New("end time must not be before start time")
	}

	// Set these fields so we can use TimeFilter.
	f.TimeGTE = f.Start
	f.TimeLT = f.End

	return nil
}
