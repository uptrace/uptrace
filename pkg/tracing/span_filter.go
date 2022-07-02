package tracing

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/tracing/xattr"
	"github.com/uptrace/uptrace/pkg/uql"
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

type SpanFilter struct {
	*bunapp.App `urlstruct:"-"`

	OrderByMixin
	urlstruct.Pager
	TimeFilter

	ProjectID uint32
	System    string
	GroupID   uint64

	Query  string
	Column string

	parts     []*uql.Part
	columnMap map[string]bool
}

func DecodeSpanFilter(app *bunapp.App, req bunrouter.Request) (*SpanFilter, error) {
	f := &SpanFilter{App: app}
	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return nil, err
	}
	return f, nil
}

var _ urlstruct.ValuesUnmarshaler = (*SpanFilter)(nil)

func (f *SpanFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if err := f.TimeFilter.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	if err := f.Pager.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	if f.System == "" {
		return errors.New("'system' query param is required")
	}

	f.parts = uql.Parse(f.Query)

	return nil
}

func (f *SpanFilter) whereClause(q *ch.SelectQuery) *ch.SelectQuery {
	q = q.Where("project_id = ?", f.ProjectID).
		Where("`span.time` >= ?", f.TimeGTE).
		Where("`span.time` < ?", f.TimeLT)

	switch {
	case f.System == allSpanType:
		q = q.Where("`span.system` != ?", internalSpanType)
	case strings.HasSuffix(f.System, ":all"):
		system := strings.TrimSuffix(f.System, ":all")
		q = q.Where("startsWith(`span.system`, ?)", system)
	default:
		q = q.Where("`span.system` = ?", f.System)
	}

	if f.GroupID != 0 {
		q = q.Where("`span.group_id` = ?", f.GroupID)
	}

	return q
}

//------------------------------------------------------------------------------

