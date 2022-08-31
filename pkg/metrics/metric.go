package metrics

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"github.com/uptrace/uptrace/pkg/bunapp"
)

type Metric struct {
	bun.BaseModel `bun:"metrics,alias:m"`

	ID        uint64 `json:"id,string" bun:",pk,autoincrement"`
	ProjectID uint32 `json:"projectId"`

	Name        string `json:"name"`
	Description string `json:"description"`
	Unit        string `json:"unit" bun:",nullzero"`
	Instrument  string `json:"instrument"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero,default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updatedAt" bun:",nullzero,default:CURRENT_TIMESTAMP"`
}

func newInvalidMetric(projectID uint32, metricName string) *Metric {
	return &Metric{
		ProjectID:  projectID,
		Name:       metricName,
		Instrument: InvalidInstrument,
	}
}

func SelectMetrics(ctx context.Context, app *bunapp.App, projectID uint32) ([]*Metric, error) {
	var metrics []*Metric
	if err := app.DB.NewSelect().
		Model(&metrics).
		Where("project_id = ?", projectID).
		OrderExpr("name ASC").
		Limit(10000).
		Scan(ctx); err != nil {
		return nil, err
	}
	return metrics, nil
}

func SelectMetricMap(
	ctx context.Context, app *bunapp.App, projectID uint32,
) (map[string]*Metric, error) {
	metrics, err := SelectMetrics(ctx, app, projectID)
	if err != nil {
		return nil, err
	}

	m := make(map[string]*Metric, len(metrics))

	for _, metric := range metrics {
		m[metric.Name] = metric
	}

	return m, nil
}

func SelectMetric(ctx context.Context, app *bunapp.App, id uint64) (*Metric, error) {
	metric := new(Metric)
	if err := app.DB.NewSelect().
		Model(metric).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return metric, nil
}

func SelectMetricByName(
	ctx context.Context, app *bunapp.App, projectID uint32, name string,
) (*Metric, error) {
	metric := new(Metric)
	if err := app.DB.NewSelect().
		Model(metric).
		Where("project_id = ?", projectID).
		Where("name = ?", name).
		Scan(ctx); err != nil {
		return nil, err
	}
	return metric, nil
}

func UpsertMetric(ctx context.Context, app *bunapp.App, m *Metric) (inserted bool, _ error) {
	m.CreatedAt = time.Now().Add(-time.Second)
	m.UpdatedAt = m.CreatedAt

	if _, err := app.DB.NewInsert().
		Model(m).
		On("CONFLICT (project_id, name) DO UPDATE").
		Set("description = EXCLUDED.description").
		Set("unit = EXCLUDED.unit").
		Set("instrument = EXCLUDED.instrument").
		Set("updated_at = EXCLUDED.updated_at").
		Returning("id, created_at, updated_at").
		Exec(ctx); err != nil {
		return false, err
	}

	inserted = m.UpdatedAt.Equal(m.CreatedAt)
	return inserted, nil
}
