package tracing

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/tracing/attrkey"
	"github.com/uptrace/uptrace/pkg/tracing/upql"

	"github.com/uptrace/uptrace/pkg/bunapp"
)

type Suggestions []Suggestion

func (ss *Suggestions) Add(sugg Suggestion) {
	*ss = append(*ss, sugg)
}

type Suggestion struct {
	Text  string `json:"text"`
	Hint  string `json:"hint,omitempty"`
	Count uint64 `json:"count"`
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

var spanKeys = []string{
	attrkey.SpanSystem,
	attrkey.SpanKind,
	attrkey.SpanName,
	attrkey.SpanEventName,
	attrkey.SpanStatusCode,
	attrkey.SpanStatusMessage,
}

func (h *SuggestionHandler) Attributes(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeSpanFilter(h.App, req)
	if err != nil {
		return err
	}
	disableColumnsAndGroups(f.parts)

	attrKeys, err := h.selectAttrKeys(ctx, f)
	if err != nil {
		return err
	}
	attrKeys = append(attrKeys, spanKeys...)

	suggestions := make([]Suggestion, len(attrKeys))
	for i, key := range attrKeys {
		suggestions[i] = Suggestion{Text: key}
	}
	suggestions = sortSuggestions(suggestions)

	return httputil.JSON(w, bunrouter.H{
		"suggestions": suggestions,
	})
}

func (h *SuggestionHandler) selectAttrKeys(ctx context.Context, f *SpanFilter) ([]string, error) {
	keys := make([]string, 0)
	if err := buildSpanIndexQuery(h.App, f, 0).
		ColumnExpr("groupUniqArrayArray(1000)(s.all_keys)").
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

	if f.AttrKey == "" {
		return fmt.Errorf(`"attr_key" query param is required`)
	}
	f.AttrKey = attrkey.Clean(f.AttrKey)

	colName, err := upql.ParseName(f.AttrKey)
	if err != nil {
		return err
	}

	for _, part := range f.parts {
		ast, ok := part.AST.(*upql.Where)
		if !ok {
			continue
		}

		for i := len(ast.Conds) - 1; i >= 0; i-- {
			cond := &ast.Conds[i]
			if cond.Left == colName {
				ast.Conds = append(ast.Conds[:i], ast.Conds[i+1:]...)
			}
		}
	}

	q := buildSpanIndexQuery(h.App, f, 0)
	q = upqlColumn(q, colName, 0).Group(f.AttrKey).
		ColumnExpr("count() AS count")
	if !strings.HasPrefix(f.AttrKey, "span.") {
		q = q.Where("has(s.all_keys, ?)", f.AttrKey)
	}
	if f.AttrValue != "" {
		q = q.Where("? like ?", CHAttrExpr(f.AttrKey), "%"+f.AttrValue+"%")
	}

	var items []map[string]interface{}

	if err := q.Scan(ctx, &items); err != nil {
		return err
	}

	suggestions := make([]Suggestion, len(items))

	for i, item := range items {
		suggestions[i] = Suggestion{
			Text:  asString(item[f.AttrKey]),
			Count: item["count"].(uint64),
		}
	}

	return httputil.JSON(w, bunrouter.H{
		"suggestions": suggestions,
	})
}
