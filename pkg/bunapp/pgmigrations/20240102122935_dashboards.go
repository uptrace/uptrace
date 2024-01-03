package pgmigrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
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
			for i := range dash.TableMetrics {
				metric := &dash.TableMetrics[i]
				metric.Name = updateMetricName(metric.Name)
			}
			dash.TableQuery = updateMetricQuery(dash.TableQuery)

			for i, grouping := range dash.TableGrouping {
				dash.TableGrouping[i] = strings.ReplaceAll(grouping, ".", "_")
			}

			if _, err := db.NewUpdate().
				Model(dash).
				Column("table_metrics", "table_query", "table_grouping").
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
			for i := range item.Params.Metrics {
				metric := &item.Params.Metrics[i]
				metric.Name = updateMetricName(metric.Name)
			}
			item.Params.Query = updateMetricQuery(item.Params.Query)

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
			for i := range item.Params.Metrics {
				metric := &item.Params.Metrics[i]
				metric.Name = updateMetricName(metric.Name)
			}
			item.Params.Query = updateMetricQuery(item.Params.Query)

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
			item.Params.Metric = updateMetricName(item.Params.Metric)
			item.Params.Query = updateMetricQuery(item.Params.Query)

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
			for i := range item.Params.Metrics {
				metric := &item.Params.Metrics[i]
				metric.Name = updateMetricName(metric.Name)
			}
			item.Params.Query = updateMetricQuery(item.Params.Query)

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

func updateMetricName(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}

func updateMetricQuery(queryStr string) string {
	query, err := mql.ParseQueryError(queryStr)
	if err != nil {
		return queryStr
	}

	for _, part := range query.Parts {
		switch expr := part.AST.(type) {
		case *ast.Selector:
			expr.Expr.Expr = renameMetricAttrs(expr.Expr.Expr)
			part.Query = expr.String()
		case *ast.Where:
			for i := range expr.Filters {
				f := &expr.Filters[i]
				f.LHS = strings.ReplaceAll(f.LHS, ".", "_")
			}
			part.Query = expr.String()
		case *ast.Grouping:
			for i := range expr.Elems {
				elem := &expr.Elems[i]
				elem.Name = strings.ReplaceAll(elem.Name, ".", "_")
			}
			part.Query = expr.String()
		default:
			panic(fmt.Errorf("unsupported: %T", expr))
		}
	}

	return query.String()
}

func renameMetricAttrs(expr ast.Expr) ast.Expr {
	switch expr := expr.(type) {
	case *ast.MetricExpr:
		expr.Name = strings.ReplaceAll(expr.Name, ".", "_")
		for i := range expr.Filters {
			f := &expr.Filters[i]
			f.LHS = strings.ReplaceAll(f.LHS, ".", "_")
		}
		for i := range expr.Grouping {
			elem := &expr.Grouping[i]
			elem.Name = strings.ReplaceAll(elem.Name, ".", "_")
		}
		return expr
	case *ast.FuncCall:
		expr.Arg = renameMetricAttrs(expr.Arg)
		return expr
	case *ast.BinaryExpr:
		expr.LHS = renameMetricAttrs(expr.LHS)
		expr.RHS = renameMetricAttrs(expr.RHS)
		return expr
	case *ast.ParenExpr:
		expr.Expr = renameMetricAttrs(expr.Expr)
		return expr
	case ast.ParenExpr:
		expr.Expr = renameMetricAttrs(expr.Expr)
		return expr
	case *ast.UniqExpr:
		renameMetricAttrs(&expr.Name)
		for i, attr := range expr.Attrs {
			expr.Attrs[i] = strings.ReplaceAll(attr, ".", "_")
		}
		return expr
	case ast.Number:
		return expr
	default:
		return expr
	}
}
