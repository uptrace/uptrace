package grafana

import (
	"encoding/json"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/org"
	"go.uber.org/fx"
)

const (
	protobufContentType = "application/protobuf"
	jsonContentType     = "application/json"
)

var Module = fx.Module("grafana",
	fx.Provide(
		fx.Private,
		org.NewMiddleware,
		NewTempoHandler,
		NewPromHandler,
	),
	fx.Invoke(
		registerTempoHandler,
		registerPromHandler,
	),
)

func promErrorHandler(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		err := next(w, req)
		if err == nil {
			return nil
		}
		switch err := err.(type) {
		case *promError:
			return err
		default:
			return newPromError(err)
		}
	}
}

//------------------------------------------------------------------------------

type promError struct {
	Wrapped error `json:"error"`
}

var _ httperror.Error = (*promError)(nil)

func newPromError(err error) *promError {
	return &promError{
		Wrapped: err,
	}
}

func (e *promError) Error() string {
	if e.Wrapped == nil {
		return ""
	}
	return e.Wrapped.Error()
}

func (e *promError) Unwrap() error {
	return e.Wrapped
}

func (e *promError) HTTPStatusCode() int {
	return http.StatusBadRequest
}

func (e *promError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"status":    "error",
		"errorType": "error",
		"error":     e.Error(),
	})
}
