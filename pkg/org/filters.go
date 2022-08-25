package org

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type OrderByMixin struct {
	SortBy   string
	SortDesc bool
}

func (f OrderByMixin) SortDir() string {
	if f.SortDesc {
		return "desc"
	}
	return "asc"
}

var _ json.Marshaler = (*OrderByMixin)(nil)

func (f OrderByMixin) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"column": f.SortBy,
		"desc":   f.SortDesc,
	})
}

var _ urlstruct.ValuesUnmarshaler = (*OrderByMixin)(nil)

func (f *OrderByMixin) UnmarshalValues(ctx context.Context, values url.Values) error {
	return nil
}

//------------------------------------------------------------------------------

type TimeFilter struct {
	TimeGTE time.Time
	TimeLT  time.Time
}

var _ urlstruct.ValuesUnmarshaler = (*TimeFilter)(nil)

func (f *TimeFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if f.TimeGTE.IsZero() {
		return fmt.Errorf("time_gte is required")
	}
	if f.TimeLT.IsZero() {
		return fmt.Errorf("time_lt is required")
	}
	return nil
}

func (f *TimeFilter) Duration() time.Duration {
	return f.TimeLT.Sub(f.TimeGTE)
}

//------------------------------------------------------------------------------

func TablePeriod(f *TimeFilter) time.Duration {
	var period time.Duration

	if d := f.TimeLT.Sub(f.TimeGTE); d >= 6*time.Hour {
		period = time.Hour
	} else {
		period = time.Minute
	}

	return period
}

func TableGroupPeriod(f *TimeFilter) (tablePeriod, groupPeriod time.Duration) {
	groupPeriod = GroupPeriod(f.TimeGTE, f.TimeLT)
	if groupPeriod >= time.Hour {
		tablePeriod = time.Hour
	} else {
		tablePeriod = time.Minute
	}
	return tablePeriod, groupPeriod
}

func GroupPeriod(gte, lt time.Time) time.Duration {
	return CalcGroupPeriod(gte, lt, 120)
}

func CompactGroupPeriod(gte, lt time.Time) time.Duration {
	return CalcGroupPeriod(gte, lt, 60)
}

var periods = []time.Duration{
	time.Minute,
	2 * time.Minute,
	3 * time.Minute,
	5 * time.Minute,
	10 * time.Minute,
	15 * time.Minute,
	30 * time.Minute,
	time.Hour,
	2 * time.Hour,
	3 * time.Hour,
	4 * time.Hour,
	5 * time.Hour,
	6 * time.Hour,
	12 * time.Hour,
}

func CalcGroupPeriod(gte, lt time.Time, n int) time.Duration {
	d := lt.Sub(gte)
	for _, period := range periods {
		if int(d/period) <= n {
			return period
		}
	}
	return 24 * time.Hour
}
