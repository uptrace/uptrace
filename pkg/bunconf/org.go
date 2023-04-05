package bunconf

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`

	Email  string `yaml:"email"`
	Avatar string `yaml:"avatar"`

	NotifyByEmail bool `yaml:"notify_by_email"`
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
	ID                  uint32   `yaml:"id"`
	Name                string   `yaml:"name"`
	Token               string   `yaml:"token"`
	PinnedAttrs         []string `yaml:"pinned_attrs"`
	GroupByEnv          bool     `yaml:"group_by_env"`
	GroupFuncsByService bool     `yaml:"group_funcs_by_service"`
}
