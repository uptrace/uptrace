package org

import (
	"errors"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"golang.org/x/exp/slices"
)

type PinnedFacetHandler struct {
	*bunapp.App
}

func NewPinnedFacetHandler(app *bunapp.App) *PinnedFacetHandler {
	return &PinnedFacetHandler{app}
}

func (h *PinnedFacetHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := UserFromContext(ctx)

	attrs, err := SelectPinnedFacets(ctx, h.App, user.ID)
	if err != nil {
		return err
	}

	slices.SortFunc(attrs, CoreAttrLess)

	return httputil.JSON(w, bunrouter.H{
		"attrs": attrs,
	})
}

func (h *PinnedFacetHandler) Add(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := UserFromContext(ctx)

	var in struct {
		Attr string `json:"attr"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if in.Attr == "" {
		return errors.New("attr can't be empty")
	}

	filter := &PinnedFacet{
		UserID: user.ID,
		Attr:   in.Attr,
	}
	if _, err := h.DB.NewInsert().
		Model(filter).
		On("CONFLICT (user_id, attr) DO UPDATE").
		Set("unpinned = false").
		Exec(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"attr": in.Attr,
	})
}

func (h *PinnedFacetHandler) Remove(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := UserFromContext(ctx)

	var in struct {
		Attr string `json:"attr"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<10); err != nil {
		return err
	}

	if in.Attr == "" {
		return errors.New("attr can't be empty")
	}

	facet := &PinnedFacet{
		UserID:   user.ID,
		Attr:     in.Attr,
		Unpinned: true,
	}
	if _, err := h.DB.NewInsert().
		Model(facet).
		On("CONFLICT (user_id, attr) DO UPDATE").
		Set("unpinned = true").
		Exec(ctx); err != nil {
		return err
	}

	return nil
}
