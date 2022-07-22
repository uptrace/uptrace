package grafana

import (
	"errors"
	"net/http"
	"strings"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
)

type GrafanaBaseHandler struct {
	*bunapp.App
}

func (h *GrafanaBaseHandler) Ready(w http.ResponseWriter, req bunrouter.Request) error {
	_, err := w.Write([]byte("ready\n"))
	return err
}

func (h *GrafanaBaseHandler) Echo(w http.ResponseWriter, req bunrouter.Request) error {
	_, err := w.Write([]byte("echo\n"))
	return err
}

func (h *GrafanaBaseHandler) CheckProjectAccess(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	userAndProject := org.NewAuthMiddleware(h.App)(next)

	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()

		dsn := h.uptraceDSN(req)
		if dsn == "" {
			if projectID := req.Params().ByName("project_id"); projectID != "" {
				return userAndProject(w, req)
			}
			return errors.New("either uptrace-dsn or x-scope-orgid header is required")
		}

		project, err := org.SelectProjectByDSN(ctx, h.App, dsn)
		if err != nil {
			return err
		}

		ctx = org.ContextWithProject(ctx, project)

		return next(w, req.WithContext(ctx))
	}
}

func (h *GrafanaBaseHandler) uptraceDSN(req bunrouter.Request) string {
	if s := req.Header.Get("uptrace-dsn"); s != "" {
		return s
	}
	if s := req.Header.Get("x-scope-orgid"); s != "" {
		return s
	}
	if s := req.URL.Query().Get("uptrace-dsn"); s != "" {
		return s
	}
	return ""
}

func (h *GrafanaBaseHandler) trimProjectID(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	cleanPath := func(path, projectID string) string {
		path = strings.TrimPrefix(path, "/"+projectID+"/loki/api/")
		return "/loki/api/" + path
	}

	return func(w http.ResponseWriter, req bunrouter.Request) error {
		projectID := req.Params().ByName("project_id")
		req.URL.Path = cleanPath(req.URL.Path, projectID)
		req.URL.RawPath = cleanPath(req.URL.RawPath, projectID)
		return next(w, req)
	}
}
