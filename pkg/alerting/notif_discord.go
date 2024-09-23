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
	"log"
)

type userStorage map[string]string

var storage = userStorage{}

func (u userStorage) Set(username, chatId string) {
	u[username] = chatId
}

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
	_ context.Context,
	app *bunapp.App,
	_ *org.Project,
	_ org.Alert,
	channel *DiscordNotifChannel,
) error {
	if channel.State != NotifChannelDelivering {
		return nil
	}

	dg, err := channel.DiscordSession(app)
	if err != nil {
		return err
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		return fmt.Errorf("error opening Discord session: %v", err)
	}

	return err
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if _, ok := storage[m.Author.Username]; !ok {
		storage.Set(m.Author.Username, m.ChannelID)
	}

	var cfg = &SMC{
		session:       s,
		messageCreate: m,
	}

	// DM logic
	if m.GuildID == "" {
		send(cfg, "test1")
		return
	}

	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		_, _ = s.ChannelMessageSend(
			m.ChannelID,
			"Something went wrong while sending the DM!",
		)
		return
	}
	cfg.channel = channel

	send(cfg, "test2")
}

type SMC struct {
	session       *discordgo.Session
	messageCreate *discordgo.MessageCreate
	channel       *discordgo.Channel
}

func send(cfg *SMC, content string) {
	_, err := cfg.session.ChannelMessageSend(cfg.messageCreate.ChannelID, content)
	if err != nil {
		log.Printf("Error - sending message: %v", err)
		_, _ = cfg.session.ChannelMessageSend(
			cfg.messageCreate.ChannelID,
			"Failed to send you a DM. "+
				"Did you disable DM in your privacy settings?",
		)
	}
}
