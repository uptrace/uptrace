package metrics

import (
	"context"
	"fmt"
	"strings"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/bununit"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/tracing"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
)

const spanMetricDur = 1

func initSpanMetrics(ctx context.Context, app *bunapp.App) error {
	conf := app.Config()
	for i := range conf.MetricsFromSpans {
		metric := &conf.MetricsFromSpans[i]
		if metric.Name == "" {
			return fmt.Errorf("metric name can't be empty")
		}
		if err := createSpanMetric(ctx, app, metric); err != nil {
			return fmt.Errorf("createSpanMetric %q failed: %w", metric.Name, err)
		}
	}
	return nil
}

func createSpanMetric(ctx context.Context, app *bunapp.App, metric *bunconf.SpanMetric) error {
	if metric.Instrument == "" {
		return fmt.Errorf("metric instrument can't be empty")
	}

	if err := createSpanMetricMeta(ctx, app, metric); err != nil {
		return fmt.Errorf("createSpanMetricMeta failed: %w", err)
	}
	if err := createMatView(ctx, app, metric); err != nil {
		return fmt.Errorf("createMatView failed: %w", err)
	}
	return nil
}

func createSpanMetricMeta(ctx context.Context, app *bunapp.App, metric *bunconf.SpanMetric) error {
	projects := app.Config().Projects
	for i := range projects {
		project := &projects[i]

		attrKeys := make([]string, len(metric.Attrs))
		for i, attr := range metric.Attrs {
			attrKeys[i], _ = splitNameAlias(attr)
		}

<<<<<<< HEAD
		if err := UpsertMetric(ctx, app, &Metric{
=======
		if _, err := UpsertMetric(ctx, app, &Metric{
>>>>>>> sync-leonyu879
			ProjectID:   project.ID,
			Name:        metric.Name,
			Description: metric.Description,
			Unit:        bununit.FromString(metric.Unit),
			Instrument:  Instrument(metric.Instrument),
			AttrKeys:    attrKeys,
		}); err != nil {
			return err
		}
	}
	return nil
}

func createMatView(ctx context.Context, app *bunapp.App, metric *bunconf.SpanMetric) error {
	conf := app.Config()
	viewName := "metrics_" + strings.ReplaceAll(metric.Name, ".", "_") + "_mv"

	if _, err := app.CH.NewDropView().
		IfExists().
		View(viewName).
		OnCluster(conf.CHSchema.Cluster).
		Exec(ctx); err != nil {
		return err
	}

	valueExpr, err := compileSpanMetricValue(metric.Value)
	if err != nil {
		return err
	}

	q := app.CH.NewCreateView().
		Materialized().
		View(viewName).
		OnCluster(conf.CHSchema.Cluster).
		ToExpr("?DB.measure_minutes").
		ColumnExpr("s.project_id").
		ColumnExpr("? AS metric", metric.Name).
		ColumnExpr("toStartOfMinute(s.time) AS time").
		ColumnExpr("? AS instrument", metric.Instrument).
		TableExpr("?DB.spans_index AS s").
		GroupExpr("s.project_id, toStartOfMinute(s.time)")

	if len(metric.Attrs) > 0 {
		attrsExpr, aliases := compileSpanMetricAttrs(metric.Attrs)
		q = q.
			ColumnExpr("xxHash64(arrayStringConcat([?], '-')) AS attrs_hash", attrsExpr).
			ColumnExpr("? AS string_keys", ch.Array(aliases)).
			ColumnExpr("[?] AS string_values", attrsExpr).
			GroupExpr(string(attrsExpr))
	}

	if len(metric.Annotations) > 0 {
		expr := compileSpanMetricAnnotations(metric.Annotations)
		q = q.ColumnExpr("toJSONString(map(?)) AS annotations", expr)
	}

	if metric.Where != "" {
		whereExpr, err := compileSpanMetricWhere(metric.Where)
		if err != nil {
			return err
		}
		if whereExpr != "" {
			q = q.Where(string(whereExpr))
		}
	}

	switch Instrument(metric.Instrument) {
	case InstrumentGauge:
		q = q.ColumnExpr("? AS value", valueExpr)
	case InstrumentAdditive:
		q = q.ColumnExpr("? AS value", valueExpr)
	case InstrumentCounter:
		q = q.ColumnExpr("? AS sum", valueExpr)
	case InstrumentHistogram:
		q = q.ColumnExpr("count() AS count").
			ColumnExpr("sum(?) AS sum", valueExpr).
			ColumnExpr("quantilesBFloat16State(0.5)(toFloat32(?)) AS histogram", valueExpr)
	default:
		return fmt.Errorf("unsupported instrument: %q", metric.Instrument)
	}

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func compileSpanMetricValue(value string) (ch.Safe, error) {
	query := mql.Parse(value)
	if len(query.Parts) != 1 {
		return "", fmt.Errorf("can't parse metric value: %q", value)
	}

	part := query.Parts[0]
	sel, ok := part.AST.(*ast.Selector)
	if !ok {
		return "", fmt.Errorf("unsupported metric value AST: %T", part.AST)
	}

	var b []byte
	b, err := appendSpanMetricExpr(b, sel.Expr.Expr)
	if err != nil {
		return "", err
	}

	return ch.Safe(b), nil
}

func appendSpanMetricExpr(b []byte, expr ast.Expr) (_ []byte, err error) {
	switch expr := expr.(type) {
	case *ast.Name:
		b = tracing.AppendCHColumn(b, tql.Name{
			FuncName: expr.Func,
			AttrKey:  expr.Name,
		}, spanMetricDur)
		return b, nil
	case *ast.Number:
		b = append(b, expr.Text...)
		return b, nil
	case *ast.ParenExpr:
		b = append(b, '(')
		b, err = appendSpanMetricExpr(b, expr.Expr)
		if err != nil {
			return nil, err
		}
		b = append(b, ')')
		return b, nil
	case *ast.BinaryExpr:
		b, err = appendSpanMetricExpr(b, expr.LHS)
		if err != nil {
			return nil, err
		}

		b = append(b, ' ')
		b = append(b, expr.Op...)
		b = append(b, ' ')

		b, err = appendSpanMetricExpr(b, expr.RHS)
		if err != nil {
			return nil, err
		}

		return b, nil
	default:
		return nil, fmt.Errorf("unsupported span metric expr: %T", expr)
	}
}

func compileSpanMetricAttrs(attrs []string) (ch.Safe, []string) {
	var b []byte
	aliases := make([]string, len(attrs))
	for i, attr := range attrs {
		attr, alias := splitNameAlias(attr)
		aliases[i] = alias

		if i > 0 {
			b = append(b, ", "...)
		}

		b = append(b, "toString("...)
		b = tracing.AppendCHAttrExpr(b, attr)
		b = append(b, ")"...)
	}
	return ch.Safe(b), aliases
}

func compileSpanMetricAnnotations(attrs []string) ch.Safe {
	var b []byte
	for i, attr := range attrs {
		attr, alias := splitNameAlias(attr)

		if i > 0 {
			b = append(b, ", "...)
		}

		b = chschema.AppendString(b, alias)
		b = append(b, ", toString(any("...)
		b = tracing.AppendCHAttrExpr(b, attr)
		b = append(b, "))"...)
	}
	return ch.Safe(b)
}

func compileSpanMetricWhere(query string) (ch.Safe, error) {
	if !strings.HasPrefix(query, "where ") {
		query = "where " + query
	}

	parts := tql.Parse(query)
	if len(parts) != 1 {
		return "", fmt.Errorf("can't parse metric where: %q", query)
	}

	part := parts[0]
	ast, ok := part.AST.(*tql.Where)
	if !ok {
		return "", fmt.Errorf("can't parse metric where: %q", query)
	}

	where, having := tracing.AppendWhereHaving(ast, spanMetricDur)
	if len(having) > 0 {
		return "", fmt.Errorf("can't filter by agg columns: %q", having)
	}
	return ch.Safe(where), nil
}

func splitNameAlias(s string) (string, string) {
	for _, sep := range []string{" as ", " AS "} {
		if ss := strings.Split(s, sep); len(ss) == 2 {
			return ss[0], ss[1]
		}
	}
	return s, s
}
