package org

import (
	"net/http"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
)

type UserHandler struct {
	*bunapp.App
}

func NewUserHandler(app *bunapp.App) *UserHandler {
	return &UserHandler{
		App: app,
	}
}

func (h *UserHandler) Current(w http.ResponseWriter, req bunrouter.Request) error {
	user, err := UserFromContext(req.Context())
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"user":     user,
		"projects": h.Config().Projects,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, req bunrouter.Request) error {
	if len(h.Config().Auth.Users) == 0 {
		return httperror.InternalServerError("Configure some users before continuing")
	}

	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
		return err
	}

	user := findUserByPassword(h.App, in.Username, in.Password)
	if user == nil {
		return httperror.BadRequest("user with such credentials not found")
	}

	token, err := encodeUserToken(h.Config().SecretKey, user.Username, tokenTTL)
	if err != nil {
		return err
	}

	cookie := bunapp.NewCookie(h.App, req)
	cookie.Name = tokenCookieName
	cookie.Value = token
	cookie.MaxAge = int(tokenTTL.Seconds())
	http.SetCookie(w, cookie)

	return nil
}

func (h *UserHandler) Logout(w http.ResponseWriter, req bunrouter.Request) error {
	cookie := bunapp.NewCookie(h.App, req)
	cookie.Name = tokenCookieName
	cookie.Expires = time.Now().Add(-time.Hour)
	http.SetCookie(w, cookie)

	return nil
}
