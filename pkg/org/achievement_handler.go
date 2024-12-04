package org

import (
	"net/http"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
)

type AchievementHandler struct {
	pg *bun.DB
}

func NewAchievementHandler(pg *bun.DB) *AchievementHandler {
	return &AchievementHandler{pg: pg}
}

func (h *AchievementHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := UserFromContext(ctx)
	project := ProjectFromContext(ctx)

	fakeApp := &bunapp.App{PG: h.pg}
	achievements, err := SelectAchievements(ctx, fakeApp, user.ID, project.ID)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"achievements": achievements,
	})
}
