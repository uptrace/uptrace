package metrics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/uptrace/bunrouter"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type QueryFilter struct {
	App *bunapp.App

	ProjectID uint32
	org.TimeFilter
	org.OrderByMixin
	urlstruct.Pager

	Metrics   []string
	Aliases   []string
	Query     string
	BaseQuery string

	metricMap map[string]*Metric

	baseQueryParts []*upql.QueryPart
	queryParts     []*upql.QueryPart
	allParts       []*upql.QueryPart
}

func decodeQueryFilter(app *bunapp.App, req bunrouter.Request) (*QueryFilter, error) {
	ctx := req.Context()

	f := new(QueryFilter)
	f.App = app
	f.ProjectID = org.ProjectFromContext(ctx).ID

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return nil, err
	}

	return f, nil
}

var _ urlstruct.ValuesUnmarshaler = (*QueryFilter)(nil)

func (f *QueryFilter) UnmarshalValues(ctx context.Context, values url.Values) (err error) {
	if values != nil {
		if err := f.OrderByMixin.UnmarshalValues(ctx, values); err != nil {
			return err
		}
		if err := f.Pager.UnmarshalValues(ctx, values); err != nil {
			return err
		}
		if err := f.TimeFilter.UnmarshalValues(ctx, values); err != nil {
			return err
		}
	}

	if len(f.Metrics) == 0 {
		return errors.New("at least one metric id is required")
	}
	if len(f.Metrics) > 5 {
		return errors.New("at most 5 metric ids are allowed")
	}
	if len(f.Metrics) != len(f.Aliases) {
		return fmt.Errorf("got %d metrics and %d aliases", len(f.Metrics), len(f.Aliases))
	}

	f.metricMap = make(map[string]*Metric, len(f.Metrics))

	for i, metricName := range f.Metrics {
		if metricName == "" {
			return fmt.Errorf("metric name can't be empty")
		}

		metricAlias := f.Aliases[i]
		if metricAlias == "" {
			return fmt.Errorf("metric alias can't be empty")
		}

		metric, err := SelectMetricByName(ctx, f.App, f.ProjectID, metricName)
		if err != nil {
			if err == sql.ErrNoRows {
				f.metricMap[metricAlias] = newInvalidMetric(f.ProjectID, metricName)
				continue
			}
			return err
		}

		f.metricMap[metricAlias] = metric
	}

	f.baseQueryParts = upql.Parse(f.BaseQuery)
	f.queryParts = upql.Parse(f.Query)
	f.allParts = append(f.queryParts, f.baseQueryParts...)

	return nil
}

func (f *QueryFilter) Clone() *QueryFilter {
	clone := *f
	return &clone
}
