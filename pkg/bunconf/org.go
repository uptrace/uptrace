package bunconf

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

type User struct {
	ID       uint64 `yaml:"id" json:"id"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"-"`

	Email  string `yaml:"email" json:"email"`
	Avatar string `yaml"avatar" json:"avatar"`
}

func (u *User) Gravatar() string {
	email := u.Email
	if email == "" {
		email = u.Username
	}
	return fmt.Sprintf("https://gravatar.com/avatar/%s?d=identicon", md5s(email))
}

func md5s(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

type CloudflareProvider struct {
	TeamURL  string `yaml:"team_url" json:"team_url"`
	Audience string `yaml:"audience" json:"audience"`
}

type OIDCProvider struct {
	ID           string   `yaml:"id" json:"id"`
	DisplayName  string   `yaml:"display_name" json:"display_name"`
	IssuerURL    string   `yaml:"issuer_url" json:"issuer_url"`
	ClientID     string   `yaml:"client_id" json:"client_id"`
	ClientSecret string   `yaml:"client_secret" json:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url" json:"redirect_url"`
	Scopes       []string `yaml:"scopes" json:"scopes"`
	Claim        string   `yaml:"claim" json:"claim"`
}

type Project struct {
	ID                  uint32   `yaml:"id" json:"id"`
	Name                string   `yaml:"name" json:"name"`
	Token               string   `yaml:"token" json:"token"`
	PinnedAttrs         []string `yaml:"pinned_attrs" json:"pinnedAttrs"`
	GroupByEnv          bool     `yaml:"group_by_env" json:"groupByEnv"`
	GroupFuncsByService bool     `yaml:"group_funcs_by_service" json:"groupFuncsByService"`
}
