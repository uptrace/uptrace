package org

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"go.uber.org/zap"
)

func SelectProjectByID(
	ctx context.Context, app *bunapp.App, projectID uint32,
) (*bunconf.Project, error) {
	projects := app.Config().Projects
	for i := range projects {
		project := &projects[i]
		if project.ID == projectID {
			return project, nil
		}
	}
	return nil, sql.ErrNoRows
}

func SelectProjectByDSN(
	ctx context.Context, app *bunapp.App, dsnStr string,
) (*bunconf.Project, error) {
	dsn, err := ParseDSN(dsnStr)
	if err != nil {
		return nil, err
	}

	if dsn.Token == "" {
		return nil, fmt.Errorf("dsn %q does not have a token", dsnStr)
	}

	projects := app.Config().Projects

	for i := range projects {
		project := &projects[i]
		if project.Token == dsn.Token {
			if project.ID != dsn.ProjectID {
				app.Zap(ctx).Error("project token and project id don't match",
					zap.String("dsn", dsnStr))
			}
			return project, nil
		}
	}

	return nil, fmt.Errorf("project with token %q not found", dsn.Token)
}
