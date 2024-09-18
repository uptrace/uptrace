package tracing

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/chquery"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
	"github.com/uptrace/uptrace/pkg/urlstruct"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type LogFilter struct {
	org.OrderByMixin
	urlstruct.Pager
	TypeFilter

	Query string

	Search       string
	SearchTokens []chquery.Token `urlstruct:"-"`

	Column []string

	AttrKey     string
	SearchInput string

	QueryParts []*tql.QueryPart `urlstruct:"-"`
}

func DecodeLogFilter(req bunrouter.Request, f *LogFilter) error {
	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return err
	}

	if f.Search != "" {
		tokens, err := chquery.Parse(f.Search)
		if err != nil {
			return err
		}
		f.SearchTokens = tokens
	}

	project := org.ProjectFromContext(req.Context())
	f.ProjectID = project.ID
	f.QueryParts = tql.ParseQuery(f.Query)

	return nil
}

var _ urlstruct.ValuesUnmarshaler = (*LogFilter)(nil)

func (f *LogFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if err := f.TypeFilter.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	if err := f.Pager.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	if err := f.OrderByMixin.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	return nil
}

//------------------------------------------------------------------------------

func (f *LogFilter) whereClause(q *ch.SelectQuery) *ch.SelectQuery {
	for _, token := range f.SearchTokens {
		switch token.ID {
		case chquery.INCLUDE_TOKEN:
			q = q.Where("multiSearchAnyCaseInsensitiveUTF8(s.display_name, ?) > 0",
				ch.Array(token.Values))
		case chquery.EXCLUDE_TOKEN:
			q = q.Where("NOT multiSearchAnyCaseInsensitiveUTF8(s.display_name, ?) > 0",
				ch.Array(token.Values))
		case chquery.REGEXP_TOKEN:
			q = q.Where("match(s.display_name, ?)", token.Values[0])
		}
	}

	return f.TypeFilter.whereClause(q)
}

//------------------------------------------------------------------------------

func NewLogIndexQuery(db *ch.DB) *ch.SelectQuery {
	return db.NewSelect().Model((*LogIndex)(nil))
}

func BuildLogIndexQuery(
	db *ch.DB, f *LogFilter, dur time.Duration,
) (*ch.SelectQuery, *orderedmap.OrderedMap[string, *ColumnInfo]) {
	q := NewLogIndexQuery(db).Apply(f.whereClause)
	return compileUQLforLogs(q, f.QueryParts, dur)
}

func compileUQLforLogs(
	q *ch.SelectQuery, parts []*tql.QueryPart, dur time.Duration,
) (*ch.SelectQuery, *orderedmap.OrderedMap[string, *ColumnInfo]) {
	columnMap := orderedmap.New[string, *ColumnInfo]()
	groupingSet := make(map[string]bool)

	for _, part := range parts {
		if part.Disabled || part.Error.Wrapped != nil {
			continue
		}

		switch ast := part.AST.(type) {
		case *tql.Grouping:
			for i := range ast.Columns {
				col := &ast.Columns[i]
				colName := tql.String(col.Value)

				chExpr, err := appendCHColumn(nil, col, dur)
				if err != nil {
					part.Error.Wrapped = err
					continue
				}

				q = q.ColumnExpr(string(chExpr))
				columnMap.Set(colName, &ColumnInfo{
					Name:    colName,
					Unit:    unitForExpr(col.Value),
					IsGroup: true,
				})

				q = q.Group(colName)
				groupingSet[colName] = true
			}
		}
	}

	for _, part := range parts {
		if part.Disabled || part.Error.Wrapped != nil {
			continue
		}

		switch ast := part.AST.(type) {
		case *tql.Selector:
			for i := range ast.Columns {
				col := &ast.Columns[i]
				colName := tql.String(col.Value)

				if !groupingSet[colName] && !isAggExpr(col.Value) {
					part.Error.Wrapped = errors.New("must be an agg or a group-by")
					continue
				}

				if _, ok := columnMap.Get(colName); ok {
					continue
				}

				chExpr, err := appendCHColumn(nil, col, dur)
				if err != nil {
					part.Error.Wrapped = err
					continue
				}

				q = q.ColumnExpr(string(chExpr))
				columnMap.Set(colName, &ColumnInfo{
					Name:  colName,
					Unit:  unitForExpr(col.Value),
					IsNum: isNumExpr(col.Value),
				})
			}
		case *tql.Where:
			where, having, err := AppendWhereHaving(ast, dur)
			if err != nil {
				part.Error.Wrapped = err
			}
			if len(where) > 0 {
				q = q.Where(string(where))
			}
			if len(having) > 0 {
				q = q.Having(string(having))
			}
		}
	}

	if _, ok := columnMap.Get(attrkey.SpanGroupID); ok {
		for _, attrKey := range []string{attrkey.SpanSystem, attrkey.DisplayName} {
			col := &tql.Column{
				Value: &tql.FuncCall{
					Func: "any",
					Arg:  tql.Attr{Name: attrKey},
				},
				Alias: attrKey,
			}

			if _, ok := columnMap.Get(attrKey); ok {
				continue
			}

			chExpr, err := appendCHColumn(nil, col, dur)
			if err != nil {
				continue
			}
			q = q.ColumnExpr(string(chExpr))
		}
	}

	return q, columnMap
}

func disableColumnsAndGroupsforLogs(parts []*tql.QueryPart) {
	for _, part := range parts {
		if part.Disabled || part.Error.Wrapped != nil {
			continue
		}

		switch ast := part.AST.(type) {
		case *tql.Selector:
			part.Disabled = true
		case *tql.Grouping:
			part.Disabled = true
		case *tql.Where:
			for _, filter := range ast.Filters {
				if isAggExpr(filter.LHS) {
					part.Disabled = true
					break
				}
			}
		}
	}
}
