package metrics

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/fx"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
)

type GridRowHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	PG     *bun.DB
}

type GridRowHandler struct {
	*GridRowHandlerParams
}

func NewGridRowHandler(p GridRowHandlerParams) *GridRowHandler {
	return &GridRowHandler{&p}
}

func registerGridRowHandler(h *GridRowHandler, p bunapp.RouterParams, m *Middleware) {
	p.RouterInternalV1.
		Use(m.UserAndProject).
		Use(m.Dashboard).
		WithGroup("/metrics/:project_id/dashboards/:dash_id/rows", func(g *bunrouter.Group) {
			g.POST("", h.Create)

			g = g.Use(h.GridRowMiddleware)

			g.GET("/:row_id", h.Show)
			g.PUT("/:row_id", h.Update)
			g.PUT("/:row_id/up", h.MoveUp)
			g.PUT("/:row_id/down", h.MoveDown)
			g.DELETE("/:row_id", h.Delete)
		})
}

func (h *GridRowHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	row := gridRowFromContext(ctx)

	gridItems := make([]*BaseGridItem, 0)

	if err := h.PG.NewSelect().
		Model(&gridItems).
		Where("dash_id = ?", row.DashID).
		Where("row_id = ?", row.ID).
		OrderExpr("y_axis ASC, x_axis ASC, id ASC").
		Scan(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"gridRow":   row,
		"gridItems": gridItems,
	})
}

type GridRowIn struct {
	Title       string `json:"title"`
	Description string `json:"description" bun:",nullzero"`
	Expanded    bool   `json:"expanded"`
}

func (in *GridRowIn) Populate(row *GridRow) error {
	row.Title = in.Title
	row.Description = in.Description
	row.Expanded = in.Expanded
	return nil
}

func (h *GridRowHandler) Create(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dash := dashFromContext(ctx)

	in := new(GridRowIn)
	if err := httputil.UnmarshalJSON(w, req, in, 100<<10); err != nil {
		return err
	}

	row := new(GridRow)
	if err := in.Populate(row); err != nil {
		return err
	}
	row.DashID = dash.ID
	if err := row.Validate(); err != nil {
		return httperror.Wrap(err)
	}

	if _, err := h.PG.NewUpdate().
		Model((*GridRow)(nil)).
		Set("index = r.index + 1").
		Where("dash_id = ?", dash.ID).
		Exec(ctx); err != nil {
		return err
	}

	if err := InsertGridRow(ctx, h.PG, row); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"gridRow": in,
	})
}

func (h *GridRowHandler) Update(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	row := gridRowFromContext(ctx)

	in := new(GridRowIn)
	if err := httputil.UnmarshalJSON(w, req, in, 100<<10); err != nil {
		return err
	}

	if err := in.Populate(row); err != nil {
		return err
	}
	row.UpdatedAt = time.Now()
	if err := row.Validate(); err != nil {
		return httperror.Wrap(err)
	}

	if _, err := h.PG.NewUpdate().
		Model(row).
		Column("title", "description", "expanded", "updated_at").
		Where("id = ?", row.ID).
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"gridRow": row,
	})
}

func (h *GridRowHandler) MoveUp(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	row := gridRowFromContext(ctx)

	if err := h.PG.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		rows, err := SelectGridRows(ctx, tx, row.DashID)
		if err != nil {
			return err
		}

		var found *GridRow
		for _, el := range rows {
			if el.ID == row.ID {
				found = el
			}
		}

		if found == nil {
			return nil
		}
		row = found

		if row.Index >= len(rows) {
			return nil
		}
		if row.Index == 0 {
			return nil
		}

		rows[row.Index-1].Index = row.Index
		row.Index--

		if _, err := h.PG.NewUpdate().
			With("_data", h.PG.NewValues(&rows)).
			Model(&rows).
			TableExpr("_data").
			Set("index = _data.index").
			Where("r.id = _data.id").
			Exec(ctx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (h *GridRowHandler) MoveDown(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	row := gridRowFromContext(ctx)

	if err := h.PG.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		rows, err := SelectGridRows(ctx, tx, row.DashID)
		if err != nil {
			return err
		}

		var found *GridRow
		for _, el := range rows {
			if el.ID == row.ID {
				found = el
			}
		}

		if found == nil {
			return nil
		}
		row = found

		if row.Index >= len(rows) {
			return nil
		}
		if row.Index == len(rows)-1 {
			return nil
		}

		rows[row.Index+1].Index = row.Index
		row.Index++

		if _, err := h.PG.NewUpdate().
			With("_data", h.PG.NewValues(&rows)).
			Model(&rows).
			TableExpr("_data").
			Set("index = _data.index").
			Where("r.id = _data.id").
			Exec(ctx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (h *GridRowHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	row := gridRowFromContext(ctx)

	if _, err := h.PG.NewDelete().
		Model(row).
		Where("id = ?", row.ID).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"gridRow": row,
	})
}

func (h *GridRowHandler) GridRowMiddleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()
		dash := dashFromContext(ctx)

		rowID, err := req.Params().Uint64("row_id")
		if err != nil {
			return err
		}

		row, err := SelectGridRow(ctx, h.PG, rowID)
		if err != nil {
			return err
		}

		if row.DashID != dash.ID {
			return org.ErrAccessDenied
		}

		ctx = context.WithValue(ctx, gridRowCtxKey{}, row)
		return next(w, req.WithContext(ctx))
	}
}

type gridRowCtxKey struct{}

func gridRowFromContext(ctx context.Context) *GridRow {
	return ctx.Value(gridRowCtxKey{}).(*GridRow)
}
