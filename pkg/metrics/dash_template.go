package metrics

import (
	"context"
	"fmt"
	"strings"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
)

func SyncDashboards(
	ctx context.Context, app *bunapp.App, projectID uint32,
) error {
	conf := app.Config()

	dashMap, err := SelectDashboardMap(ctx, app, projectID)
	if err != nil {
		return fmt.Errorf("SelectDashboardMap failed: %w", err)
	}

	for _, tpl := range conf.Dashboards {
		if err := tpl.Validate(); err != nil {
			return err
		}

		dash, ok := dashMap[tpl.ID]
		if !ok {
			dash = &Dashboard{
				TemplateID: tpl.ID,
				ProjectID:  projectID,
			}
		}

		builder := &DashBuilder{
			dash: dash,
		}
		if err := builder.Build(tpl); err != nil {
			return fmt.Errorf("building dashboard %s failed: %w", tpl.ID, err)
		}

		if dash.ID != 0 {
			if err := DeleteDashboard(ctx, app, dash.ID); err != nil {
				return fmt.Errorf("DeleteDashboard failed: %w", err)
			}
		}

		if err := builder.Save(ctx, app); err != nil {
			return fmt.Errorf("saving dashboard %s failed: %w", tpl.ID, err)
		}
	}

	return nil
}

type DashBuilder struct {
	dash    *Dashboard
	entries []*DashEntry
}

func (b *DashBuilder) Build(tpl *bunconf.Dashboard) error {
	metrics, err := upql.ParseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	b.dash.Name = tpl.Name
	b.dash.Metrics = metrics
	b.dash.Query = strings.Join(tpl.Query, " | ")
	b.dash.Columns = tpl.Columns
	b.dash.IsTable = len(b.dash.Metrics) > 0 && b.dash.Query != ""

	for _, entry := range tpl.Entries {
		if err := b.entry(entry); err != nil {
			return err
		}
	}

	if err := b.dash.Validate(); err != nil {
		return err
	}
	for _, entry := range b.entries {
		if err := entry.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (b *DashBuilder) entry(tpl *bunconf.DashEntry) error {
	metrics, err := upql.ParseMetrics(tpl.Metrics)
	if err != nil {
		return err
	}

	b.entries = append(b.entries, &DashEntry{
		Name:    tpl.Name,
		Metrics: metrics,
		Query:   tpl.Query,
		Columns: tpl.Columns,
	})
	return nil
}

func (b *DashBuilder) Save(ctx context.Context, app *bunapp.App) error {
	if err := InsertDashboard(ctx, app, b.dash); err != nil {
		return err
	}

	for i, entry := range b.entries {
		entry.DashID = b.dash.ID
		entry.ProjectID = b.dash.ProjectID
		entry.Weight = len(b.entries) - i
	}

	if err := InsertDashEntries(ctx, app, b.entries); err != nil {
		return err
	}

	return nil
}
