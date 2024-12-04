package org

import (
	"net/http"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httputil"
)

type ProjectHandler struct {
	conf   *bunconf.Config
	logger *otelzap.Logger
	pg     *bun.DB
}

func NewProjectHandler(conf *bunconf.Config, logger *otelzap.Logger, pg *bun.DB) *ProjectHandler {
	return &ProjectHandler{
		conf:   conf,
		logger: logger,
		pg:     pg,
	}
}

func (h *ProjectHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	projectID, err := req.Params().Uint32("project_id")
	if err != nil {
		return err
	}

	fakeApp := &bunapp.App{PG: h.pg}
	project, err := SelectProject(ctx, fakeApp, projectID)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"project": project,
		"dsn":     BuildDSN(h.conf, project.Token),
	})
}
