package org

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunconf"
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

func SelectProject(ctx context.Context, pg *bun.DB, projectID uint32) (*Project, error) {
	project := new(Project)
	if err := pg.NewSelect().
		Model(project).
		Where("id = ?", projectID).
		Limit(1).
		Scan(ctx); err != nil {
		return nil, err
	}
	return project, nil
}

func SelectProjectByDSN(ctx context.Context, pg *bun.DB, dsnStr string) (*Project, error) {
	dsn, err := ParseDSN(dsnStr)
	if err != nil {
		return nil, err
	}

	if dsn.Token == "" {
		return nil, fmt.Errorf("dsn %q does not have a token", dsnStr)
	}

	project, err := SelectProjectByToken(ctx, pg, dsn.Token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("can't find project with token=%q", dsn.Token)
		}
		return nil, err
	}

	return project, nil
}

func SelectProjectByToken(ctx context.Context, pg *bun.DB, token string) (*Project, error) {
	project := new(Project)
	if err := pg.NewSelect().
		Model(project).
		Where("token = ?", token).
		Limit(1).
		Scan(ctx); err != nil {
		return nil, err
	}
	return project, nil
}

func SelectProjects(ctx context.Context, pg *bun.DB) ([]*Project, error) {
	projects := make([]*Project, 0)
	if err := pg.NewSelect().
		Model(&projects).
		Scan(ctx); err != nil {
		return nil, err
	}
	return projects, nil
}
