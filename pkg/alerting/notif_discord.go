package alerting

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
)

type DiscordNotifier struct {
	session *discordgo.Session
}

type DiscordNotifChannel struct {
	*BaseNotifChannel `bun:",inherit"`

	Params DiscordParams `json:"params"`
}

type DiscordParams struct {
	ChatID string `json:"chatId"`
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

func (d *DiscordNotifChannel) DiscordSession(app *bunapp.App) (*discordgo.Session, error) {
	conf := app.Config()
	if conf.Discord.BotToken == "" {
		return nil, errors.New("discord.bot_token is empty")
	}
	if conf.Discord.PublicKey == "" {
		return nil, errors.New("discord.public_key is empty")
	}
	if conf.Discord.AppId == "" {
		return nil, errors.New("discord.app_id is empty")
	}
	if d.Params.ChatID == "" {
		return nil, errors.New("chat id can't be empty")
	}

	dg, err := discordgo.New("Bot " + app.Config().Discord.BotToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %v", err)
	}

	return dg, nil
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
	ctx context.Context,
	app *bunapp.App,
	project *org.Project,
	alert org.Alert,
	channel *DiscordNotifChannel,
) error {
	if channel.State != NotifChannelDelivering {
		return nil
	}

	bot, err := channel.DiscordSession(app)
	if err != nil {
		return err
	}

	_, err = bot.ChannelMessageSend(channel.Params.ChatID, fmt.Sprintf("%v", alert))
	return err
}
