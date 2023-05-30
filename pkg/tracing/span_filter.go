package tracing

import (
	"context"
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

func (f *SpanFilter) columns(items []map[string]any) []ColumnInfo {
	var item map[string]any
	if len(items) > 0 {
		item = items[0]
	}

	columns := make([]ColumnInfo, 0)

	for _, part := range f.parts {
		if part.Disabled || part.Error != "" {
			continue
		}

		switch ast := part.AST.(type) {
		case *tql.Columns:
			for _, name := range ast.Names {
				columns = append(columns, ColumnInfo{
					Name:  name.String(),
					IsNum: isNumColumn(item[name.String()]),
					Unit:  unitFromName(name),
				})
			}
		case *tql.Group:
			for _, name := range ast.Names {
				columns = append(columns, ColumnInfo{
					Name:    name.String(),
					IsGroup: true,
					Unit:    unitFromName(name),
				})
			}
		}
	}

	return columns
}

func isNumColumn(v any) bool {
	switch v.(type) {
	case int64, uint64, float32, float64:
		return true
	default:
		return false
	}
}

func unitFromName(name tql.Name) string {
	var unit string

	switch name.AttrKey {
	case attrkey.SpanErrorPct:
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

func buildSpanIndexQuery(app *bunapp.App, f *SpanFilter, dur time.Duration) *ch.SelectQuery {
	q := NewSpanIndexQuery(app).Apply(f.whereClause)
	q, f.columnMap = compileUQL(q, f.parts, dur)
	return q
}

func compileUQL(
	q *ch.SelectQuery, parts []*tql.QueryPart, dur time.Duration,
) (*ch.SelectQuery, map[string]bool) {
	groupSet := make(map[string]bool)
	columnSet := make(map[string]bool)

	for _, part := range parts {
		if part.Disabled || part.Error != "" {
			continue
		}

		switch ast := part.AST.(type) {
		case *tql.Group:
			for _, name := range ast.Names {
				q = tqlColumn(q, name, dur)
				columnSet[name.String()] = true

				q = q.Group(name.String())
				groupSet[name.String()] = true
			}
		}
	}

	for _, part := range parts {
		if part.Disabled || part.Error != "" {
			continue
		}

		switch ast := part.AST.(type) {
		case *tql.Columns:
			for _, name := range ast.Names {
				if !groupSet[name.String()] && !IsAggColumn(name) {
					part.SetError("must be an agg or a group-by")
					continue
				}

				if columnSet[name.String()] {
					continue
				}

				q = tqlColumn(q, name, dur)
				columnSet[name.String()] = true
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

	if columnSet[tql.Name{AttrKey: attrkey.SpanGroupID}.String()] {
		for _, key := range []string{attrkey.SpanSystem, attrkey.SpanName, attrkey.SpanEventName} {
			name := tql.Name{FuncName: "any", AttrKey: key}
			if !columnSet[name.String()] {
				q = tqlColumn(q, name, dur)
				columnSet[name.String()] = true
			}
		}
	}

	return q, columnSet
}

func IsAggColumn(col tql.Name) bool {
	if col.FuncName != "" {
		return true
	}
	return isAggAttr(col.AttrKey)
}

func isAggAttr(attrKey string) bool {
	switch attrKey {
	case attrkey.SpanCount, attrkey.SpanCountPerMin, attrkey.SpanErrorCount, attrkey.SpanErrorPct:
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
	b = name.Append(b)
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
	case attrkey.SpanErrorPct:
		return chschema.AppendQuery(
			b, "sumIf(s.count, s.status_code = 'error') / sum(s.count)", dur.Minutes())
	case attrkey.SpanIsEvent:
		return chschema.AppendQuery(
			b, "s.type IN (?)", ch.In(EventTypes))
	default:
		if name.FuncName != "" {
			b = append(b, name.FuncName...)
			b = append(b, '(')
		}

		b = AppendCHAttrExpr(b, name.AttrKey)

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
	if strings.HasPrefix(key, "span.") {
		key = strings.TrimPrefix(key, "span.")
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

	for _, cond := range ast.Conds {
		bb := AppendCond(cond, dur)
		if bb == nil {
			continue
		}

		if IsAggColumn(cond.Left) {
			having = appendCond(having, cond, bb)
		} else {
			where = appendCond(where, cond, bb)
		}
	}

	return where, having
}

func AppendCond(cond tql.Cond, dur time.Duration) []byte {
	var b []byte

	switch cond.Op {
	case tql.ExistsOp, tql.DoesNotExistOp:
		if strings.HasPrefix(cond.Left.AttrKey, "span.") {
			if cond.Op == tql.DoesNotExistOp {
				b = append(b, '0')
			} else {
				b = append(b, '1')
			}
			return b
		}

		if cond.Op == tql.DoesNotExistOp {
			b = append(b, "NOT "...)
		}
		b = chschema.AppendQuery(b, "has(s.all_keys, ?)", cond.Left.AttrKey)
		return b
	case tql.ContainsOp, tql.DoesNotContainOp:
		if cond.Op == tql.DoesNotContainOp {
			b = append(b, "NOT "...)
		}

		values := strings.Split(cond.Right.Text, "|")
		b = append(b, "multiSearchAnyCaseInsensitiveUTF8("...)
		b = AppendCHColumn(b, cond.Left, dur)
		b = append(b, ", "...)
		b = chschema.AppendQuery(b, "[?]", ch.In(values))
		b = append(b, ")"...)

		return b
	}

	b = AppendCHColumn(b, cond.Left, dur)

	b = append(b, ' ')
	b = append(b, cond.Op...)
	b = append(b, ' ')

	if cond.Right.IsNum() {
		b = append(b, "toString("...)
	}
	b = cond.Right.Append(b)
	if cond.Right.IsNum() {
		b = append(b, ')')
	}

	return b
}

func appendCond(b []byte, cond tql.Cond, bb []byte) []byte {
	if len(b) > 0 {
		b = append(b, cond.Sep.Op...)
		b = append(b, ' ')
	}
	if cond.Sep.Negate {
		b = append(b, "NOT "...)
	}
	return append(b, bb...)
}

func disableColumnsAndGroups(parts []*tql.QueryPart) {
	for _, part := range parts {
		if part.Disabled || part.Error != "" {
			continue
		}

		switch ast := part.AST.(type) {
		case *tql.Columns:
			part.Disabled = true
		case *tql.Group:
			part.Disabled = true
		case *tql.Where:
			for _, cond := range ast.Conds {
				if IsAggColumn(cond.Left) {
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
