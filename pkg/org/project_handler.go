package org

import (
	"net/http"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httputil"
	"go.uber.org/fx"
)

type ProjectHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	Conf   *bunconf.Config
	PG     *bun.DB
}

type ProjectHandler struct {
	*ProjectHandlerParams
}

func NewProjectHandler(p ProjectHandlerParams) *ProjectHandler {
	return &ProjectHandler{&p}
}

func registerProjectHandler(h *ProjectHandler, p bunapp.RouterParams, m *Middleware) {
	p.RouterInternalV1.
		Use(m.User).
		WithGroup("/projects/:project_id", func(g *bunrouter.Group) {
			g.GET("", h.Show)
		})
}

func (h *ProjectHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	projectID, err := req.Params().Uint32("project_id")
	if err != nil {
		return err
	}

	project, err := SelectProject(ctx, h.PG, projectID)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"project": project,
		"dsn":     BuildDSN(h.Conf, project.Token),
	})
}
