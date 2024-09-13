package alerting

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
)

func notifyByDiscordHandler(ctx context.Context, eventID, channelID uint64) error {
	app := bunapp.AppFromContext(ctx)

	alert, err := selectAlertWithEvent(ctx, app, eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	baseAlert := alert.Base()

	project, err := org.SelectProject(ctx, app, baseAlert.ProjectID)
	if err != nil {
		return err
	}

	fmt.Println("-------------------", project.Name, eventID, channelID)
	return nil
}
