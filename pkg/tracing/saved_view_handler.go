package tracing

import (
	"context"
	"errors"
	"net/http"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
)

type SavedViewDetails struct {
	SavedView `bun:",inherit"`

	User *org.User `json:"user" bun:"-"`
}

type SavedViewHandler struct {
	*bunapp.App
}

func NewSavedViewHandler(app *bunapp.App) *SavedViewHandler {
	return &SavedViewHandler{app}
}

func (h *SavedViewHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	project := org.ProjectFromContext(ctx)

	views := make([]*SavedViewDetails, 0)
	if err := h.PG.NewSelect().
		Model(&views).
		Where("project_id = ?", project.ID).
		OrderExpr("pinned DESC, created_at DESC").
		Limit(100).
		Scan(ctx); err != nil {
		return err
	}

	if len(views) == 0 {
		return httputil.JSON(w, bunrouter.H{
			"views": views,
		})
	}

	userIDs := make([]uint64, 0, len(views))
	userMap := make(map[uint64]*org.User)

	for _, view := range views {
		if _, ok := userMap[view.UserID]; ok {
			continue
		}
		userMap[view.UserID] = nil
		userIDs = append(userIDs, view.UserID)
	}

	var users []*org.User
	if err := h.PG.NewSelect().
		Model(&users).
		Where("id IN (?)", bun.In(userIDs)).
		Scan(ctx); err != nil {
		return err
	}

	for _, user := range users {
		userMap[user.ID] = user
	}
	for _, view := range views {
		view.User = userMap[view.UserID]
	}

	return httputil.JSON(w, bunrouter.H{
		"views": views,
	})
}

func (h *SavedViewHandler) Create(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)
	project := org.ProjectFromContext(ctx)

	var in struct {
		Name   string         `json:"name"`
		Route  string         `json:"route"`
		Params map[string]any `json:"params"`
		Query  map[string]any `json:"query"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if in.Name == "" {
		return errors.New("saved view name can't be empty")
	}
	if in.Route == "" {
		return errors.New("saved view route can't be empty")
	}

	view := &SavedView{
		UserID:    user.ID,
		ProjectID: project.ID,

		Name:   in.Name,
		Route:  in.Route,
		Params: in.Params,
		Query:  in.Query,
	}
	if _, err := h.PG.NewInsert().
		Model(view).
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"view": view,
	})
}

func (h *SavedViewHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	viewID, err := req.Params().Uint64("view_id")
	if err != nil {
		return err
	}

	if _, err := h.PG.NewDelete().
		Model(((*SavedView)(nil))).
		Where("id = ?", viewID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (h *SavedViewHandler) selectView(ctx context.Context, viewID uint64) (*SavedView, error) {
	view := new(SavedView)
	if err := h.PG.NewSelect().
		Model(view).
		Where("id = ?", viewID).
		Scan(ctx); err != nil {
		return nil, err
	}
	return view, nil
}

func (h *SavedViewHandler) Pin(w http.ResponseWriter, req bunrouter.Request) error {
	return h.updateViewPinned(w, req, true)
}

func (h *SavedViewHandler) Unpin(w http.ResponseWriter, req bunrouter.Request) error {
	return h.updateViewPinned(w, req, false)
}

func (h *SavedViewHandler) updateViewPinned(
	w http.ResponseWriter, req bunrouter.Request, pinned bool,
) error {
	ctx := req.Context()

	id, err := req.Params().Uint64("view_id")
	if err != nil {
		return err
	}

	if _, err := h.PG.NewUpdate().
		Model((*SavedView)(nil)).
		Where("id = ?", id).
		Set("pinned = ?", pinned).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}
