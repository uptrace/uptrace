package org

import (
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
)

type ProjectHandler struct {
	*Org
}

func NewProjectHandler(org *Org) *ProjectHandler {
	return &ProjectHandler{
		Org: org,
	}
}

func (h *ProjectHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	projectID, err := req.Params().Uint32("project_id")
	if err != nil {
		return err
	}

	fakeApp := &bunapp.App{PG: h.PG}
	project, err := SelectProject(ctx, fakeApp, projectID)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"project": project,
		"dsn":     BuildDSN(h.conf, project.Token),
	})
}
