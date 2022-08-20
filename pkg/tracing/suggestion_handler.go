package tracing

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/tracing/upql"

	"github.com/uptrace/uptrace/pkg/bunapp"
)

type Suggestions []Suggestion

func (ss *Suggestions) Add(sugg Suggestion) {
	*ss = append(*ss, sugg)
}

type Suggestion struct {
	Text string `json:"text"`
	Hint string `json:"hint,omitempty"`
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

type SuggestionHandler struct {
	*bunapp.App
}

func NewSuggestionHandler(app *bunapp.App) *SuggestionHandler {
	return &SuggestionHandler{
		App: app,
	}
}

func (h *SuggestionHandler) Attributes(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}
	disableColumnsAndGroups(f.parts)

	attrKeys, err := selectAttrKeys(ctx, h.App, f)
	if err != nil {
		return err
	}

	suggestions := make([]Suggestion, len(attrKeys))
	for i, key := range attrKeys {
		suggestions[i] = Suggestion{Text: key}
	}
	suggestions = sortSuggestions(suggestions)

	return httputil.JSON(w, bunrouter.H{
		"suggestions": suggestions,
	})
}

func selectAttrKeys(ctx context.Context, app *bunapp.App, f *SpanFilter) ([]string, error) {
	keys := make([]string, 0)
	if err := buildSpanIndexQuery(app, f, 0).
		ColumnExpr("groupUniqArrayArray(1000)(all_keys)").
		Scan(ctx, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

func (h *SuggestionHandler) Values(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}
	disableColumnsAndGroups(f.parts)

	if f.Column == "" {
		return fmt.Errorf(`"column" query param is required`)
	}
	colName, err := upql.ParseName(f.Column)
	if err != nil {
		return err
	}

	q := buildSpanIndexQuery(h.App, f, 0)
	q = upqlColumn(q, colName, 0).Group(f.Column)
	if !strings.HasPrefix(f.Column, "span.") {
		q = q.Where("has(all_keys, ?)", f.Column)
	}

	var items []map[string]interface{}
	if err := q.Scan(ctx, &items); err != nil {
		return err
	}

	suggestions := make([]Suggestion, len(items))
	for i, item := range items {
		suggestions[i] = Suggestion{
			Text: asString(item[f.Column]),
		}
	}

	return httputil.JSON(w, bunrouter.H{
		"suggestions": suggestions,
	})
}
