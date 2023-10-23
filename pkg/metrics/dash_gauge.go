package metrics

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
)

type DashGauge struct {
	bun.BaseModel `bun:"dash_gauges,alias:g"`

	ID uint64 `json:"id" bun:",pk,autoincrement"`

	ProjectID uint32     `json:"projectId"`
	DashID    uint64     `json:"dashId"`
	Dash      *Dashboard `json:"-" bun:"rel:belongs-to,on_delete:CASCADE"`

	DashKind DashKind      `json:"dashKind"`
	Index    sql.NullInt64 `json:"-"`

	Name        string `json:"name"`
	Description string `json:"description"`
	Template    string `json:"template" bun:",nullzero"`

	Metrics   []mql.MetricAlias        `json:"metrics"`
	Query     string                   `json:"query"`
	ColumnMap map[string]*MetricColumn `json:"columnMap" bun:",nullzero"`

	GridQueryTemplate string         `json:"gridQueryTemplate" bun:",nullzero"`
	ValueMappings     []ValueMapping `json:"valueMappings" bun:",nullzero"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
	UpdatedAt time.Time `json:"updatedAt" bun:",nullzero"`
}

func (g *DashGauge) FromTemplate(tpl *DashGaugeTpl) error {
	metrics, err := parseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	g.Name = tpl.Name
	g.Description = tpl.Description
	g.Template = tpl.Template

	g.Metrics = metrics
	g.Query = mql.JoinQuery(tpl.Query)
	g.ColumnMap = tpl.Columns

	g.GridQueryTemplate = tpl.GridQueryTemplate
	g.ValueMappings = tpl.ValueMappings

	return nil
}

func (g *DashGauge) Validate() error {
	if g.Name == "" {
		return fmt.Errorf("name can't be empty")
	}
	if g.ProjectID == 0 {
		return fmt.Errorf("project id can't be zero")
	}
	if g.DashKind == "" {
		return fmt.Errorf("dash kind can't be empty")
	}
	if g.Description == "" {
		return fmt.Errorf("description can't be empty")
	}

	if len(g.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}
	if len(g.Metrics) > 6 {
		return errors.New("at most 6 metrics are allowed")
	}
	for _, metric := range g.Metrics {
		if err := metric.Validate(); err != nil {
			return err
		}
	}

	if g.Query == "" {
		return fmt.Errorf("query can't be empty")
	}
	if _, err := mql.ParseQueryError(g.Query); err != nil {
		return fmt.Errorf("can't parse query: %w", err)
	}

	if g.ColumnMap == nil {
		g.ColumnMap = make(map[string]*MetricColumn)
	}

	if false {
		if _, err := mql.ParseQueryError(g.GridQueryTemplate); err != nil {
			return fmt.Errorf("can't parse grid query template: %w", err)
		}
	}
	for i := range g.ValueMappings {
		if err := g.ValueMappings[i].Validate(); err != nil {
			return fmt.Errorf("invalid value mapping: %w", err)
		}
	}

	if g.CreatedAt.IsZero() {
		now := time.Now()
		g.CreatedAt = now
		g.UpdatedAt = now
	}

	return nil
}

type ValueMapping struct {
	Op    MappingOp   `json:"op" yaml:"op"`
	Value json.Number `json:"value" yaml:"value"`
	Text  string      `json:"text" yaml:"text"`
	Color string      `json:"color" yaml:"color"`
}

type MappingOp string

const (
	MappingAny   = "any"
	MappingEqual = "eq"
	MappingLT    = "lt"
	MappingLTE   = "lte"
	MappingGT    = "gt"
	MappingGTE   = "gte"
)

func (m *ValueMapping) Validate() error {
	switch m.Op {
	case "":
		return fmt.Errorf("mapping op can't be empty")
	case MappingAny, MappingEqual, MappingLT, MappingLTE, MappingGT, MappingGTE:
		// okay
	default:
		return fmt.Errorf("invalid mapping op: %q", m.Op)
	}

	return nil
}

func SelectDashGauge(
	ctx context.Context, app *bunapp.App, dashID, gaugeID uint64,
) (*DashGauge, error) {
	gauge := new(DashGauge)
	if err := app.PG.NewSelect().
		Model(gauge).
		Where("dash_id = ?", dashID).
		Where("id = ?", gaugeID).
		Scan(ctx); err != nil {
		return nil, err
	}
	return gauge, nil
}

func SelectDashGauges(
	ctx context.Context, app *bunapp.App, dashID uint64, dashKind DashKind,
) ([]*DashGauge, error) {
	gauges := make([]*DashGauge, 0)

	q := app.PG.NewSelect().
		Model(&gauges).
		Where("dash_id = ?", dashID).
		OrderExpr("index ASC NULLS LAST, id ASC")

	if dashKind != "" {
		q = q.Where("dash_kind = ?", dashKind)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return gauges, nil
}

func InsertDashGauges(ctx context.Context, app *bunapp.App, gauges []*DashGauge) error {
	if _, err := app.PG.NewInsert().
		Model(&gauges).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
