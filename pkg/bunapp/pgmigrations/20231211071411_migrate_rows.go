package pgmigrations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"github.com/uptrace/uptrace/pkg/metrics"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		var ids []uint64

		if err := db.NewSelect().
			Model((*metrics.Dashboard)(nil)).
			Column("id").
			Scan(ctx, &ids); err != nil {
			return err
		}

		for _, dashID := range ids {
			if err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
				return migrateDashboard(ctx, tx, dashID)
			}); err != nil {
				return err
			}
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}

func migrateDashboard(ctx context.Context, tx bun.Tx, dashID uint64) error {
	now := time.Now()

	var tableGauges []*DashGauge
	var gridGauges []*DashGauge

	if err := tx.NewSelect().
		Model(&tableGauges).
		Where("dash_id = ?", dashID).
		Where("dash_kind = ?", metrics.DashKindTable).
		OrderExpr("index ASC NULLS LAST, id ASC").
		Scan(ctx); err != nil {
		return err
	}

	if err := tx.NewSelect().
		Model(&gridGauges).
		Where("dash_id = ?", dashID).
		Where("dash_kind = ?", metrics.DashKindGrid).
		OrderExpr("index ASC NULLS LAST, id ASC").
		Scan(ctx); err != nil {
		return err
	}

	var cols []*BaseGridColumn

	if err := tx.NewSelect().
		Model(&cols).
		Where("dash_id = ?", dashID).
		OrderExpr("y_axis ASC, x_axis ASC, id ASC").
		Scan(ctx); err != nil {
		return err
	}

	if len(tableGauges) > 0 {
		var xAxis int
		for _, gauge := range tableGauges {
			gridItem := &metrics.BaseGridItem{
				DashID:   dashID,
				DashKind: metrics.DashKindTable,

				Title:       gauge.Name,
				Description: gauge.Description,

				XAxis: xAxis,
				YAxis: 0,

				Type: metrics.GridItemGauge,
				Params: bunutil.Params{
					Any: &metrics.GaugeGridItemParams{
						Metrics:   gauge.Metrics,
						Query:     gauge.Query,
						ColumnMap: gauge.ColumnMap,

						Template: gauge.Template,
					},
				},

				CreatedAt: gauge.CreatedAt,
				UpdatedAt: gauge.UpdatedAt,
			}
			if err := gridItem.Validate(); err != nil {
				return err
			}
			xAxis += gridItem.XAxis

			if _, err := tx.NewInsert().Model(gridItem).Exec(ctx); err != nil {
				return err
			}
		}
	}

	var rows []*metrics.GridRow

	if len(gridGauges) > 0 {
		row := &metrics.GridRow{
			DashID:   dashID,
			Title:    "Gauges",
			Expanded: true,
			Index:    len(rows),

			CreatedAt: now,
			UpdatedAt: now,
		}
		rows = append(rows, row)

		if err := row.Validate(); err != nil {
			return err
		}

		if _, err := tx.NewInsert().Model(row).Exec(ctx); err != nil {
			return err
		}

		var xAxis int
		var yAxis int
		var rowHeight int
		for _, gauge := range gridGauges {
			gridItem := &metrics.BaseGridItem{
				DashID:   dashID,
				DashKind: metrics.DashKindGrid,
				RowID:    row.ID,

				Title:       gauge.Name,
				Description: gauge.Description,

				XAxis: xAxis,
				YAxis: 0,

				Type: metrics.GridItemGauge,
				Params: bunutil.Params{
					Any: &metrics.GaugeGridItemParams{
						Metrics:   gauge.Metrics,
						Query:     gauge.Query,
						ColumnMap: gauge.ColumnMap,

						Template: gauge.Template,
					},
				},

				CreatedAt: gauge.CreatedAt,
				UpdatedAt: gauge.UpdatedAt,
			}
			if err := gridItem.Validate(); err != nil {
				return err
			}

			if xAxis+gridItem.Width > 12 {
				yAxis += rowHeight
				xAxis = 0
				rowHeight = 0
			}

			gridItem.XAxis = xAxis
			gridItem.YAxis = yAxis

			xAxis += gridItem.Width
			rowHeight = max(rowHeight, gridItem.Height)

			if _, err := tx.NewInsert().Model(gridItem).Exec(ctx); err != nil {
				return err
			}
		}
	}

	if len(cols) > 0 {
		row := &metrics.GridRow{
			DashID:   dashID,
			Title:    "General",
			Expanded: true,
			Index:    len(rows),

			CreatedAt: now,
			UpdatedAt: now,
		}
		rows = append(rows, row)

		if err := row.Validate(); err != nil {
			return err
		}

		if _, err := tx.NewInsert().Model(row).Exec(ctx); err != nil {
			return err
		}

		var xAxis int
		var yAxis int
		var rowHeight int
		for _, col := range cols {
			gridItem := &metrics.BaseGridItem{
				DashID:   dashID,
				DashKind: metrics.DashKindGrid,
				RowID:    row.ID,

				Title:       col.Name,
				Description: col.Description,

				Width:  int(col.Width),
				Height: int(col.Height * 2),

				Type:   metrics.GridItemType(col.Type),
				Params: col.Params,

				CreatedAt: col.CreatedAt,
				UpdatedAt: col.UpdatedAt,
			}

			if gridItem.Height > 30 {
				gridItem.Height -= 2
			}

			if err := gridItem.Validate(); err != nil {
				return err
			}

			if xAxis+gridItem.Width > 12 {
				yAxis += rowHeight
				xAxis = 0
				rowHeight = 0
			}

			gridItem.XAxis = xAxis
			gridItem.YAxis = yAxis

			xAxis += gridItem.Width
			rowHeight = max(rowHeight, gridItem.Height)

			if _, err := tx.NewInsert().Model(gridItem).Exec(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

type DashGauge struct {
	bun.BaseModel `bun:"dash_gauges,alias:g"`

	ID uint64 `json:"id" bun:",pk,autoincrement"`

	ProjectID uint32 `json:"projectId"`
	DashID    uint64 `json:"dashId"`

	DashKind metrics.DashKind `json:"dashKind"`
	Index    sql.NullInt64    `json:"-"`

	Name        string `json:"name"`
	Description string `json:"description"`
	Template    string `json:"template" bun:",nullzero"`

	Metrics   []mql.MetricAlias               `json:"metrics"`
	Query     string                          `json:"query"`
	ColumnMap map[string]*metrics.GaugeColumn `json:"columnMap" bun:",nullzero"`

	GridQueryTemplate string `json:"gridQueryTemplate" bun:",nullzero"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
	UpdatedAt time.Time `json:"updatedAt" bun:",nullzero"`
}

type BaseGridColumn struct {
	bun.BaseModel `bun:"dash_grid_columns,alias:g"`

	ID        uint64 `json:"id" bun:",pk,autoincrement"`
	ProjectID uint32 `json:"projectId"`
	DashID    uint64 `json:"dashId"`

	Name        string `json:"name"`
	Description string `json:"description" bun:",nullzero"`

	Width  int32 `json:"width"`
	Height int32 `json:"height"`
	XAxis  int32 `json:"xAxis"`
	YAxis  int32 `json:"yAxis"`

	GridQueryTemplate string `json:"gridQueryTemplate" bun:",nullzero"`

	Type   GridColumnType `json:"type"`
	Params bunutil.Params `json:"params"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
	UpdatedAt time.Time `json:"updatedAt" bun:",nullzero"`
}

type GridColumnType string

const (
	GridColumnChart   GridColumnType = "chart"
	GridColumnTable   GridColumnType = "table"
	GridColumnHeatmap GridColumnType = "heatmap"
)
