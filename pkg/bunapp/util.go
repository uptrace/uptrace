package bunapp

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/urlstruct"
)

func WaitExitSignal() os.Signal {
	ch := make(chan os.Signal, 3)
	signal.Notify(
		ch,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	return <-ch
}

func UnmarshalValues(req bunrouter.Request, filter any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	values := req.Form

	for _, p := range req.Params().Slice() {
		values[p.Key] = []string{p.Value}
	}

	return urlstruct.Unmarshal(req.Context(), values, filter)
}

func NewCookie(req bunrouter.Request) *http.Cookie {
	cookie := &http.Cookie{
		Path:     "/",
		HttpOnly: true,
	}

	if false {
		cookie.Secure = true
		cookie.SameSite = http.SameSiteLaxMode
	}

	return cookie
}
