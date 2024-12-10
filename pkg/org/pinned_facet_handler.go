package org

import (
	"errors"
	"net/http"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httputil"
	"go.uber.org/fx"
	"golang.org/x/exp/slices"
)

type PinnedFacetHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	Conf   *bunconf.Config
	PG     *bun.DB
}

type PinnedFacetHandler struct {
	*PinnedFacetHandlerParams
}

func NewPinnedFacetHandler(p PinnedFacetHandlerParams) *PinnedFacetHandler {
	return &PinnedFacetHandler{&p}
}

func registerPinnedFacetHandler(h *PinnedFacetHandler, p bunapp.RouterParams, m *Middleware) {
	p.RouterInternalV1.
		Use(m.User).
		WithGroup("/pinned-facets", func(g *bunrouter.Group) {
			g.GET("", h.List)
			g.POST("", h.Add)
			g.DELETE("", h.Remove)
		})
}

func (h *PinnedFacetHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := UserFromContext(ctx)

	attrs, err := SelectPinnedFacets(ctx, h.PG, user.ID)
	if err != nil {
		return err
	}

	slices.SortFunc(attrs, CompareAttrs)

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
	if _, err := h.PG.NewInsert().
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
	if _, err := h.PG.NewInsert().
		Model(facet).
		On("CONFLICT (user_id, attr) DO UPDATE").
		Set("unpinned = true").
		Exec(ctx); err != nil {
		return err
	}

	return nil
}
