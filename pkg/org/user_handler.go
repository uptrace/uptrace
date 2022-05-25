package org

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
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
	user := UserFromRequest(h.App, req)
	if user == nil {
		return ErrUnauthorized
	}

	cfg := h.Config()

	return httputil.JSON(w, bunrouter.H{
		"user":     user,
		"projects": cfg.Projects,
		"hasLoki":  cfg.Loki.Addr != "",
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, req bunrouter.Request) error {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
		return err
	}

	user := findUserByPassword(h.App, in.Username, in.Password)
	if user == nil {
		return sql.ErrNoRows
	}

	token, err := encodeUserToken(h.App, user.ID, tokenTTL)
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
