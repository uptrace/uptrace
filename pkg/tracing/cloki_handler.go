package tracing

import (
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
)

type ClokiHandler struct {
	*bunapp.App
}

func NewClokiHandler(app *bunapp.App) *ClokiHandler {
	return &ClokiHandler{
		App: app,
	}
}

func (h *ClokiHandler) Test(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	// req.Request is *http.Request

	if err := h.CH().Ping(ctx); err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"foo": "bar",
		"now": time.Now(),
	})
}
