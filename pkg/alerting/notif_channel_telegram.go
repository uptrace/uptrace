package alerting

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/segmentio/encoding/json"

	"github.com/uptrace/bun"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/org"
)

type TelegramNotifChannel struct {
	*BaseNotifChannel `bun:",inherit"`

	Params TelegramParams `json:"params"`
}

type TelegramParams struct {
	ChatID int64 `json:"chatId"`
}

func newTelegramNotifChannel(src *BaseNotifChannel) (*TelegramNotifChannel, error) {
	channel := &TelegramNotifChannel{
		BaseNotifChannel: src,
	}

	dec := json.NewDecoder(bytes.NewReader(src.ParamsData))
	dec.UseNumber()
	if err := dec.Decode(&channel.Params); err != nil {
		return nil, err
	}

	return channel, nil
}

var _ NotifChannel = (*TelegramNotifChannel)(nil)

func (c *TelegramNotifChannel) Base() *BaseNotifChannel {
	return c.BaseNotifChannel
}

func (c *TelegramNotifChannel) TelegramBot(conf *bunconf.Config) (*tgbotapi.BotAPI, error) {
	if conf.Telegram.BotToken == "" {
		return nil, errors.New("telegram.bot_token is empty")
	}
	if c.Params.ChatID == 0 {
		return nil, errors.New("chat id can't be empty")
	}

	bot, err := tgbotapi.NewBotAPI(conf.Telegram.BotToken)
	if err != nil {
		return nil, err
	}

	return bot, nil
}

func SelectTelegramNotifChannel(
	ctx context.Context, pg *bun.DB, channelID uint64,
) (*TelegramNotifChannel, error) {
	channelAny, err := SelectNotifChannel(ctx, pg, channelID)
	if err != nil {
		return nil, err
	}

	channel, ok := channelAny.(*TelegramNotifChannel)
	if !ok {
		return nil, fmt.Errorf("unexpected notif channel: %T", channelAny)
	}
	return channel, nil
}

//------------------------------------------------------------------------------

func (h *NotifChannelHandler) notifyByTelegramHandler(ctx context.Context, eventID, channelID uint64) error {
	alert, err := selectAlertWithEvent(ctx, h.PG, h.Users, eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	baseAlert := alert.Base()

	project, err := h.Projects.SelectByID(ctx, baseAlert.ProjectID)
	if err != nil {
		return err
	}

	channel, err := SelectTelegramNotifChannel(ctx, h.PG, channelID)
	if err != nil {
		return err
	}

	return notifyByTelegramChannel(ctx, h.Conf, project, alert, channel)
}

func notifyByTelegramChannel(
	ctx context.Context,
	conf *bunconf.Config,
	project *org.Project,
	alert org.Alert,
	channel *TelegramNotifChannel,
) error {
	if channel.State != NotifChannelDelivering {
		return nil
	}

	msg, err := telegramMsg(conf, project, channel.Params.ChatID, alert)
	if err != nil {
		return err
	}

	bot, err := channel.TelegramBot(conf)
	if err != nil {
		return err
	}

	if _, err := bot.Send(msg); err != nil {
		return err
	}

	return nil
}

func telegramMsg(
	conf *bunconf.Config, project *org.Project, chatID int64, alert org.Alert,
) (*tgbotapi.MessageConfig, error) {
	baseAlert := alert.Base()
	msg := tgbotapi.NewMessage(chatID, "")

	switch alert := alert.(type) {
	case *ErrorAlert:
		msg.Text = telegramErrorFormatter.Format(project, alert)
	case *MetricAlert:
		msg.Text = telegramMetricFormatter.Format(project, alert)
	default:
		return nil, fmt.Errorf("unsupported alert type: %T", alert)
	}

	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(
				"View on Uptrace",
				conf.SiteURL(baseAlert.URL()),
			),
		),
	)

	return &msg, nil
}
