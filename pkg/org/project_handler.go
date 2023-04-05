package org

import (
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
)

type ProjectHandler struct {
	*bunapp.App
}

func NewProjectHandler(app *bunapp.App) *ProjectHandler {
	return &ProjectHandler{
		App: app,
	}
}

func (h *ProjectHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	projectID, err := req.Params().Uint32("project_id")
	if err != nil {
		return err
	}

	project, err := SelectProject(ctx, h.App, projectID)
	if err != nil {
		return err
	}

	conf := h.Config()
	return httputil.JSON(w, bunrouter.H{
		"project": project,
		"grpc": bunrouter.H{
			"endpoint": conf.GRPCEndpoint(),
			"dsn":      conf.GRPCDsn(project.ID, project.Token),
		},
		"http": bunrouter.H{
			"endpoint": conf.HTTPEndpoint(),
			"dsn":      conf.HTTPDsn(project.ID, project.Token),
		},
	})
}
