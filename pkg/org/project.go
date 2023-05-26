package org

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

type Project struct {
	bun.BaseModel `bun:"projects,alias:p"`

	ID                  uint32   `json:"id" bun:",pk,autoincrement"`
	Name                string   `json:"name" bun:",nullzero"`
	Token               string   `json:"token" bun:",nullzero"`
	PinnedAttrs         []string `json:"pinnedAttrs" bun:",array"`
	GroupByEnv          bool     `json:"groupByEnv"`
	GroupFuncsByService bool     `json:"groupFuncsByService"`

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

func SelectProject(
	ctx context.Context, app *bunapp.App, projectID uint32,
) (*Project, error) {
	project := new(Project)
	if err := app.PG.NewSelect().
		Model(project).
		Where("id = ?", projectID).
		Limit(1).
		Scan(ctx); err != nil {
		return nil, err
	}
	return project, nil
}

func SelectProjectByDSN(
	ctx context.Context, app *bunapp.App, dsnStr string,
) (*Project, error) {
	dsn, err := ParseDSN(dsnStr)
	if err != nil {
		return nil, err
	}

	if dsn.Token == "" {
		return nil, fmt.Errorf("dsn %q does not have a token", dsnStr)
	}

	project, err := SelectProjectByToken(ctx, app, dsn.Token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("can't find project with token=%q", dsn.Token)
		}
		return nil, err
	}

	return project, nil
}

func SelectProjectByToken(
	ctx context.Context, app *bunapp.App, token string,
) (*Project, error) {
	project := new(Project)
	if err := app.PG.NewSelect().
		Model(project).
		Where("token = ?", token).
		Limit(1).
		Scan(ctx); err != nil {
		return nil, err
	}
	return project, nil
}

func SelectProjects(ctx context.Context, app *bunapp.App) ([]*Project, error) {
	projects := make([]*Project, 9)
	if err := app.PG.NewSelect().
		Model(&projects).
		Scan(ctx); err != nil {
		return nil, err
	}
	return projects, nil
}
