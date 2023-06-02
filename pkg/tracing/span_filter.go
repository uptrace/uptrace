package tracing

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bununit"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
	"github.com/uptrace/uptrace/pkg/urlstruct"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type SpanFilter struct {
	*bunapp.App `urlstruct:"-"`

	org.OrderByMixin
	urlstruct.Pager
	SystemFilter

	Query string

	// For stats explorer.
	Column []string

	// For attrs suggestions.
	AttrKey     string
	SearchInput string

	parts     []*tql.QueryPart
	columnMap map[string]bool
}

func DecodeSpanFilter(app *bunapp.App, req bunrouter.Request) (*SpanFilter, error) {
	f := &SpanFilter{App: app}

	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return nil, err
	}

	project := org.ProjectFromContext(req.Context())
	f.ProjectID = project.ID
	f.parts = tql.Parse(f.Query)

	return f, nil
}

var _ urlstruct.ValuesUnmarshaler = (*SpanFilter)(nil)

func (f *SpanFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if err := f.SystemFilter.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	if err := f.Pager.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	return nil
}

//------------------------------------------------------------------------------

type ColumnInfo struct {
	Name    string `json:"name"`
	Unit    string `json:"unit"`
	IsNum   bool   `json:"isNum"`
	IsGroup bool   `json:"isGroup"`
}

func isNumColumn(v any) bool {
	switch v.(type) {
	case int64, uint64, float32, float64,
		[]int64, []uint64, []float32, []float64:
		return true
	default:
		return false
	}
}

func unitFromName(name tql.Name) string {
	var unit string

	switch name.AttrKey {
	case attrkey.SpanErrorPct, attrkey.SpanErrorRate:
		unit = bununit.Utilization
	case attrkey.SpanDuration:
		unit = bununit.Nanoseconds
	}

	switch name.FuncName {
	case "",
		"sum", "avg", "min", "max",
		"any", "anyLast",
		"p50", "p75", "p90", "p95", "p99":
		return unit
	default:
		return ""
	}
}

func (f *SpanFilter) spanqlWhere(q *ch.SelectQuery) *ch.SelectQuery {
	for _, part := range f.parts {
		if part.Disabled || part.Error != "" {
			continue
		}

		switch ast := part.AST.(type) {
		case *tql.Where:
			where, _ := AppendWhereHaving(ast, f.TimeFilter.Duration())
			if len(where) > 0 {
				q = q.Where(string(where))
			}
		}
	}
	return q
}

//------------------------------------------------------------------------------

func NewSpanIndexQuery(app *bunapp.App) *ch.SelectQuery {
	return app.CH.NewSelect().
		TableExpr("? AS s", app.DistTable("spans_index_buffer"))
}

func buildSpanIndexQuery(
	app *bunapp.App, f *SpanFilter, dur time.Duration,
) (*ch.SelectQuery, *orderedmap.OrderedMap[string, *ColumnInfo]) {
	q := NewSpanIndexQuery(app).Apply(f.whereClause)
	return compileUQL(q, f.parts, dur)
}

func compileUQL(
	q *ch.SelectQuery, parts []*tql.QueryPart, dur time.Duration,
) (*ch.SelectQuery, *orderedmap.OrderedMap[string, *ColumnInfo]) {
	columnMap := orderedmap.New[string, *ColumnInfo]()
	groupingSet := make(map[string]bool)

	for _, part := range parts {
		if part.Disabled || part.Error != "" {
			continue
		}

		switch ast := part.AST.(type) {
		case *tql.Grouping:
			for _, name := range ast.Names {
				colName := name.String()

				q = tqlColumn(q, name, dur)
				columnMap.Set(colName, &ColumnInfo{
					Name:    colName,
					Unit:    unitFromName(name),
					IsGroup: true,
				})

				q = q.Group(colName)
				groupingSet[colName] = true
			}
		}
	}

	for _, part := range parts {
		if part.Disabled || part.Error != "" {
			continue
		}

		switch ast := part.AST.(type) {
		case *tql.Selector:
			for _, col := range ast.Columns {
				colName := col.Name.String()

				if !groupingSet[colName] && !IsAggColumn(col.Name) {
					part.SetError("must be an agg or a group-by")
					continue
				}

				if _, ok := columnMap.Get(colName); ok {
					continue
				}

				q = tqlColumn(q, col.Name, dur)
				columnMap.Set(colName, &ColumnInfo{
					Name:  colName,
					Unit:  unitFromName(col.Name),
					IsNum: col.Name.IsNum(),
				})
			}
		case *tql.Where:
			where, having := AppendWhereHaving(ast, dur)
			if len(where) > 0 {
				q = q.Where(string(where))
			}
			if len(having) > 0 {
				q = q.Having(string(having))
			}
		}
	}

	groupIDCol := tql.Name{AttrKey: attrkey.SpanGroupID}
	if _, ok := columnMap.Get(groupIDCol.String()); ok {
		for _, key := range []string{attrkey.SpanSystem, attrkey.DisplayName} {
			name := tql.Name{FuncName: "any", AttrKey: key}
			colName := name.String()

			if _, ok := columnMap.Get(colName); !ok {
				q = tqlColumn(q, name, dur)
			}
		}
	}

	return q, columnMap
}

