package org

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"go.uber.org/fx"
)

type Project struct {
	bun.BaseModel `bun:"projects,alias:p"`

	ID    uint32 `json:"id" bun:",pk,autoincrement"`
	Name  string `json:"name" bun:",nullzero"`
	Token string `json:"token" bun:",nullzero"`

	PinnedAttrs         []string `json:"pinnedAttrs" bun:",array"`
	GroupByEnv          bool     `json:"groupByEnv"`
	GroupFuncsByService bool     `json:"groupFuncsByService"`
	PromCompat          bool     `json:"promCompat"`
	ForceSpanName       []string `json:"forceSpanName" bun:",array"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewProjectFromConfig(src *bunconf.Project) (*Project, error) {
	p := &Project{
		ID:                  src.ID,
		Name:                src.Name,
		Token:               src.Token,
		PinnedAttrs:         src.PinnedAttrs,
		GroupByEnv:          src.GroupByEnv,
		GroupFuncsByService: src.GroupFuncsByService,
		PromCompat:          src.PromCompat,
		ForceSpanName:       src.ForceSpanName,
		CreatedAt:           time.Now(),
	}
	p.UpdatedAt = p.CreatedAt

	return p, nil
}

func (p *Project) Init() error {
	if p.ID == 0 {
		return errors.New("project id can't be zero")
	}
	if p.Name == "" {
		return errors.New("project name can't be empty")
	}
	if p.Token == "" {
		return errors.New("project token can't be empty")
	}
	return nil
}

func (p *Project) SettingsURL() string {
	return fmt.Sprintf("/projects/%d", p.ID)
}

func (p *Project) EmailSettingsURL() string {
	return fmt.Sprintf("/alerting/%d/email", p.ID)
}

func (p *Project) DSN(conf *bunconf.Config) string {
	return BuildDSN(conf, p.Token)
}

type ProjectGatewayParams struct {
	fx.In

	Conf *bunconf.Config
}
type ProjectGateway struct {
	*ProjectGatewayParams

	projects []*Project
}

func NewProjectGateway(p ProjectGatewayParams) (*ProjectGateway, error) {
	var projects []*Project

	for _, src := range p.Conf.Projects {
		p, err := NewProjectFromConfig(&src)
		if err != nil {
			return nil, err
		}

		if err := p.Init(); err != nil {
			return nil, err
		}

		projects = append(projects, p)
	}

	return &ProjectGateway{
		ProjectGatewayParams: &p,
		projects:             projects,
	}, nil
}

func findProject(projects []*Project, f func(*Project) bool) (*Project, error) {
	idx := slices.IndexFunc(projects, f)
	if idx == -1 {
		return nil, sql.ErrNoRows
	}
	return projects[idx], nil
}

func (g *ProjectGateway) SelectByID(ctx context.Context, projectID uint32) (*Project, error) {
	return findProject(g.projects, func(p *Project) bool {
		return p.ID == projectID
	})
}

func (g *ProjectGateway) SelectByDSN(ctx context.Context, dsnStr string) (*Project, error) {
	dsn, err := ParseDSN(dsnStr)
	if err != nil {
		return nil, err
	}

	if dsn.Token == "" {
		return nil, fmt.Errorf("dsn %q does not have a token", dsnStr)
	}

	project, err := g.SelectByToken(ctx, dsn.Token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("can't find project with token=%q", dsn.Token)
		}
		return nil, err
	}

	return project, nil
}

func (g *ProjectGateway) SelectByToken(ctx context.Context, token string) (*Project, error) {
	return findProject(g.projects, func(p *Project) bool {
		return p.Token == token
	})
}

func (g *ProjectGateway) SelectAll(ctx context.Context) ([]*Project, error) {
	return g.projects, nil
}
