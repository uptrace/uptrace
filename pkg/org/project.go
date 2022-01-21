package org

import (
	"context"
	"database/sql"

	"github.com/uptrace/uptrace/pkg/bunapp"
)

func SelectProjectByID(
	ctx context.Context, app *bunapp.App, projectID uint32,
) (*bunapp.Project, error) {
	projects := app.Config().Projects
	for i := range projects {
		project := &projects[i]
		if project.ID == projectID {
			return project, nil
		}
	}
	return nil, sql.ErrNoRows
}