type ColumnInfo struct {
	Name    string `json:"name"`
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
		case *uql.Columns:
			for _, name := range ast.Names {
				columns = append(columns, ColumnInfo{
					Name:  name.String(),
					IsNum: isNumColumn(item[name.String()]),
				})
			}
		case *uql.Group:
			for _, name := range ast.Names {
				columns = append(columns, ColumnInfo{
					Name:    name.String(),
					IsGroup: true,
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

//------------------------------------------------------------------------------

func buildSpanIndexQuery(app *bunapp.App, f *SpanFilter, minutes float64) *ch.SelectQuery {
	q := f.CH().NewSelect().
		TableExpr("? AS s", app.DistTable("spans_index")).
		WithQuery(f.whereClause)
	q, f.columnMap = compileUQL(q, f.parts, minutes)
	return q
}

func compileUQL(
	q *ch.SelectQuery, parts []*uql.Part, minutes float64,
) (*ch.SelectQuery, map[string]bool) {
	groupSet := make(map[string]bool)
	columnSet := make(map[string]bool)

	for _, part := range parts {
		if part.Disabled || part.Error != "" {
			continue
		}

		switch ast := part.AST.(type) {
		case *uql.Group:
			for _, name := range ast.Names {
				q = uqlColumn(q, name, minutes)
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
		case *uql.Columns:
			for _, name := range ast.Names {
				if !groupSet[name.String()] && !isAggColumn(name) {
					part.SetError("must be an agg or a group-by")
					continue
				}

				if columnSet[name.String()] {
					continue
				}

				q = uqlColumn(q, name, minutes)
				columnSet[name.String()] = true
			}
		case *uql.Where:
			q = uqlWhere(q, ast, minutes)
		}
	}

	if columnSet[uql.Name{AttrKey: xattr.SpanGroupID}.String()] {
		for _, key := range []string{xattr.SpanSystem, xattr.SpanName, xattr.SpanEventName} {
			name := uql.Name{FuncName: "any", AttrKey: key}
			if !columnSet[name.String()] {
				q = uqlColumn(q, name, minutes)
				columnSet[name.String()] = true
			}
		}
	}

	return q, columnSet
}

func isAggColumn(col uql.Name) bool {
	if col.FuncName != "" {
		return true
	}
	switch col.AttrKey {
	case xattr.SpanCount, xattr.SpanCountPerMin, xattr.SpanErrorCount, xattr.SpanErrorPct:
		return true
	default:
		return false
	}
}

func uqlColumn(q *ch.SelectQuery, name uql.Name, minutes float64) *ch.SelectQuery {
	var b []byte
	b = appendUQLColumn(b, name, minutes)
	b = append(b, " AS "...)
	b = append(b, '"')
	b = name.Append(b)
	b = append(b, '"')
	return q.ColumnExpr(string(b))
}

func appendUQLColumn(b []byte, name uql.Name, minutes float64) []byte {
	switch name.FuncName {
	case "p50", "p75", "p90", "p99":
		return chschema.AppendQuery(b, "quantileTDigest(?)(toFloat64OrDefault(?))",
			quantileLevel(name.FuncName), chColumn(name.AttrKey))
	case "top3":
		return chschema.AppendQuery(b, "topK(3)(?)", chColumn(name.AttrKey))
	case "top10":
		return chschema.AppendQuery(b, "topK(10)(?)", chColumn(name.AttrKey))
	}

	switch name.String() {
	case xattr.SpanCount:
		return chschema.AppendQuery(b, "sum(`span.count`)")
	case xattr.SpanCountPerMin:
		return chschema.AppendQuery(b, "sum(`span.count`) / ?", minutes)
	case xattr.SpanErrorCount:
		return chschema.AppendQuery(b, "sumIf(`span.count`, `span.status_code` = 'error')", minutes)
	case xattr.SpanErrorPct:
		return chschema.AppendQuery(
			b, "sumIf(`span.count`, `span.status_code` = 'error') / sum(`span.count`)", minutes)
	default:
		if name.FuncName != "" {
			b = append(b, name.FuncName...)
			b = append(b, '(')
		}

		b = appendCHColumn(b, name.AttrKey)

		if name.FuncName != "" {
			b = append(b, ')')
		}

		return b
	}
}

func chColumn(key string) ch.Safe {
	return ch.Safe(appendCHColumn(nil, key))
}

func appendCHColumn(b []byte, key string) []byte {
	if strings.HasPrefix(key, "span.") {
		return chschema.AppendIdent(b, key)
	}

	if _, ok := indexedAttrSet[key]; ok {
		return chschema.AppendIdent(b, key)
	}
	return chschema.AppendQuery(b, "attr_values[indexOf(attr_keys, ?)]", key)
}

func uqlWhere(q *ch.SelectQuery, ast *uql.Where, minutes float64) *ch.SelectQuery {
	var where []byte
	var having []byte

	for _, cond := range ast.Conds {
		bb, isAgg := uqlWhereCond(cond, minutes)
		if bb == nil {
			continue
		}

		if isAgg {
			having = appendCond(having, cond, bb)
		} else {
			where = appendCond(where, cond, bb)
		}
	}

	if len(where) > 0 {
		q = q.Where(string(where))
	}
	if len(having) > 0 {
		q = q.Having(string(having))
	}

	return q
}

func uqlWhereCond(cond uql.Cond, minutes float64) (b []byte, isAgg bool) {
	isAgg = isAggColumn(cond.Left)

	switch cond.Op {
	case uql.ExistsOp, uql.DoesNotExistOp:
		if isAgg {
			return nil, false
		}

		if strings.HasPrefix(cond.Left.AttrKey, "span.") {
			b = append(b, '1')
			return b, false
		}
		b = chschema.AppendQuery(b, "has(all_keys, ?)", cond.Left.AttrKey)
		return b, false
	case uql.ContainsOp, uql.DoesNotContainOp:
		if cond.Op == uql.DoesNotContainOp {
			b = append(b, "NOT "...)
		}

		values := strings.Split(cond.Right.Text, "|")
		b = append(b, "multiSearchAnyCaseInsensitiveUTF8("...)
		b = appendUQLColumn(b, cond.Left, minutes)
		b = append(b, ", "...)
		b = chschema.AppendQuery(b, "[?]", ch.In(values))
		b = append(b, ")"...)

		return b, isAgg
	}

	if cond.Right.Kind == uql.NumberValue {
		b = append(b, "toFloat64OrDefault("...)
	}
	b = appendUQLColumn(b, cond.Left, minutes)
	if cond.Right.Kind == uql.NumberValue {
		b = append(b, ")"...)
	}

	b = append(b, ' ')
	b = append(b, cond.Op...)
	b = append(b, ' ')

	b = cond.Right.Append(b)

	return b, isAgg
}

func appendCond(b []byte, cond uql.Cond, bb []byte) []byte {
	if len(b) > 0 {
		b = append(b, cond.Sep.Op...)
		b = append(b, ' ')
	}
	if cond.Sep.Negate {
		b = append(b, "NOT "...)
	}
	return append(b, bb...)
}

func disableColumnsAndGroups(parts []*uql.Part) {
	for _, part := range parts {
		if part.Disabled || part.Error != "" {
			continue
		}

		switch part.AST.(type) {
		case *uql.Columns:
			part.Disabled = true
		case *uql.Group:
			part.Disabled = true
		}
	}
}

//------------------------------------------------------------------------------

func spanSystemTableForWhere(app *bunapp.App, f *TimeFilter) ch.Ident {
	return spanSystemTable(app, tablePeriod(f))
}

func spanSystemTableForGroup(app *bunapp.App, f *TimeFilter) (ch.Ident, time.Duration) {
	tablePeriod, groupPeriod := tableGroupPeriod(f)
	return spanSystemTable(app, tablePeriod), groupPeriod
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

func spanServiceTableForGroup(app *bunapp.App, f *TimeFilter) (ch.Ident, time.Duration) {
	tablePeriod, groupPeriod := tableGroupPeriod(f)
	return spanServiceTable(app, tablePeriod), groupPeriod
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

func spanHostTableForWhere(app *bunapp.App, f *TimeFilter) ch.Ident {
	return spanHostTable(app, tablePeriod(f))
}

func spanHostTableForGroup(app *bunapp.App, f *TimeFilter) (ch.Ident, time.Duration) {
	tablePeriod, groupPeriod := tableGroupPeriod(f)
	return spanHostTable(app, tablePeriod), groupPeriod
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

//------------------------------------------------------------------------------

func tablePeriod(f *TimeFilter) time.Duration {
	var period time.Duration

	if d := f.TimeLT.Sub(f.TimeGTE); d >= 6*time.Hour {
		period = time.Hour
	} else {
		period = time.Minute
	}

	return period
}

func tableGroupPeriod(f *TimeFilter) (tablePeriod, groupPeriod time.Duration) {
	groupPeriod = calcGroupPeriod(f, 200)
	if groupPeriod >= time.Hour {
		tablePeriod = time.Hour
	} else {
		tablePeriod = time.Minute
	}
	return tablePeriod, groupPeriod
}

func calcGroupPeriod(f *TimeFilter, n int) time.Duration {
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
