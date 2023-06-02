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
)

type DashKind string

const (
	DashGrid  DashKind = "grid"
	DashTable DashKind = "table"
)

type Dashboard struct {
	bun.BaseModel `bun:"dashboards,alias:d"`

	ID         uint64 `json:"id" bun:",pk,autoincrement"`
	ProjectID  uint32 `json:"projectId"`
	TemplateID string `json:"templateId" bun:",nullzero"`

	Name   string `json:"name"`
	Pinned bool   `json:"pinned"`

	GridQuery string `json:"gridQuery" bun:",nullzero"`

	TableMetrics   []mql.MetricAlias        `json:"tableMetrics" bun:",type:jsonb,nullzero"`
	TableQuery     string                   `json:"tableQuery" bun:",nullzero"`
	TableGrouping  []string                 `json:"tableGrouping" bun:",type:jsonb,nullzero"`
	TableColumnMap map[string]*MetricColumn `json:"tableColumnMap" bun:",type:jsonb,nullzero"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
	UpdatedAt time.Time `json:"updatedAt" bun:",nullzero"`
}

func (d *Dashboard) FromTemplate(tpl *DashboardTpl) error {
	if tpl.Schema != "v1" {
		return fmt.Errorf("unsupported template schema: %q", tpl.Schema)
	}
	if d.TemplateID != "" && d.TemplateID != tpl.ID {
		return fmt.Errorf("template id does not match: got %q, has %q", tpl.ID, d.TemplateID)
	}

	metrics, err := parseMetrics(tpl.Table.Metrics)
	if err != nil {
		return err
	}

	d.TemplateID = tpl.ID
	d.Name = tpl.Name
	d.TableMetrics = metrics
	d.TableQuery = mql.JoinQuery(tpl.Table.Query)
	d.TableColumnMap = tpl.Table.Columns

	return nil
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
	} else if len(d.TableMetrics) > 6 {
		return errors.New("you can't use more than 6 metrics in a single table")
	}

	if d.TableQuery != "" {
		query, err := mql.ParseError(d.TableQuery)
		if err != nil {
			return fmt.Errorf("can't parse query: %w", err)
		}

		d.TableGrouping = make([]string, 0)
		for _, part := range query.Parts {
			grouping, ok := part.AST.(*ast.Grouping)
			if !ok {
				continue
			}
			d.TableGrouping = append(d.TableGrouping, grouping.Names...)
		}
	}
	if d.TableColumnMap == nil {
		d.TableColumnMap = make(map[string]*MetricColumn)
	}

	if d.CreatedAt.IsZero() {
		now := time.Now()
		d.CreatedAt = now
		d.UpdatedAt = now
	}

	return nil
}

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

func InsertDashboard(ctx context.Context, app *bunapp.App, dash *Dashboard) error {
	if _, err := app.PG.NewInsert().
		Model(dash).
		On("CONFLICT DO NOTHING").
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func DeleteDashboard(ctx context.Context, app *bunapp.App, id uint64) error {
	if _, err := app.PG.NewDelete().
		Model((*Dashboard)(nil)).
		Where("id = ?", id).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
