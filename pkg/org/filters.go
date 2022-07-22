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
	SortBy  string
	SortDir string
}

var _ json.Marshaler = (*OrderByMixin)(nil)

func (f OrderByMixin) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"column": f.SortBy,
		"desc":   f.SortDir == "desc",
	})
}

var _ urlstruct.ValuesUnmarshaler = (*OrderByMixin)(nil)

func (f *OrderByMixin) UnmarshalValues(ctx context.Context, values url.Values) error {
	if f.SortDir == "" {
		f.SortDir = "desc"
	}
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
	groupPeriod = CalcGroupPeriod(f, 200)
	if groupPeriod >= time.Hour {
		tablePeriod = time.Hour
	} else {
		tablePeriod = time.Minute
	}
	return tablePeriod, groupPeriod
}

func CalcGroupPeriod(f *TimeFilter, n int) time.Duration {
	d := f.TimeLT.Sub(f.TimeGTE)
	period := time.Minute
	for i := 0; i < 100; i++ {
		if int(d/period) <= n {
			return period
		}
		period *= 2
	}
	return 24 * time.Hour
}
