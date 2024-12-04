package org

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	conf   *bunconf.Config
	logger *otelzap.Logger
	pg     *bun.DB
}

func NewUserHandler(conf *bunconf.Config, logger *otelzap.Logger, pg *bun.DB) *UserHandler {
	return &UserHandler{
		conf:   conf,
		logger: logger,
		pg:     pg,
	}
}

func (h *UserHandler) Current(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := UserFromContext(ctx)

	fakeApp := &bunapp.App{PG: h.pg}
	projects, err := SelectProjects(ctx, fakeApp)
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

	token, err := encodeUserToken(h.conf.SecretKey, user.Email, tokenTTL)
	if err != nil {
		return err
	}

	cookie := bunapp.NewCookie(req)
	cookie.Name = tokenCookieName
	cookie.Value = token
	cookie.MaxAge = int(tokenTTL.Seconds())
	http.SetCookie(w, cookie)

	return nil
}

func (h *UserHandler) userByEmail(email string) (*User, error) {
	conf := h.conf
	for i := range conf.Auth.Users {
		user := &conf.Auth.Users[i]
		if user.Email == email {
			return NewUserFromConfig(user)
		}
	}
	return nil, sql.ErrNoRows
}

func (h *UserHandler) Logout(w http.ResponseWriter, req bunrouter.Request) error {
	cookie := bunapp.NewCookie(req)
	cookie.Name = tokenCookieName
	cookie.Expires = time.Now().Add(-time.Hour)
	http.SetCookie(w, cookie)

	return nil
}
