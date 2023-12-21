package pgmigrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		var items []*metrics.GaugeGridItem

		if err := db.NewSelect().
			Model(&items).
			Where("type = ?", metrics.GridItemGauge).
			Scan(ctx); err != nil {
			return err
		}

		for _, item := range items {
			for _, col := range item.Params.ColumnMap {
				if col.AggFunc == "" {
					col.AggFunc = "last"
				}
			}

			if _, err := db.NewUpdate().
				Model(item).
				Column("params").
				Where("id = ?", item.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		var dashboards []*metrics.Dashboard

		if err := db.NewSelect().
			Model(&dashboards).
			Where("table_query != ''").
			Scan(ctx); err != nil {
			return err
		}

		for _, dash := range dashboards {
			if err := fixupDashboard(dash); err != nil {
				fmt.Println(err)
				continue
			}

			if _, err := db.NewUpdate().
				Model(dash).
				Column("table_query", "table_column_map").
				Where("id = ?", dash.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

func fixupDashboard(dash *metrics.Dashboard) error {
	query, err := mql.ParseQueryError(dash.TableQuery)
	if err != nil {
		return err
	}

	for _, part := range query.Parts {
		sel, ok := part.AST.(*ast.Selector)
		if !ok {
			continue
		}

		namedExpr := sel.Expr

		if col, ok := dash.TableColumnMap[namedExpr.Alias]; ok && col.AggFunc == "" {
			col.AggFunc = mql.TableFuncName(namedExpr.Expr)
		}

		fn, ok := namedExpr.Expr.(*ast.FuncCall)
		if !ok {
			continue
		}

		if fn.Func == "last" {
			part.Query = ast.String(fn.Arg)
		}
	}

	dash.TableQuery = query.String()
	for _, col := range dash.TableColumnMap {
		if col.AggFunc == "" {
			col.AggFunc = "last"
		}
	}

	return nil
}
