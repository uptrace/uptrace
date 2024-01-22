package metrics

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
	"github.com/uptrace/uptrace/pkg/unixtime"
)

type DashKind string

const (
	DashKindGrid  DashKind = "grid"
	DashKindTable DashKind = "table"
)

type Dashboard struct {
	bun.BaseModel `bun:"dashboards,alias:d"`

	ID         uint64 `json:"id" bun:",pk,autoincrement"`
	ProjectID  uint32 `json:"projectId"`
	TemplateID string `json:"templateId" bun:",nullzero"`

	Name   string `json:"name"`
	Pinned bool   `json:"pinned"`

	MinInterval  unixtime.Millis `json:"minInterval"`
	TimeOffset   unixtime.Millis `json:"timeOffset"`
	GridQuery    string          `json:"gridQuery" bun:",nullzero"`
	GridMaxWidth int             `json:"gridMaxWidth" bun:",nullzero"`

	TableMetrics   []mql.MetricAlias       `json:"tableMetrics" bun:",type:jsonb,nullzero"`
	TableQuery     string                  `json:"tableQuery" bun:",nullzero"`
	TableGrouping  []string                `json:"tableGrouping" bun:",type:jsonb,nullzero"`
	TableColumnMap map[string]*TableColumn `json:"tableColumnMap" bun:",type:jsonb,nullzero"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
	UpdatedAt time.Time `json:"updatedAt" bun:",nullzero"`
}

func (d *Dashboard) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("dashboard name is required")
	}
	if err := d.validate(); err != nil {
		return fmt.Errorf("dashboard %q is invalid: %w", d.Name, err)
	}
	return nil
}

func (d *Dashboard) validate() error {
	if d.ProjectID == 0 {
		return fmt.Errorf("project id can't be zero")
	}

	if d.TableMetrics == nil {
		d.TableMetrics = make([]mql.MetricAlias, 0)
	} else if len(d.TableMetrics) > 10 {
		return errors.New("you can't use more than 10 metrics in a single table")
	}

	if d.TableQuery != "" {
		query, err := mql.ParseQueryError(d.TableQuery)
		if err != nil {
			return fmt.Errorf("can't parse table query: %w", err)
		}

		d.TableGrouping = make([]string, 0)
		for _, part := range query.Parts {
			grouping, ok := part.AST.(*ast.Grouping)
			if !ok {
				continue
			}
			for _, elem := range grouping.Elems {
				d.TableGrouping = append(d.TableGrouping, elem.Alias)
			}
		}
	}

	for _, col := range d.TableColumnMap {
		if err := col.Validate(); err != nil {
			return err
		}
	}
	if d.TableColumnMap == nil {
		d.TableColumnMap = make(map[string]*TableColumn)
	}

	if d.CreatedAt.IsZero() {
		now := time.Now()
		d.CreatedAt = now
		d.UpdatedAt = now
	}

	return nil
}

type TableColumn struct {
	MetricColumn `yaml:",inline"`

	AggFunc           string `json:"aggFunc" yaml:"agg_func,omitempty"`
	SparklineDisabled bool   `json:"sparklineDisabled" yaml:"sparkline_disabled,omitempty"`
}

func (c *TableColumn) Validate() error {
	if err := c.MetricColumn.Validate(); err != nil {
		return err
	}
	if c.AggFunc == "" {
		c.AggFunc = mql.TableFuncMedian
	}
	return nil
}

//------------------------------------------------------------------------------

func SelectDashboard(ctx context.Context, app *bunapp.App, id uint64) (*Dashboard, error) {
	dash := new(Dashboard)
	if err := app.PG.NewSelect().
		Model(dash).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return dash, nil
}

func InsertDashboard(ctx context.Context, db bun.IDB, dash *Dashboard) error {
	if _, err := db.NewInsert().
		Model(dash).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func DeleteDashboard(ctx context.Context, db bun.IDB, id uint64) error {
	if _, err := db.NewDelete().
		Model((*Dashboard)(nil)).
		Where("id = ?", id).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
