package org

import (
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
)

type AchievementHandler struct {
	*Org
}

func NewAchievementHandler(org *Org) *AchievementHandler {
	return &AchievementHandler{Org: org}
}

func (h *AchievementHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := UserFromContext(ctx)
	project := ProjectFromContext(ctx)

	fakeApp := &bunapp.App{PG: h.PG}
	achievements, err := SelectAchievements(ctx, fakeApp, user.ID, project.ID)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"achievements": achievements,
	})
}
