package bunconf

type User struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`

	Name   string `yaml:"name"`
	Avatar string `yaml:"avatar"`

	NotifyByEmail bool `yaml:"notify_by_email"`
}

type CloudflareProvider struct {
	TeamURL  string `yaml:"team_url"`
	Audience string `yaml:"audience"`
}

type OIDCProvider struct {
	ID           string   `yaml:"id"`
	DisplayName  string   `yaml:"display_name"`
	IssuerURL    string   `yaml:"issuer_url"`
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`
	EmailClaim   string   `yaml:"claim"`
	NameClaim    string   `yaml:"name_claim"`
}

type Project struct {
	ID                  uint32   `yaml:"id"`
	Name                string   `yaml:"name"`
	Token               string   `yaml:"token"`
	PinnedAttrs         []string `yaml:"pinned_attrs"`
	GroupByEnv          bool     `yaml:"group_by_env"`
	GroupFuncsByService bool     `yaml:"group_funcs_by_service"`
	PromCompat          bool     `yaml:"prom_compat"`
	ForceSpanName       bool     `yaml:"force_span_name"`
}
