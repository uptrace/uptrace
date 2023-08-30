package org

import (
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
)

type AchievementHandler struct {
	*bunapp.App
}

func NewAchievementHandler(app *bunapp.App) *AchievementHandler {
	return &AchievementHandler{
		App: app,
	}
}

func (h *AchievementHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := UserFromContext(ctx)
	project := ProjectFromContext(ctx)

	achievements, err := SelectAchievements(ctx, h.App, user.ID, project.ID)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"achievements": achievements,
	})
}
