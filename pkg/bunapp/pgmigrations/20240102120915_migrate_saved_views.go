package pgmigrations

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/tracing"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
	"github.com/uptrace/uptrace/pkg/unsafeconv"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		var views []*tracing.SavedView
		if err := db.NewSelect().Model(&views).Scan(ctx); err != nil {
			return err
		}

		for _, view := range views {
			if query, _ := view.Query["query"].(string); query != "" {
				queryParts, err := tql.ParseQueryError(query)
				if err != nil {
					continue
				}

				for _, part := range queryParts {
					switch expr := part.AST.(type) {
					case *tql.Selector:
						for i := range expr.Columns {
							col := &expr.Columns[i]
							col.Value = renameAttrs(col.Value)
						}
						part.Query = expr.String()
					case *tql.Where:
						for i := range expr.Filters {
							f := &expr.Filters[i]
							f.LHS = renameAttrs(f.LHS)
						}
						part.Query = expr.String()
					case *tql.Grouping:
						for i := range expr.Columns {
							col := &expr.Columns[i]
							col.Value = renameAttrs(col.Value)
						}
						part.Query = expr.String()
					}
				}

				view.Query["query"] = joinQuery(queryParts)
			}
			if column, _ := view.Query["column"].(string); column != "" {
				view.Query["column"] = strings.ReplaceAll(column, ".", "_")
			}
			if sortBy, _ := view.Query["sort_by"].(string); sortBy != "" {
				view.Query["sort_by"] = strings.ReplaceAll(sortBy, ".", "_")
			}

			if _, err := db.NewUpdate().
				Model(view).
				Column("query").
				Where("id = ?", view.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

func renameAttrs(expr tql.Expr) tql.Expr {
	switch expr := expr.(type) {
	case tql.Attr:
		expr.Name = strings.ReplaceAll(expr.Name, ".", "_")
		return expr
	case *tql.FuncCall:
		expr.Arg = renameAttrs(expr.Arg)
		return expr
	case *tql.BinaryExpr:
		expr.LHS = renameAttrs(expr.LHS)
		expr.RHS = renameAttrs(expr.RHS)
		return expr
	case *tql.ParenExpr:
		expr.Expr = renameAttrs(expr.Expr)
		return expr
	case tql.ParenExpr:
		expr.Expr = renameAttrs(expr.Expr)
		return expr
	default:
		return expr
	}
}

func joinQuery(parts []*tql.QueryPart) string {
	b := make([]byte, 0, len(parts)*20)
	for i, part := range parts {
		if i > 0 {
			b = append(b, " | "...)
		}
		b = append(b, part.Query...)
	}
	return unsafeconv.String(b)
}
