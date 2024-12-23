package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/segmentio/encoding/json"

	"go.uber.org/fx"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
)

type GridItemHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	PG     *bun.DB
}

type GridItemHandler struct {
	*GridItemHandlerParams
}

func NewGridItemHandler(p GridItemHandlerParams) *GridItemHandler {
	return &GridItemHandler{&p}
}

func registerGridItemHandler(h *GridItemHandler, p bunapp.RouterParams, m *Middleware) {
	p.RouterInternalV1.
		Use(m.UserAndProject).
		Use(m.Dashboard).
		WithGroup("/metrics/:project_id/dashboards/:dash_id/grid", func(g *bunrouter.Group) {
			g.POST("", h.Create)
			g.PUT("/layout", h.UpdateLayout)

			g = g.Use(h.GridItemMiddleware)

			g.PUT("/:row_id", h.Update)
			g.DELETE("/:row_id", h.Delete)
		})
}

type GridItemPos struct {
	ID     uint64 `json:"id"`
	Width  int32  `json:"width"`
	Height int32  `json:"height"`
	XAxis  int32  `json:"xAxis"`
	YAxis  int32  `json:"yAxis"`
}

func (h *GridItemHandler) UpdateLayout(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	var in struct {
		Items []GridItemPos `json:"items"`
		RowID uint64        `json:"rowId"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if len(in.Items) == 0 {
		return nil
	}

	q := h.PG.NewUpdate().
		With("_data", h.PG.NewValues(&in.Items)).
		Model((*BaseGridItem)(nil)).
		TableExpr("_data").
		Set("width = _data.width").
		Set("height = _data.height").
		Set("x_axis = _data.x_axis").
		Set("y_axis = _data.y_axis").
		Where("g.id = _data.id").
		Where("g.dash_id = ?", dash.ID)

	if in.RowID != 0 {
		q = q.Set("row_id = ?", in.RowID)
	} else {
		q = q.Set("row_id = NULL")
	}

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------

type GridItemIn struct {
	DashKind DashKind `json:"dashKind"`

	Title       string `json:"title"`
	Description string `json:"description"`

	Width  int `json:"width"`
	Height int `json:"height"`

	Type   GridItemType    `json:"type"`
	Params json.RawMessage `json:"params"`
}

func (in *GridItemIn) Populate(baseItem *BaseGridItem) error {
	baseItem.DashKind = in.DashKind
	baseItem.Title = in.Title
	baseItem.Description = in.Description
	baseItem.Width = in.Width
	baseItem.Height = in.Height
	baseItem.UpdatedAt = time.Now()

	switch in.Type {
	case GridItemGauge:
		baseItem.Type = in.Type
		baseItem.Params.Any = new(GaugeGridItemParams)
		return in.populate(baseItem)
	case GridItemChart:
		baseItem.Type = in.Type
		baseItem.Params.Any = new(ChartGridItemParams)
		return in.populate(baseItem)
	case GridItemTable:
		baseItem.Type = in.Type
		baseItem.Params.Any = new(TableGridItemParams)
		return in.populate(baseItem)
	case GridItemHeatmap:
		baseItem.Type = in.Type
		baseItem.Params.Any = new(HeatmapGridItemParams)
		return in.populate(baseItem)
	default:
		return fmt.Errorf("unsupported grid item type: %q", in.Type)
	}
}

func (in *GridItemIn) populate(item *BaseGridItem) error {
	if err := json.Unmarshal(in.Params, item.Params.Any); err != nil {
		return err
	}
	return nil
}

func (h *GridItemHandler) Create(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	grid, err := SelectBaseGridItems(ctx, h.PG, dash.ID)
	if err != nil {
		return err
	}
	if len(grid) >= 20 {
		return fmt.Errorf("dashboard grid can't have more than 20 columns")
	}

	in := new(GridItemIn)
	if err := httputil.UnmarshalJSON(w, req, in, 100<<10); err != nil {
		return err
	}

	gridItem := new(BaseGridItem)
	if err := in.Populate(gridItem); err != nil {
		return httperror.Wrap(err)
	}

	gridItem.DashID = dash.ID
	if gridItem.DashKind == DashKindGrid {
		gridRow, err := SelectOrCreateGridRow(ctx, h.PG, dash.ID)
		if err != nil {
			return err
		}
		gridItem.RowID = gridRow.ID
	}

	if err := gridItem.Validate(); err != nil {
		return httperror.Wrap(err)
	}

	if err := h.PG.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := h.PG.NewUpdate().
			Model((*BaseGridItem)(nil)).
			Set("y_axis = ?", gridItem.Height).
			Where("dash_id = ?", gridItem.DashID).
			Where("dash_kind = ?", gridItem.DashKind).
			Where("row_id = ?", gridItem.RowID).
			Where("x_axis >= 0").
			Where("x_axis < ?", gridItem.Width).
			Where("y_axis >= 0").
			Where("y_axis < ?", gridItem.Height).
			Exec(ctx); err != nil {
			return err
		}

		if _, err := h.PG.NewInsert().
			Model(gridItem).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"gridItem": gridItem,
	})
}

func (h *GridItemHandler) Update(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	gridItem := gridItemFromContext(ctx).Base()

	in := new(GridItemIn)
	if err := httputil.UnmarshalJSON(w, req, in, 100<<10); err != nil {
		return err
	}

	if err := in.Populate(gridItem); err != nil {
		return httperror.Wrap(err)
	}
	if err := gridItem.Validate(); err != nil {
		return httperror.Wrap(err)
	}
	gridItem.UpdatedAt = time.Now()

	if _, err := h.PG.NewUpdate().
		Model(gridItem).
		Column("title", "description", "params", "updated_at").
		Where("id = ?", gridItem.ID).
		Where("dash_id = ?", gridItem.DashID).
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"gridItem": gridItem,
	})
}

func (h *GridItemHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	gridItem := gridItemFromContext(ctx).Base()

	if _, err := h.PG.NewDelete().
		Model(gridItem).
		Where("id = ?", gridItem.ID).
		Where("dash_id = ?", gridItem.DashID).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"gridItem": gridItem,
	})
}

func (h *GridItemHandler) GridItemMiddleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()
		dash := dashFromContext(ctx)

		itemID, err := req.Params().Uint64("row_id")
		if err != nil {
			return err
		}

		item, err := SelectGridItem(ctx, h.PG, itemID)
		if err != nil {
			return err
		}

		if item.Base().DashID != dash.ID {
			return org.ErrAccessDenied
		}

		ctx = context.WithValue(ctx, gridItemCtxKey{}, item)
		return next(w, req.WithContext(ctx))
	}
}

type gridItemCtxKey struct{}

func gridItemFromContext(ctx context.Context) GridItem {
	return ctx.Value(gridItemCtxKey{}).(GridItem)
}

func chartGridItemFromContext(ctx context.Context) (*ChartGridItem, error) {
	anyGridItem := ctx.Value(gridItemCtxKey{})
	gridItem, ok := anyGridItem.(*ChartGridItem)
	if !ok {
		return nil, fmt.Errorf("got %T, expected *ChartGridItem", anyGridItem)
	}
	return gridItem, nil
}

func tableGridItemFromContext(ctx context.Context) (*TableGridItem, error) {
	anyGridItem := ctx.Value(gridItemCtxKey{})
	gridItem, ok := anyGridItem.(*TableGridItem)
	if !ok {
		return nil, fmt.Errorf("got %T, expected *TableGridItem", anyGridItem)
	}
	return gridItem, nil
}

func heatmapGridItemFromContext(ctx context.Context) (*HeatmapGridItem, error) {
	anyGridItem := ctx.Value(gridItemCtxKey{})
	gridItem, ok := anyGridItem.(*HeatmapGridItem)
	if !ok {
		return nil, fmt.Errorf("got %T, expected *HeatmapGridItem", anyGridItem)
	}
	return gridItem, nil
}
