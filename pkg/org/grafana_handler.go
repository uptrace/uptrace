package org

import (
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

type BaseGrafanaHandler struct {
	*bunapp.App
}

func (h *BaseGrafanaHandler) Ready(w http.ResponseWriter, req bunrouter.Request) error {
	_, err := w.Write([]byte("ready\n"))
	return err
}

func (h *BaseGrafanaHandler) Echo(w http.ResponseWriter, req bunrouter.Request) error {
	_, err := w.Write([]byte("echo\n"))
	return err
}

func (h *BaseGrafanaHandler) CheckProjectAccess(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	middleware := NewMiddleware(h.App)
	userAndProject := middleware.UserAndProject(next)

	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()

		dsn, err := DSNFromRequest(req, "x-scope-orgid")
		if err != nil {
			if projectID := req.Params().ByName("project_id"); projectID != "" {
				return userAndProject(w, req)
			}
			return err
		}

		project, err := SelectProjectByDSN(ctx, h.App, dsn)
		if err != nil {
			return err
		}

		ctx = ContextWithProject(ctx, project)

		return next(w, req.WithContext(ctx))
	}
}
