package org

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/codemodus/kace"
	"github.com/segmentio/encoding/json"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/pgquery"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type OrderByMixin struct {
	SortBy   string
	SortDesc bool
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

func (f *OrderByMixin) PGOrder(q *bun.SelectQuery) *bun.SelectQuery {
	if f.SortBy == "" {
		return q
	}
	return q.Order(kace.Snake(f.SortBy) + " " + f.SortDir() + " NULLS LAST")
}

func (f OrderByMixin) SortDir() string {
	if f.SortDesc {
		return "desc"
	}
	return "asc"
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

type Facet struct {
	Key   string       `json:"key"`
	Items []*FacetItem `json:"items"`
}

type FacetItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Count uint32 `json:"count"`
}

func BuildFacetMap(items []*FacetItem) map[string]*Facet {
	m := make(map[string]*Facet)

	for _, item := range items {
		facet, ok := m[item.Key]
		if !ok {
			facet = &Facet{Key: item.Key}
			m[item.Key] = facet
		}
		facet.Items = append(facet.Items, item)
	}

	return m
}

func FacetMapToList(m map[string]*Facet) []*Facet {
	facets := maps.Values(m)
	slices.SortFunc(facets, func(a, b *Facet) bool {
		return a.Key < b.Key
	})
	return facets
}

type FacetFilter struct {
	Q     string
	Attrs map[string][]string
}

var _ urlstruct.ValuesUnmarshaler = (*FacetFilter)(nil)

func (f *FacetFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	f.Q = strings.TrimSpace(f.Q)
	return nil
}

func (f *FacetFilter) WhereClause(q *bun.SelectQuery) *bun.SelectQuery {
	if f.Q != "" {
		q = q.Where("tsv @@ websearch_to_tsquery('english', ?)", f.Q)
	}

	for key, values := range f.Attrs {
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			for _, value := range values {
				attr := pgquery.BuildAttr(pgquery.EscapeWord(key), pgquery.EscapeWord(value))
				q = q.WhereOr("tsv @@ ?::tsquery", attr)
			}
			return q
		})
	}

	return q
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

func TableGroupingPeriod(f *TimeFilter) (tablePeriod, groupingPeriod time.Duration) {
	groupingPeriod = GroupingPeriod(f.TimeGTE, f.TimeLT)
	if groupingPeriod >= time.Hour {
		tablePeriod = time.Hour
	} else {
		tablePeriod = time.Minute
	}
	return tablePeriod, groupingPeriod
}

func GroupingPeriod(gte, lt time.Time) time.Duration {
	return CalcGroupingPeriod(gte, lt, 120)
}

func CompactGroupingPeriod(gte, lt time.Time) time.Duration {
	return CalcGroupingPeriod(gte, lt, 60)
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

func CalcGroupingPeriod(gte, lt time.Time, n int) time.Duration {
	d := lt.Sub(gte)
	for _, period := range periods {
		if int(d/period) <= n {
			return period
		}
	}
	return 24 * time.Hour
}
