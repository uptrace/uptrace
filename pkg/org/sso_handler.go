package org

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/httputil"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type SSOHandler struct {
	*bunapp.App

	methods []*SSOMethod
}

type SSOMethod struct {
	ID          string `json:"id"`
	DisplayName string `json:"name"`
	RedirectURL string `json:"url"`
}

func NewSSOHandler(app *bunapp.App, router *bunrouter.Group) *SSOHandler {
	conf := app.Config()
	methods := make([]*SSOMethod, 0)

	for _, oidcConf := range conf.Auth.OIDC {
		if oidcConf.RedirectURL == "" {
			oidcConf.RedirectURL = conf.SiteURL(fmt.Sprintf("/api/v1/sso/%s/callback", oidcConf.ID))
		}

		handler, err := NewSSOMethodHandler(app, oidcConf)
		if err != nil {
			app.Logger.Error("failed to initialize OIDC user provider", zap.Error(err))
			continue
		}

		methods = append(methods, &SSOMethod{
			ID:          oidcConf.ID,
			DisplayName: oidcConf.DisplayName,
			RedirectURL: fmt.Sprintf("/api/v1/sso/%s/start", oidcConf.ID),
		})

		router.GET(fmt.Sprintf("/%s/start", oidcConf.ID), handler.Start)
		router.GET(fmt.Sprintf("/%s/callback", oidcConf.ID), handler.Callback)
	}

	return &SSOHandler{
		App:     app,
		methods: methods,
	}
}

func (h *SSOHandler) ListMethods(w http.ResponseWriter, req bunrouter.Request) error {
	return httputil.JSON(w, bunrouter.H{
		"methods": h.methods,
	})
}

//------------------------------------------------------------------------------

type SSOMethodHandler struct {
	*bunapp.App

	conf     *bunconf.OIDCProvider
	provider *oidc.Provider
	oauth    *oauth2.Config
}

const stateCookieName = "oidc-state"

func NewSSOMethodHandler(
	app *bunapp.App, conf *bunconf.OIDCProvider,
) (*SSOMethodHandler, error) {
	provider, err := oidc.NewProvider(app.Context(), conf.IssuerURL)
	if err != nil {
		return nil, err
	}

	scopes := []string{oidc.ScopeOpenID}

	if len(conf.Scopes) > 0 {
		scopes = append(scopes, conf.Scopes...)
	} else {
		scopes = append(scopes, "profile")
	}

	oauth := &oauth2.Config{
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		RedirectURL:  conf.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	return &SSOMethodHandler{
		App:      app,
		conf:     conf,
		provider: provider,
		oauth:    oauth,
	}, nil
}

func (h *SSOMethodHandler) Start(w http.ResponseWriter, req bunrouter.Request) error {
	// Generates a 'state' token to prevent CSRF attacks.
	// It will be validated when redirected back to the app.
	state, err := randState(32)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		MaxAge:   int(time.Hour.Seconds()),
		HttpOnly: true,
	})

	http.Redirect(w, req.Request, h.oauth.AuthCodeURL(state), http.StatusFound)
	return nil
}

func (h *SSOMethodHandler) Callback(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	user, err := h.exchange(w, req)
	if err != nil {
		return err
	}

	if err := GetOrCreateUser(ctx, h.App, user); err != nil {
		return err
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

	http.Redirect(w, req.Request, "/", http.StatusFound)
	return nil
}

func (h *SSOMethodHandler) exchange(
	w http.ResponseWriter, req bunrouter.Request,
) (*User, error) {
	ctx := req.Context()

	existingState, _ := req.Cookie(stateCookieName)
	if existingState == nil {
		return nil, errors.New("oidc: no state")
	}

	if req.URL.Query().Get("state") != existingState.Value {
		return nil, errors.New("oidc: state did not match")
	}

	// Unset state to prevent repeat calls
	emptyState := &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	}
	http.SetCookie(w, emptyState)

	token, err := h.oauth.Exchange(ctx, req.URL.Query().Get("code"))
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to exchange code: %w", err)
	}

	userInfo, err := h.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to get user info: %w", err)
	}

	var claims *map[string]interface{}
	err = userInfo.Claims(&claims)

	if err != nil {
		return nil, fmt.Errorf("oidc: failed to read claims: %w", err)
	}

	emailKey := "email"
	if len(h.conf.EmailClaim) > 0 {
		emailKey = h.conf.EmailClaim
	}

	var email string
	emailValue := (*claims)[emailKey]

	switch emailValue := emailValue.(type) {
	case string:
		email = emailValue
	case nil:
		return nil, fmt.Errorf("oidc: email claim is unset: %s", emailKey)
	default:
		return nil, fmt.Errorf("oidc: email claim must be a string: %s", emailKey)
	}

	if email == "" {
		return nil, fmt.Errorf("oidc: email claim is empty: %s", emailKey)
	}

	var name string
	if len(h.conf.NameClaim) > 0 {
		name, _ = (*claims)[h.conf.NameClaim].(string)
	}
	if name == "" {
		for _, key := range []string{"name", "preferred_username"} {
			found, _ := (*claims)[key].(string)
			if found != "" {
				name = found
				break
			}
		}
	}

	return &User{
		Name:          name,
		Email:         email,
		NotifyByEmail: true,
	}, nil
}

func randState(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
