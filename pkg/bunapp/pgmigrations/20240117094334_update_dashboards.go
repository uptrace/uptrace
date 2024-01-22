package pgmigrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/org"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		var dashboards []*metrics.Dashboard

		if err := db.NewSelect().
			Model(&dashboards).
			Scan(ctx); err != nil {
			return err
		}

		for _, dash := range dashboards {
			dash.TableQuery = addDefaultAgg(ctx, dash.TableQuery)

			if _, err := db.NewUpdate().
				Model(dash).
				Column("table_query").
				Where("id = ?", dash.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		var charts []*metrics.ChartGridItem

		if err := db.NewSelect().
			Model(&charts).
			Where("type = ?", metrics.GridItemChart).
			Scan(ctx); err != nil {
			return err
		}

		for _, item := range charts {
			item.Params.Query = addDefaultAgg(ctx, item.Params.Query)

			if _, err := db.NewUpdate().
				Model(item).
				Column("params").
				Where("id = ?", item.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		var tables []*metrics.TableGridItem

		if err := db.NewSelect().
			Model(&charts).
			Where("type = ?", metrics.GridItemTable).
			Scan(ctx); err != nil {
			return err
		}

		for _, item := range tables {
			item.Params.Query = addDefaultAgg(ctx, item.Params.Query)

			if _, err := db.NewUpdate().
				Model(item).
				Column("params").
				Where("id = ?", item.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		var heatmaps []*metrics.HeatmapGridItem

		if err := db.NewSelect().
			Model(&charts).
			Where("type = ?", metrics.GridItemHeatmap).
			Scan(ctx); err != nil {
			return err
		}

		for _, item := range heatmaps {
			item.Params.Query = addDefaultAgg(ctx, item.Params.Query)

			if _, err := db.NewUpdate().
				Model(item).
				Column("params").
				Where("id = ?", item.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		var gauges []*metrics.GaugeGridItem

		if err := db.NewSelect().
			Model(&gauges).
			Where("type = ?", metrics.GridItemGauge).
			Scan(ctx); err != nil {
			return err
		}

		for _, item := range gauges {
			item.Params.Query = addDefaultAgg(ctx, item.Params.Query)

			if _, err := db.NewUpdate().
				Model(item).
				Column("params").
				Where("id = ?", item.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		var monitors []*org.MetricMonitor

		if err := db.NewSelect().
			Model(&monitors).
			Where("type = ?", org.MonitorMetric).
			Scan(ctx); err != nil {
			return err
		}

		for _, item := range monitors {
			item.Params.Query = addDefaultAgg(ctx, item.Params.Query)

			if _, err := db.NewUpdate().
				Model(item).
				Column("params").
				Where("id = ?", item.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}

func addDefaultAgg(ctx context.Context, queryStr string) string {
	query, err := mql.ParseQueryError(queryStr)
	if err != nil {
		return queryStr
	}

	for _, part := range query.Parts {
		switch expr := part.AST.(type) {
		case *ast.Selector:
			expr.Expr.Expr = addCHAggExpr(ctx, expr.Expr.Expr, nil)
			part.Query = expr.String()
		}
	}

	return query.String()
}

func addCHAggExpr(
	ctx context.Context, expr ast.Expr, parent ast.Expr,
) ast.Expr {
	switch expr := expr.(type) {
	case *ast.MetricExpr:
		fn, ok := parent.(*ast.FuncCall)
		if !ok {
			return wrapExpr(ctx, expr)
		}

		if !isCHFunc(fn.Func) {
			return wrapExpr(ctx, expr)
		}

		return expr
	case *ast.FuncCall:
		expr.Arg = addCHAggExpr(ctx, expr.Arg, expr)
		return expr
	case *ast.BinaryExpr:
		expr.LHS = addCHAggExpr(ctx, expr.LHS, expr)
		expr.RHS = addCHAggExpr(ctx, expr.RHS, expr)
		return expr
	case *ast.ParenExpr:
		expr.Expr = addCHAggExpr(ctx, expr.Expr, expr)
		return expr
	case ast.ParenExpr:
		expr.Expr = addCHAggExpr(ctx, expr.Expr, expr)
		return expr
	case *ast.UniqExpr:
		addCHAggExpr(ctx, expr.Name, expr)
		return expr
	case ast.Number:
		return expr
	default:
		return expr
	}
}

func wrapExpr(ctx context.Context, expr *ast.MetricExpr) ast.Expr {
	chAgg := defaultCHAgg(ctx, expr.Name)
	return &ast.FuncCall{
		Func: chAgg,
		Arg:  expr,
	}
}

func defaultCHAgg(ctx context.Context, metric string) string {
	app := bunapp.AppFromContext(ctx)

	var instrument metrics.Instrument
	if err := app.CH.NewSelect().
		TableExpr(metrics.TableDatapointHours).
		ColumnExpr("instrument").
		Where("metric = ?", metric).
		Limit(1).
		Scan(ctx, &instrument); err != nil {
		return "sum"
	}

	if instrument == metrics.InstrumentGauge {
		return "avg"
	}
	return "sum"
}

func isCHFunc(name string) bool {
	switch name {
	case mql.CHAggMin, mql.CHAggMax, mql.CHAggSum, mql.CHAggAvg, mql.CHAggMedian,
		mql.CHAggP50, mql.CHAggP75, mql.CHAggP90, mql.CHAggP95, mql.CHAggP99, mql.CHAggCount,
		mql.CHAggUniq:
		return true
	default:
		return false
	}
}
