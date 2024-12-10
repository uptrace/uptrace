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
	"go.uber.org/fx"
	"golang.org/x/crypto/bcrypt"
)

type UserHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	Conf   *bunconf.Config
	PG     *bun.DB
}

type UserHandler struct {
	*UserHandlerParams
}

func NewUserHandler(p UserHandlerParams) *UserHandler {
	return &UserHandler{&p}
}

func registerUserHandler(h *UserHandler, p bunapp.RouterParams, m *Middleware) {
	p.RouterInternalV1.WithGroup("/users", func(g *bunrouter.Group) {
		g.POST("/login", h.Login)
		g.POST("/logout", h.Logout)

		g = g.Use(m.User)

		g.GET("/current", h.Current)
	})
}

func (h *UserHandler) Current(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := UserFromContext(ctx)

	projects, err := SelectProjects(ctx, h.PG)
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

	token, err := encodeUserToken(h.Conf.SecretKey, user.Email, tokenTTL)
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
	for i := range h.Conf.Auth.Users {
		user := &h.Conf.Auth.Users[i]
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
