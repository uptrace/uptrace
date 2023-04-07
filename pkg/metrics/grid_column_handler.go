package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/segmentio/encoding/json"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
)

type GridColumnHandler struct {
	*bunapp.App
}

func NewGridColumnHandler(app *bunapp.App) *GridColumnHandler {
	return &GridColumnHandler{
		App: app,
	}
}

func (h *GridColumnHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	gridCol := gridColumnFromContext(ctx).Base()

	if _, err := h.App.PG.
		NewDelete().
		Model(gridCol).
		Where("id = ?", gridCol.ID).
		Where("project_id = ?", gridCol.ProjectID).
		Where("dash_id = ?", gridCol.DashID).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"column": gridCol,
	})
}

type GridColumnIndex struct {
	ID     uint64 `json:"id"`
	Width  int32  `json:"width"`
	Height int32  `json:"height"`
	XAxis  int32  `json:"xAxis"`
	YAxis  int32  `json:"yAxis"`
}

func (h *GridColumnHandler) UpdateOrder(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)
	dash := dashFromContext(ctx)

	var data []GridColumnIndex
	if err := httputil.UnmarshalJSON(w, req, &data, 10<<10); err != nil {
		return err
	}

	var grid []BaseGridColumn

	if _, err := h.PG.NewUpdate().
		With("_data", h.PG.NewValues(&data)).
		Model(&grid).
		TableExpr("_data").
		Set("width = _data.width").
		Set("height = _data.height").
		Set("x_axis = _data.x_axis").
		Set("y_axis = _data.y_axis").
		Where("g.id = _data.id").
		Where("g.dash_id = ?", dash.ID).
		Where("g.project_id = ?", project.ID).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"grid": grid,
	})
}

//------------------------------------------------------------------------------

type GridColumnIn struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	GridQueryTemplate string `json:"gridQueryTemplate"`

	Type   GridColumnType  `json:"type"`
	Params json.RawMessage `json:"params"`
}

func (in *GridColumnIn) Validate(
	ctx context.Context, app *bunapp.App, baseCol *BaseGridColumn,
) error {
	baseCol.Name = in.Name
	baseCol.Description = in.Description
	baseCol.GridQueryTemplate = in.GridQueryTemplate
	baseCol.UpdatedAt = time.Now()

	switch in.Type {
	case GridColumnChart:
		col := &ChartGridColumn{
			BaseGridColumn: baseCol,
		}
		if err := json.Unmarshal(in.Params, &col.Params); err != nil {
			return err
		}

		col.Type = in.Type
		baseCol.Params.Any = &col.Params

		return col.Validate()
	case GridColumnTable:
		col := &TableGridColumn{
			BaseGridColumn: baseCol,
		}
		if err := json.Unmarshal(in.Params, &col.Params); err != nil {
			return err
		}

		col.Type = in.Type
		baseCol.Params.Any = &col.Params

		return col.Validate()
	case GridColumnHeatmap:
		col := &HeatmapGridColumn{
			BaseGridColumn: baseCol,
		}
		if err := json.Unmarshal(in.Params, &col.Params); err != nil {
			return err
		}

		col.Type = in.Type
		baseCol.Params.Any = &col.Params

		return col.Validate()
	default:
		return fmt.Errorf("unknown grid column type: %q", in.Type)
	}
}

func (h *GridColumnHandler) Create(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)
	dash := dashFromContext(ctx)

	grid, err := SelectBaseGridColumns(ctx, h.App, dash.ID)
	if err != nil {
		return err
	}
	if len(grid) >= 20 {
		return fmt.Errorf("dashboard grid can't have more than 20 columns")
	}

	in := new(GridColumnIn)
	if err := httputil.UnmarshalJSON(w, req, in, 100<<10); err != nil {
		return err
	}

	gridCol := new(BaseGridColumn)
	gridCol.ProjectID = project.ID
	gridCol.DashID = dash.ID

	if err := in.Validate(ctx, h.App, gridCol); err != nil {
		return httperror.Wrap(err)
	}

	if _, err := h.PG.NewInsert().
		Model(gridCol).
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"column": gridCol,
	})
}

func (h *GridColumnHandler) Update(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	gridCol := gridColumnFromContext(ctx).Base()

	in := new(GridColumnIn)
	if err := httputil.UnmarshalJSON(w, req, in, 100<<10); err != nil {
		return err
	}

	if err := in.Validate(ctx, h.App, gridCol); err != nil {
		return httperror.Wrap(err)
	}

	if _, err := h.PG.NewUpdate().
		Model(gridCol).
		Column("name", "description", "grid_query_template", "params", "updated_at").
		Where("id = ?", gridCol.ID).
		Where("project_id = ?", gridCol.ProjectID).
		Where("dash_id = ?", gridCol.DashID).
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"column": gridCol,
	})
}
