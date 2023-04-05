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
	org.TimeFilter
	org.OrderByMixin
	urlstruct.Pager

	Project *org.Project `urlstruct:"-"`

	Metric []string
	Alias  []string
	Query  string

	parsedQuery *upql.ParsedQuery
	allParts    []*upql.QueryPart
}

func DecodeQueryFilter(req bunrouter.Request, f *QueryFilter) error {
	ctx := req.Context()
	f.Project = org.ProjectFromContext(ctx)

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return err
	}

	return nil
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

	if len(f.Metric) == 0 {
		return errors.New("at least one metric is required")
	}
	if len(f.Metric) > 6 {
		return errors.New("at most 6 metrics are allowed")
	}
	if len(f.Metric) != len(f.Alias) {
		return fmt.Errorf("got %d metrics and %d aliases", len(f.Metric), len(f.Alias))
	}

	f.parsedQuery = upql.Parse(f.Query)
	f.allParts = f.parsedQuery.Parts

	return nil
}

func (f *QueryFilter) Clone() *QueryFilter {
	clone := *f
	return &clone
}

func (f *QueryFilter) MetricMap(ctx context.Context, app *bunapp.App) (map[string]*Metric, error) {
	metricMap := make(map[string]*Metric, len(f.Metric))

	for i, metricName := range f.Metric {
		if metricName == "" {
			return nil, fmt.Errorf("metric name can't be empty")
		}

		metricAlias := f.Alias[i]
		if metricAlias == "" {
			return nil, fmt.Errorf("metric alias can't be empty")
		}

		metric, err := SelectMetricByName(ctx, app, f.Project.ID, metricName)
		if err != nil {
			if err == sql.ErrNoRows {
				metricMap[metricAlias] = newDeletedMetric(f.Project.ID, metricName)
				continue
			}
			return nil, err
		}

		metricMap[metricAlias] = metric
	}

	return metricMap, nil
}
