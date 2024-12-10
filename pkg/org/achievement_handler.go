package org

import (
	"net/http"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"go.uber.org/fx"
)

type AchievementHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	PG     *bun.DB
}

type AchievementHandler struct {
	*AchievementHandlerParams
}

func NewAchievementHandler(p AchievementHandlerParams) *AchievementHandler {
	return &AchievementHandler{&p}
}

func registerAchievementHandler(h *AchievementHandler, p bunapp.RouterParams, m *Middleware) {
	p.RouterInternalV1.
		Use(m.UserAndProject).
		WithGroup("/projects/:project_id", func(g *bunrouter.Group) {
			g.GET("/achievements", h.List)
		})
}

func (h *AchievementHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := UserFromContext(ctx)
	project := ProjectFromContext(ctx)

	achievements, err := SelectAchievements(ctx, h.PG, user.ID, project.ID)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"achievements": achievements,
	})
}
