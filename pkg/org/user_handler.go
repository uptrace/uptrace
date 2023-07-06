package org

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"golang.org/x/crypto/bcrypt"
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
	ctx := req.Context()
	user := UserFromContext(ctx)

	projects, err := SelectProjects(ctx, h.App)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"user":     user,
		"projects": projects,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, req bunrouter.Request) error {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
		return err
	}

	user, err := h.userByEmail(in.Email)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return httperror.BadRequest("credentials", "user with such credentials not found")
	}

	token, err := encodeUserToken(h.Config().SecretKey, user.Email, tokenTTL)
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

func (h *UserHandler) userByEmail(email string) (*User, error) {
	conf := h.Config()
	for i := range conf.Auth.Users {
		user := &conf.Auth.Users[i]
		if user.Email == email {
			return NewUserFromConfig(user)
		}
	}
	return nil, sql.ErrNoRows
}

func (h *UserHandler) Logout(w http.ResponseWriter, req bunrouter.Request) error {
	cookie := bunapp.NewCookie(h.App, req)
	cookie.Name = tokenCookieName
	cookie.Expires = time.Now().Add(-time.Hour)
	http.SetCookie(w, cookie)

	return nil
}
