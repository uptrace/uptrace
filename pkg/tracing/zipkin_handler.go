package tracing

import (
	"encoding/json"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

type ZipkinHandler struct {
	*bunapp.App
}

func NewZipkinHandler(app *bunapp.App) *ZipkinHandler {
	return &ZipkinHandler{
		App: app,
	}
}

type ZipkinSpan struct {
	ID             string         `json:"id"`
	TraceID        string         `json:"traceId"`
	ParentID       string         `json:"parentId"`
	Name           string         `json:"name"`
	Timestamp      int64          `json:"timestamp"`
	Duration       int64          `json:"duration"`
	Kind           string         `json:"kind"`
	LocalEndpoint  ZipkinEndpoint `json:"localEndpoint"`
	RemoteEndpoint ZipkinEndpoint `json:"remoteEndpoint"`
	Tags           AttrMap        `json:"tags"`
}

type ZipkinEndpoint struct {
	ServiceName string `json:"serviceName"`
	IPV4        string `json:"string"`
	Port        int    `json:"port"`
}

func (h *ZipkinHandler) PostSpans(w http.ResponseWriter, req bunrouter.Request) error {
	dec := json.NewDecoder(req.Body)

	var spans []ZipkinSpan
	if err := dec.Decode(&spans); err != nil {
		return err
	}

	w.WriteHeader(http.StatusAccepted)
	return nil
}
