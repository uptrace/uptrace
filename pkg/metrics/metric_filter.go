package metrics

import (
	"context"
	"net/url"

	"github.com/uptrace/bunrouter"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type MetricFilter struct {
	App *bunapp.App

	ProjectID uint32
	org.TimeFilter

	Alias string
}

func DecodeMetricFilter(app *bunapp.App, req bunrouter.Request) (*MetricFilter, error) {
	ctx := req.Context()

	f := new(MetricFilter)
	f.App = app
	f.ProjectID = org.ProjectFromContext(ctx).ID

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return nil, err
	}

	return f, nil
}

var _ urlstruct.ValuesUnmarshaler = (*MetricFilter)(nil)

func (f *MetricFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if err := f.TimeFilter.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	return nil
}
