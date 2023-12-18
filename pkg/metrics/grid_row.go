package metrics

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

type GridRow struct {
	bun.BaseModel `bun:"grid_rows,alias:r"`

	ID     uint64 `json:"id" bun:",pk,autoincrement"`
	DashID uint64 `json:"dashId"`

	Title       string `json:"title"`
	Description string `json:"description" bun:",nullzero"`
	Expanded    bool   `json:"expanded"`
	Index       int    `json:"index"`

	Items []GridItem `json:"items" bun:"-"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
	UpdatedAt time.Time `json:"updatedAt" bun:",nullzero"`
}

func (row *GridRow) Validate() error {
	if row.Title == "" {
		return errors.New("row title can't be empty")
	}

	if row.CreatedAt.IsZero() {
		now := time.Now()
		row.CreatedAt = now
		row.UpdatedAt = now
	}

	return nil
}

func SelectGridRow(
	ctx context.Context, app *bunapp.App, rowID uint64,
) (*GridRow, error) {
	row := new(GridRow)

	if err := app.PG.NewSelect().
		Model(row).
		Where("id = ?", rowID).
		Scan(ctx); err != nil {
		return nil, err
	}

	return row, nil
}

func SelectGridRows(
	ctx context.Context, db bun.IDB, dashID uint64,
) ([]*GridRow, error) {
	rows := make([]*GridRow, 0)

	if err := db.NewSelect().
		Model(&rows).
		Where("dash_id = ?", dashID).
		OrderExpr("index ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	for i := range rows {
		rows[i].Index = i
	}

	return rows, nil
}

func SelectOrCreateGridRow(ctx context.Context, app *bunapp.App, dashID uint64) (*GridRow, error) {
	row := &GridRow{
		DashID:   dashID,
		Title:    "Row Title",
		Expanded: true,
	}

	if err := app.PG.NewSelect().
		Model(row).
		Where("dash_id = ?", dashID).
		OrderExpr("index ASC").
		Limit(1).
		Scan(ctx); err == nil {
		return row, nil
	} else if err != sql.ErrNoRows {
		return nil, err
	}

	if err := row.Validate(); err != nil {
		return nil, err
	}
	if _, err := app.PG.NewInsert().Model(row).Exec(ctx); err != nil {
		return nil, err
	}
	return row, nil
}

func InsertGridRow(
	ctx context.Context, db bun.IDB, gridRow *GridRow,
) error {
	if _, err := db.NewInsert().
		Model(gridRow).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
