package tracing

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

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

var _ urlstruct.ValuesUnmarshaler = (*OrderByMixin)(nil)

func (f *OrderByMixin) UnmarshalValues(ctx context.Context, values url.Values) error {
	if f.SortDir == "" {
		f.SortDir = "desc"
	}
	return nil
}

func (f *OrderByMixin) CHOrder(q *ch.SelectQuery) *ch.SelectQuery {
	if f.SortBy == "" {
		return q
	}
	return q.Order(f.SortBy + " " + f.SortDir)
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

	System  string
	GroupID uint64

	Query  string
	Column string

	parts     []*uql.Part
	columnMap map[uql.Name]bool
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
	q = q.Where("s.`span.time` >= ?", f.TimeGTE).
		Where("s.`span.time` < ?", f.TimeLT)

	switch {
	case f.System == allSpanType:
		q = q.Where("s.`span.system` != ?", internalSpanType)
	case strings.HasSuffix(f.System, ":all"):
		system := strings.TrimSuffix(f.System, ":all")
		q = q.Where("startsWith(s.`span.system`, ?)", system)
	default:
		q = q.Where("s.`span.system` = ?", f.System)
	}

	if f.GroupID != 0 {
		q = q.Where("s.`span.group_id` = ?", f.GroupID)
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

	var columns []ColumnInfo
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

func buildSpanIndexQuery(f *SpanFilter, minutes float64) *ch.SelectQuery {
	return buildSpanIndexQuerySlow(f, minutes)
}

func buildSpanIndexQuerySlow(f *SpanFilter, minutes float64) *ch.SelectQuery {
	q := f.CH().NewSelect().Model((*SpanIndex)(nil))
	q = f.whereClause(q)
	q, f.columnMap = compileUQLSlow(q, f.parts, minutes)
	return q
}

func compileUQLSlow(
	q *ch.SelectQuery, parts []*uql.Part, minutes float64,
) (*ch.SelectQuery, map[uql.Name]bool) {
	columnMap := make(map[uql.Name]bool)
	var groups []uql.Name
	var where []*uql.Part

	for _, part := range parts {
		if part.Disabled || part.Error != "" {
			continue
		}

		switch ast := part.AST.(type) {
		case *uql.Columns:
			for _, name := range ast.Names {
				q = uqlColumnSlow(q, name, minutes)
				columnMap[name] = true
			}
		case *uql.Group:
			groups = append(groups, ast.Names...)
		case *uql.Where:
			where = append(where, part)
		}
	}

	for _, name := range groups {
		if !columnMap[name] {
			q = uqlColumnSlow(q, name, minutes)
			columnMap[name] = true
		}
		q = q.Group(name.String())
	}

	if columnMap[uql.Name{AttrKey: xattr.SpanGroupID}] {
		for _, key := range []string{xattr.SpanSystem, xattr.SpanName} {
			name := uql.Name{FuncName: "any", AttrKey: key}
			if !columnMap[name] {
				q = uqlColumnSlow(q, name, minutes)
				columnMap[name] = true
			}
		}
	}

	for _, part := range where {
		ast := part.AST.(*uql.Where)
		q = uqlWhere(q, ast)
	}

	return q, columnMap
}

func uqlColumnSlow(q *ch.SelectQuery, name uql.Name, minutes float64) *ch.SelectQuery {
	switch name.FuncName {
	case "p50":
		return q.ColumnExpr("quantileTDigest(0.5)(toFloat64OrDefault(?)) AS ?",
			chColumn(name.AttrKey), ch.Ident(name.String()))
	case "p75":
		return q.ColumnExpr("quantileTDigest(0.75)(toFloat64OrDefault(?)) AS ?",
			chColumn(name.AttrKey), ch.Ident(name.String()))
	case "p90":
		return q.ColumnExpr("quantileTDigest(0.9)(toFloat64OrDefault(?)) AS ?",
			chColumn(name.AttrKey), ch.Ident(name.String()))
	case "p99":
		return q.ColumnExpr("quantileTDigest(0.99)(toFloat64OrDefault(?)) AS ?",
			chColumn(name.AttrKey), ch.Ident(name.String()))
	case "top3":
		return q.ColumnExpr("topK(3)(?) AS ?", chColumn(name.AttrKey), ch.Ident(name.String()))
	case "top10":
		return q.ColumnExpr("topK(10)(?) AS ?", chColumn(name.AttrKey), ch.Ident(name.String()))
	}

	switch name.String() {
	case xattr.SpanCount:
		return q.ColumnExpr("count() AS ?", ch.Ident(xattr.SpanCount))
	case xattr.SpanCountPerMin:
		return q.ColumnExpr("count() / ? AS ?", minutes, ch.Ident(xattr.SpanCountPerMin))
	case xattr.SpanErrorCount:
		return q.ColumnExpr("countIf(`span.status_code` = 'error') AS ?",
			ch.Ident(xattr.SpanErrorCount))
	case xattr.SpanErrorPct:
		return q.ColumnExpr("countIf(`span.status_code` = 'error') / count() AS ?",
			ch.Ident(xattr.SpanErrorPct))
	default:
		var b []byte

		if name.FuncName != "" {
			b = append(b, name.FuncName...)
			b = append(b, '(')
		}

		b = appendCHColumn(b, name.AttrKey)

		if name.FuncName != "" {
			b = append(b, ')')
		}

		b = append(b, " AS "...)
		b = append(b, '"')
		b = name.Append(b)
		b = append(b, '"')

		return q.ColumnExpr(string(b))
	}
}

func chColumn(key string) ch.Safe {
	return ch.Safe(appendCHColumn(nil, key))
}

func appendCHColumn(b []byte, key string) []byte {
	switch key {
	case xattr.SpanSystem, xattr.SpanGroupID, xattr.SpanTraceID,
		xattr.SpanName, xattr.SpanKind, xattr.SpanDuration,
		xattr.SpanStatusCode, xattr.SpanStatusMessage,
		xattr.ServiceName, xattr.HostName:
		b = append(b, "s."...)
		b = chschema.AppendIdent(b, key)
		return b
	default:
		return chschema.AppendQuery(b, "attr_values[indexOf(attr_keys, ?)]", key)
	}
}

func uqlWhere(q *ch.SelectQuery, ast *uql.Where) *ch.SelectQuery {
	var b []byte

	for i, cond := range ast.Conds {
		if cond.Sep.Negate {
			b = append(b, "NOT "...)
		}
		if i > 0 {
			b = append(b, cond.Sep.Op...)
			b = append(b, ' ')
		}
		b = uqlWhereCond(b, cond)
	}

	return q.Where(string(b))
}

func uqlWhereCond(b []byte, cond uql.Cond) []byte {
	switch cond.Op {
	case uql.ExistsOp, uql.DoesNotExistOp:
		if strings.HasPrefix(cond.Left.AttrKey, "span.") {
			return append(b, '1')
		}
		return chschema.AppendQuery(b, "has(attr_keys, ?)", cond.Left.AttrKey)
	case uql.ContainsOp, uql.DoesNotContainOp:
		if cond.Op == uql.DoesNotContainOp {
			b = append(b, "NOT "...)
		}

		values := strings.Split(cond.Right.Text, "|")
		b = append(b, "multiSearchAnyCaseInsensitiveUTF8("...)
		b = appendCHColumn(b, cond.Left.AttrKey)
		b = append(b, ", "...)
		b = chschema.AppendQuery(b, "[?]", ch.In(values))
		b = append(b, ")"...)

		return b
	}

	if cond.Right.Kind == uql.NumberValue {
		b = append(b, "toFloat64OrDefault("...)
	}
	b = appendCHColumn(b, cond.Left.AttrKey)
	if cond.Right.Kind == uql.NumberValue {
		b = append(b, ")"...)
	}

	b = append(b, ' ')
	b = append(b, cond.Op...)
	b = append(b, ' ')

	b = cond.Right.Append(b)

	return b
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

func spanSystemTableForWhere(f *TimeFilter) string {
	return spanSystemTable(tablePeriod(f))
}

func spanSystemTableForGroup(f *TimeFilter) (string, time.Duration) {
	tablePeriod, groupPeriod := tableGroupPeriod(f)
	return spanSystemTable(tablePeriod), groupPeriod
}

func spanSystemTable(period time.Duration) string {
	switch period {
	case time.Minute:
		return "span_system_minutes AS s"
	case time.Hour:
		return "span_system_hours AS s"
	}
	panic("not reached")
}

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
