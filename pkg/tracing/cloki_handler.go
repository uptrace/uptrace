package tracing

import (
	"context"
	"net/http"
	"net/url"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

type ClokiFilter struct {
	*bunapp.App `urlstruct:"-"`

	TimeFilter

	TraceID string
}

func DecodeClokiFilter(app *bunapp.App, req bunrouter.Request) (*ClokiFilter, error) {
	f := &ClokiFilter{App: app}
	if err := bunapp.UnmarshalValues(req, f); err != nil {
		return nil, err
	}
	return f, nil
}

var _ urlstruct.ValuesUnmarshaler = (*ClokiFilter)(nil)

func (f *ClokiFilter) UnmarshalValues(ctx context.Context, values url.Values) error {
	if err := f.TimeFilter.UnmarshalValues(ctx, values); err != nil {
		return err
	}
	return nil
}

//------------------------------------------------------------------------------

type ClokiHandler struct {
	*bunapp.App
}

func NewClokiHandler(app *bunapp.App) *ClokiHandler {
	return &ClokiHandler{
		App: app,
	}
}

func (h *ClokiHandler) ListSamples(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := DecodeClokiFilter(h.App, req)
	if err != nil {
		return err
	}

	samples, err := SelectClokiSamples(ctx, h.App, f)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"samples": samples,
	})
}
