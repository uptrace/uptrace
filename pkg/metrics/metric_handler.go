package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/uptrace/bunrouter"

	"github.com/uptrace/go-clickhouse/ch"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
)

type MetricHandler struct {
	*bunapp.App
}

func NewMetricHandler(app *bunapp.App) *MetricHandler {
	return &MetricHandler{
		App: app,
	}
}

func (h *MetricHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	metrics, err := SelectMetrics(ctx, h.App, project.ID)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"metrics": metrics,
	})
}

func (h *MetricHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	return httputil.JSON(w, bunrouter.H{
		"metric": metricFromContext(ctx),
	})
}

type Suggestions []Suggestion

func (ss *Suggestions) Add(sugg Suggestion) {
	*ss = append(*ss, sugg)
}

type Suggestion struct {
	Text string `json:"text"`
	Hint string `json:"hint,omitempty"`
	Kind string `json:"kind,omitempty"`
}

func sortSuggestions(suggestions []Suggestion) []Suggestion {
	seen := make(map[string]struct{}, len(suggestions))

	for i := len(suggestions) - 1; i >= 0; i-- {
		key := suggestions[i].Text
		if _, ok := seen[key]; ok {
			suggestions = append(suggestions[:i], suggestions[i+1:]...)
		} else {
			seen[key] = struct{}{}
		}
	}

	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Text < suggestions[j].Text
	})

	return suggestions
}

//------------------------------------------------------------------------------

type MetricKeys struct {
	Alias    string   `json:"alias"`
	Metric   *Metric  `json:"metric"`
	AttrKeys []string `json:"attrKeys"`
}

func (h *MetricHandler) Attributes(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeMetricFilter(h.App, req)
	if err != nil {
		return err
	}

	keys, err := selectAttrKeys(ctx, h.App, metricFromContext(ctx))
	if err != nil {
		return err
	}

	suggestions := make([]Suggestion, len(keys))

	for i, key := range keys {
		suggestions[i] = Suggestion{
			Text: fmt.Sprintf("$%s.%s", f.Alias, key),
		}
	}

	suggestions = sortSuggestions(suggestions)

	return httputil.JSON(w, bunrouter.H{
		"suggestions": suggestions,
	})
}

//------------------------------------------------------------------------------

func (h *MetricHandler) Where(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeMetricFilter(h.App, req)
	if err != nil {
		return err
	}

	suggestions, err := h.selectWhereSuggestions(ctx, f)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"suggestions": suggestions,
	})
}

type WhereSuggestion struct {
	Text  string `json:"text"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (h *MetricHandler) selectWhereSuggestions(
	ctx context.Context, f *MetricFilter,
) ([]*WhereSuggestion, error) {
	metric := metricFromContext(ctx)

	attrKeys, err := selectAttrKeys(ctx, h.App, metric)
	if err != nil {
		return nil, err
	}

	suggestions := make([]*WhereSuggestion, 0)

	if len(attrKeys) == 0 {
		return suggestions, nil
	}

	keys := make([]string, 0, len(attrKeys))
	vals := make([]ch.Safe, 0, len(attrKeys))

	for _, key := range attrKeys {
		keys = append(keys, key)
		vals = append(vals, CHColumn(key))
	}

	tableName := measureTableForWhere(h.App, &f.TimeFilter)
	if err := f.App.CH.NewSelect().
		TableExpr("? AS ?", tableName, ch.Ident(f.Alias)).
		ColumnExpr("DISTINCT key, value").
		Join("ARRAY JOIN [?] AS key, [?] AS value", ch.In(keys), ch.In(vals)).
		Where("project_id = ?", f.ProjectID).
		Where("metric = ?", metric.Name).
		Where("time >= ?", f.TimeGTE).
		Where("time < ?", f.TimeLT).
		Where("has(attr_keys, key)").
		OrderExpr("key ASC, value ASC").
		Scan(ctx, &suggestions); err != nil {
		return nil, err
	}

	for _, sugg := range suggestions {
		sugg.Key = fmt.Sprintf("$%s.%s", f.Alias, sugg.Key)
		sugg.Text = fmt.Sprintf("%s = %q", sugg.Key, sugg.Value)
	}

	return suggestions, nil
}

//------------------------------------------------------------------------------

const numKeyLimit = 1000

func selectAttrKeys(ctx context.Context, app *bunapp.App, metric *Metric) ([]string, error) {
	keys := make([]string, 0)
	if err := app.CH.NewSelect().
		ColumnExpr("arrayJoin(attr_keys) AS key").
		TableExpr("?", app.DistTable("measure_minutes_buffer")).
		Where("project_id = ?", metric.ProjectID).
		Where("metric = ?", metric.Name).
		Where("time >= ?", time.Now().Add(-time.Hour)).
		GroupExpr("key").
		OrderExpr("key ASC").
		Limit(numKeyLimit).
		ScanColumns(ctx, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}
