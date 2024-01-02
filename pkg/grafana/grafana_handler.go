package grafana

import (
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
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
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()

		dsn, err := org.DSNFromRequest(req, "x-scope-orgid")
		if err != nil {
			return err
		}

		project, err := org.SelectProjectByDSN(ctx, h.App, dsn)
		if err != nil {
			return err
		}

		ctx = org.ContextWithProject(ctx, project)
		return next(w, req.WithContext(ctx))
	}
}