func IsAggColumn(col tql.Name) bool {
	if col.FuncName != "" {
		return true
	}
	return isAggAttr(col.AttrKey)
}

func isAggAttr(attrKey string) bool {
	switch attrKey {
	case attrkey.SpanCount,
		attrkey.SpanCountPerMin,
		attrkey.SpanErrorCount,
		attrkey.SpanErrorPct,
		attrkey.SpanErrorRate:
		return true
	default:
		return false
	}
}

func tqlColumn(q *ch.SelectQuery, name tql.Name, dur time.Duration) *ch.SelectQuery {
	var b []byte
	b = AppendCHColumn(b, name, dur)
	b = append(b, " AS "...)
	b = append(b, '"')
	b = name.AppendString(b)
	b = append(b, '"')
	return q.ColumnExpr(string(b))
}

func AppendCHColumn(b []byte, name tql.Name, dur time.Duration) []byte {
	switch name.FuncName {
	case "p50", "p75", "p90", "p99":
		return chschema.AppendQuery(b, "quantileTDigest(?)(toFloat64OrDefault(?))",
			quantileLevel(name.FuncName), CHAttrExpr(name.AttrKey))
	case "top3":
		return chschema.AppendQuery(b, "topK(3)(?)", CHAttrExpr(name.AttrKey))
	case "top10":
		return chschema.AppendQuery(b, "topK(10)(?)", CHAttrExpr(name.AttrKey))
	}

	switch name.String() {
	case attrkey.SpanCount:
		return chschema.AppendQuery(b, "sum(s.count)")
	case attrkey.SpanCountPerMin:
		return chschema.AppendQuery(b, "sum(s.count) / ?", dur.Minutes())
	case attrkey.SpanErrorCount:
		return chschema.AppendQuery(b, "sumIf(s.count, s.status_code = 'error')", dur.Minutes())
	case attrkey.SpanErrorPct, attrkey.SpanErrorRate:
		return chschema.AppendQuery(
			b, "sumIf(s.count, s.status_code = 'error') / sum(s.count)", dur.Minutes())
	case attrkey.SpanIsEvent:
		return chschema.AppendQuery(
			b, "s.type IN ?", ch.In(EventTypes))
	default:
		if name.FuncName != "" {
			b = append(b, name.FuncName...)
			b = append(b, '(')
		}

		isNum := name.IsNum()
		if isNum {
			b = append(b, "toFloat64OrDefault("...)
		}

		b = AppendCHAttrExpr(b, name.AttrKey)

		if isNum {
			b = append(b, ')')
		}

		if name.FuncName != "" {
			b = append(b, ')')
		}

		return b
	}
}

func CHAttrExpr(key string) ch.Safe {
	return ch.Safe(AppendCHAttrExpr(nil, key))
}

func AppendCHAttrExpr(b []byte, key string) []byte {
	if strings.HasPrefix(key, ".") {
		key = strings.TrimPrefix(key, ".")
		b = append(b, "s."...)
		return chschema.AppendIdent(b, key)
	}

	if _, ok := indexedAttrSet[key]; ok {
		key = strings.ReplaceAll(key, ".", "_")
		b = append(b, "s."...)
		return chschema.AppendIdent(b, key)
	}

	return chschema.AppendQuery(b, "s.attr_values[indexOf(s.attr_keys, ?)]", key)
}

func AppendWhereHaving(ast *tql.Where, dur time.Duration) ([]byte, []byte) {
	var where []byte
	var having []byte

	for _, filter := range ast.Filters {
		bb := AppendFilter(filter, dur)
		if bb == nil {
			continue
		}

		if IsAggColumn(filter.LHS) {
			having = appendFilter(having, filter, bb)
		} else {
			where = appendFilter(where, filter, bb)
		}
	}

	return where, having
}

