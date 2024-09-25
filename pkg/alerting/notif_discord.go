package alerting

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
)

type DiscordNotifChannel struct {
	*BaseNotifChannel `bun:",inherit"`
	Params            DiscordParams `json:"params"`
}

type DiscordParams struct {
	WebhookId    int64  `json:"webhook_id"`
	WebhookToken string `json:"webhook_token"`
}

func newDiscordNotifChannel(src *BaseNotifChannel) (*DiscordNotifChannel, error) {
	channel := &DiscordNotifChannel{
		BaseNotifChannel: src,
	}

	dec := json.NewDecoder(bytes.NewReader(src.ParamsData))
	dec.UseNumber()
	if err := dec.Decode(&channel.Params); err != nil {
		return nil, err
	}

	return channel, nil
}

func (d *DiscordNotifChannel) Base() *BaseNotifChannel {
	return d.BaseNotifChannel
}

func SelectDiscordNotifChannel(
	ctx context.Context, app *bunapp.App, channelID uint64,
) (*DiscordNotifChannel, error) {
	channelAny, err := SelectNotifChannel(ctx, app, channelID)
	if err != nil {
		return nil, err
	}

	channel, ok := channelAny.(*DiscordNotifChannel)
	if !ok {
		return nil, fmt.Errorf("unexpected notif channel: %T", channelAny)
	}
	return channel, nil
}

//------------------------------------------------------------------------------

func notifyByDiscordHandler(ctx context.Context, eventID, channelID uint64) error {
	app := bunapp.AppFromContext(ctx)

	alert, err := selectAlertWithEvent(ctx, app, eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	baseAlert := alert.Base()

	project, err := org.SelectProject(ctx, app, baseAlert.ProjectID)
	if err != nil {
		return err
	}

	channel, err := SelectDiscordNotifChannel(ctx, app, channelID)
	if err != nil {
		return err
	}

	return notifyByDiscordChannel(ctx, app, project, alert, channel)
}

func notifyByDiscordChannel(
	_ context.Context,
	app *bunapp.App,
	_ *org.Project,
	_ org.Alert,
	channel *DiscordNotifChannel,
) error {
	fmt.Println("<<<------888------->>>")

	// dev data
	client := webhook.New(1287925691186417674, "yW5CszHQYP3l_x9qJlX5DLEiKToFE2je7SGQ3SgqA99p7hyAf1ShbY7dh4m1X9QxtjDy")

	_, err := client.CreateMessage(discord.WebhookMessageCreate{
		Content: "hello world!",
	})
	
	return err
}
