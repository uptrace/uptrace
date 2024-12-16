package metrics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/ch/chschema"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/tracing"
	"github.com/uptrace/uptrace/pkg/tracing/tql"
)

const spanMetricDur = time.Minute

func initSpanMetrics(ctx context.Context, conf *bunconf.Config, pg *bun.DB, ch *ch.DB) error {
	for i := range conf.MetricsFromSpans {
		metric := &conf.MetricsFromSpans[i]

		if metric.Name == "" {
			return fmt.Errorf("metric name can't be empty")
		}

		if err := createSpanMetric(ctx, conf, pg, ch, metric); err != nil {
			return fmt.Errorf("createSpanMetric %q failed: %w", metric.Name, err)
		}
	}
	return nil
}

func createSpanMetric(
	ctx context.Context,
	conf *bunconf.Config,
	pg *bun.DB,
	ch *ch.DB,
	metric *bunconf.SpanMetric,
) error {
	if metric.Instrument == "" {
		return fmt.Errorf("metric instrument can't be empty")
	}

	if err := createSpanMetricMeta(ctx, conf, pg, metric); err != nil {
		return fmt.Errorf("createSpanMetricMeta failed: %w", err)
	}
	if err := createMatView(ctx, conf, ch, metric); err != nil {
		return fmt.Errorf("createMatView failed: %w", err)
	}
	return nil
}

func createSpanMetricMeta(ctx context.Context, conf *bunconf.Config, pg *bun.DB, metric *bunconf.SpanMetric) error {
	projects := conf.ProjectGateway
	for i := range projects {
		project := &projects[i]

		attrKeys := make([]string, len(metric.Attrs))
		for i, attr := range metric.Attrs {
			attrKeys[i], _ = splitNameAlias(attr)
		}

		if err := UpsertMetric(ctx, pg, &Metric{
			ProjectID:   project.ID,
			Name:        metric.Name,
			Description: metric.Description,
			Unit:        bunconv.NormUnit(metric.Unit),
			Instrument:  Instrument(metric.Instrument),
			AttrKeys:    attrKeys,
		}); err != nil {
			return err
		}
	}
	return nil
}

func createMatView(ctx context.Context, conf *bunconf.Config, chdb *ch.DB, metric *bunconf.SpanMetric) error {
	viewName := metric.ViewName()

	if _, err := chdb.NewDropView().
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

	q := chdb.NewCreateView().
		Materialized().
		View(viewName).
		OnCluster(conf.CHSchema.Cluster).
		ToExpr("?DB.datapoint_minutes").
		ColumnExpr("s.project_id").
		ColumnExpr("? AS metric", metric.Name).
		ColumnExpr("toStartOfMinute(s.time) AS time").
		ColumnExpr("? AS instrument", metric.Instrument).
		TableExpr("?DB.spans_index AS s").
		GroupExpr("s.project_id, toStartOfMinute(s.time)").
		Setting("allow_experimental_analyzer = 1")

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
		where, having, err := compileSpanMetricWhere(metric.Where)
		if err != nil {
			return err
		}
		if where != "" {
			q = q.Where(string(where))
		}
		if having != "" {
			return errors.New("having not supported")
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
			ColumnExpr("quantilesBFloat16State(0.5)(toFloat32(?)) AS histogram", valueExpr).
			ColumnExpr("min(toFloat64(?)) AS min", valueExpr).
			ColumnExpr("max(toFloat64(?)) AS max", valueExpr)
	default:
		return fmt.Errorf("unsupported instrument: %q", metric.Instrument)
	}

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func compileSpanMetricValue(value string) (ch.Safe, error) {
	parts, err := tql.ParseQueryError(value)
	if err != nil {
		return "", err
	}

	part := parts[0]
	sel, ok := part.AST.(*tql.Selector)
	if !ok {
		return "", fmt.Errorf("expected a column, got %T", part.AST)
	}

	if len(sel.Columns) != 1 {
		return "", fmt.Errorf("expected 1 column, got %d", len(sel.Columns))
	}
	col := &sel.Columns[0]

	b, err := tracing.AppendCHExpr(nil, col.Value, spanMetricDur)
	if err != nil {
		return "", err
	}

	return ch.Safe(b), nil
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
		b = tracing.AppendCHAttr(b, tql.Attr{Name: attr})
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
		b = tracing.AppendCHAttr(b, tql.Attr{Name: attr})
		b = append(b, "))"...)
	}
	return ch.Safe(b)
}

func compileSpanMetricWhere(query string) (ch.Safe, ch.Safe, error) {
	if !strings.HasPrefix(query, "where ") {
		query = "where " + query
	}

	parts, err := tql.ParseQueryError(query)
	if err != nil {
		return "", "", err
	}

	if len(parts) != 1 {
		return "", "", fmt.Errorf("expected 1 part, got %d", len(parts))
	}

	part := parts[0]
	ast, ok := part.AST.(*tql.Where)
	if !ok {
		return "", "", fmt.Errorf("expected a where clause, got %T", part.AST)
	}

	where, having, err := tracing.AppendWhereHaving(ast, spanMetricDur)
	if err != nil {
		return "", "", err
	}
	return ch.Safe(where), ch.Safe(having), nil
}

func splitNameAlias(s string) (string, string) {
	for _, sep := range []string{" as ", " AS "} {
		if ss := strings.Split(s, sep); len(ss) == 2 {
			return ss[0], ss[1]
		}
	}
	return s, s
}