func AppendFilter(filter tql.Filter, dur time.Duration) []byte {
	var b []byte

	switch filter.Op {
	case tql.FilterExists, tql.FilterNotExists:
		if strings.HasPrefix(filter.LHS.AttrKey, ".") {
			if filter.Op == tql.FilterNotExists {
				b = append(b, '0')
			} else {
				b = append(b, '1')
			}
			return b
		}

		if filter.Op == tql.FilterNotExists {
			b = append(b, "NOT "...)
		}
		b = chschema.AppendQuery(b, "has(s.all_keys, ?)", filter.LHS.AttrKey)
		return b
	case tql.FilterIn, tql.FilterNotIn:
		if filter.Op == tql.FilterNotIn {
			b = append(b, "NOT "...)
		}

		var values []string
		switch rhs := filter.RHS.(type) {
		case tql.StringValues:
			values = rhs.Values
		case tql.StringValue:
			values = []string{rhs.Text}
		default:
			panic(fmt.Errorf("unsupported IN filter value type: %T", filter.RHS))
		}

		b = AppendCHColumn(b, filter.LHS, dur)
		b = append(b, " IN "...)
		b = chschema.AppendQuery(b, "?", ch.In(values))
		return b
	case tql.FilterContains, tql.FilterNotContains:
		if filter.Op == tql.FilterNotContains {
			b = append(b, "NOT "...)
		}

		values := strings.Split(filter.RHS.String(), "|")
		b = append(b, "multiSearchAnyCaseInsensitiveUTF8("...)
		b = AppendCHColumn(b, filter.LHS, dur)
		b = append(b, ", "...)
		b = chschema.AppendQuery(b, "?", ch.Array(values))
		b = append(b, ")"...)

		return b
	}

	b = AppendCHColumn(b, filter.LHS, dur)

	b = append(b, ' ')
	b = append(b, filter.Op...)
	b = append(b, ' ')

	b = chschema.AppendString(b, filter.RHS.String())

	return b
}

func appendFilter(b []byte, filter tql.Filter, bb []byte) []byte {
	if len(b) > 0 {
		b = append(b, filter.BoolOp...)
		b = append(b, ' ')
	}
	return append(b, bb...)
}

func disableColumnsAndGroups(parts []*tql.QueryPart) {
	for _, part := range parts {
		if part.Disabled || part.Error != "" {
			continue
		}

		switch ast := part.AST.(type) {
		case *tql.Selector:
			part.Disabled = true
		case *tql.Grouping:
			part.Disabled = true
		case *tql.Where:
			for _, filter := range ast.Filters {
				if IsAggColumn(filter.LHS) {
					part.Disabled = true
					break
				}
			}
		}
	}
}

//------------------------------------------------------------------------------

func spanSystemTableForWhere(app *bunapp.App, f *org.TimeFilter) ch.Ident {
	return spanSystemTable(app, org.TablePeriod(f))
}

func spanSystemTableForGroup(app *bunapp.App, f *org.TimeFilter) (ch.Ident, time.Duration) {
	tablePeriod, groupingPeriod := org.TableGroupingPeriod(f)
	return spanSystemTable(app, tablePeriod), groupingPeriod
}

func spanSystemTable(app *bunapp.App, period time.Duration) ch.Ident {
	switch period {
	case time.Minute:
		return app.DistTable("span_system_minutes")
	case time.Hour:
		return app.DistTable("span_system_hours")
	}
	panic("not reached")
}

//------------------------------------------------------------------------------

func spanServiceTableForGroup(app *bunapp.App, f *org.TimeFilter) (ch.Ident, time.Duration) {
	tablePeriod, groupingPeriod := org.TableGroupingPeriod(f)
	return spanServiceTable(app, tablePeriod), groupingPeriod
}

func spanServiceTable(app *bunapp.App, period time.Duration) ch.Ident {
	switch period {
	case time.Minute:
		return app.DistTable("span_service_minutes")
	case time.Hour:
		return app.DistTable("span_service_hours")
	}
	panic("not reached")
}

//------------------------------------------------------------------------------

func spanHostTableForWhere(app *bunapp.App, f *org.TimeFilter) ch.Ident {
	return spanHostTable(app, org.TablePeriod(f))
}

func spanHostTableForGroup(app *bunapp.App, f *org.TimeFilter) (ch.Ident, time.Duration) {
	tablePeriod, groupingPeriod := org.TableGroupingPeriod(f)
	return spanHostTable(app, tablePeriod), groupingPeriod
}

func spanHostTable(app *bunapp.App, period time.Duration) ch.Ident {
	switch period {
	case time.Minute:
		return app.DistTable("span_host_minutes")
	case time.Hour:
		return app.DistTable("span_host_hours")
	}
	panic("not reached")
}
