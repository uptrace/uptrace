package metrics

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
)

type DashEntry struct {
	bun.BaseModel `bun:"dash_entries,alias:e"`

	ID        uint64     `json:"id,string" bun:",pk,autoincrement"`
	DashID    uint64     `json:"dashId,string"`
	Dash      *Dashboard `json:"-" bun:"rel:belongs-to,on_delete:CASCADE"`
	ProjectID uint32     `json:"projectId"`

	Name        string `json:"name"`
	Description string `json:"description,nullzero"`
	Weight      int    `json:"weight"`
	ChartType   string `json:"chartType" bun:",nullzero,default:'line'"`

	Metrics []upql.Metric            `json:"metrics"`
	Query   string                   `json:"query"`
	Columns map[string]*MetricColumn `json:"columnMap" bun:",nullzero"`
}

func (e *DashEntry) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("entry name is required")
	}
	if len(e.Metrics) == 0 {
		return fmt.Errorf("entry requires at least one metric")
	}
	if e.Query == "" {
		return fmt.Errorf("entry query is required")
	}
	return nil
}

func SelectDashEntries(
	ctx context.Context, app *bunapp.App, dash *Dashboard,
) ([]*DashEntry, error) {
	var entries []*DashEntry
	if err := app.DB.NewSelect().
		Model(&entries).
		Where("dash_id = ?", dash.ID).
		OrderExpr("weight DESC, id ASC").
		Scan(ctx); err != nil {
		return nil, err
	}
	return entries, nil
}

func InsertDashEntries(ctx context.Context, app *bunapp.App, entries []*DashEntry) error {
	if len(entries) == 0 {
		return nil
	}

	for _, entry := range entries {
		if entry.Columns == nil {
			entry.Columns = make(map[string]*MetricColumn)
		}
	}

	if _, err := app.DB.NewInsert().
		Model(&entries).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
