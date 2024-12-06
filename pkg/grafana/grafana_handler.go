package grafana

import (
	"net/http"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/org"
)

type BaseGrafanaHandler struct {
	logger *otelzap.Logger
	conf   *bunconf.Config
	pg     *bun.DB
	ch     *ch.DB
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

		fakeApp := &bunapp.App{PG: h.pg}
		project, err := org.SelectProjectByDSN(ctx, fakeApp, dsn)
		if err != nil {
			return err
		}

		ctx = org.ContextWithProject(ctx, project)
		return next(w, req.WithContext(ctx))
	}
}
