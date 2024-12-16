package alerting

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/slack-go/slack"
	"go.uber.org/zap"

	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/org"
)

func (h *NotifChannelHandler) notifyBySlackHandler(ctx context.Context, eventID, channelID uint64) error {
	alert, err := selectAlertWithEvent(ctx, h.PG, eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	baseAlert := alert.Base()

	project, err := h.PS.SelectProject(ctx, baseAlert.ProjectID)
	if err != nil {
		return err
	}

	channel, err := SelectSlackNotifChannel(ctx, h.PG, channelID)
	if err != nil {
		return err
	}

	return h.notifyBySlackChannel(ctx, project, alert, channel)
}

func (h *NotifChannelHandler) notifyBySlackChannel(
	ctx context.Context,
	project *org.Project,
	alert org.Alert,
	channel *SlackNotifChannel,
) error {
	if channel.State != NotifChannelDelivering {
		return nil
	}

	webhookURL := channel.Params.WebhookURL

	if webhookURL == "" {
		if err := UpdateNotifChannelState(
			ctx, h.PG, channel.Base(), NotifChannelDisabled,
		); err != nil {
			h.Logger.Error("UpdateNotifChannelState failed", zap.Error(err))
			return err
		}
		return nil
	}

	block, err := slackAlertBlock(h.Conf, project, alert)
	if err != nil {
		return err
	}

	baseAlert := alert.Base()
	msg := &slack.WebhookMessage{
		Text: fmt.Sprintf("[%s] %s", project.Name, baseAlert.Name),
		Blocks: &slack.Blocks{
			BlockSet: []slack.Block{block},
		},
	}

	switch err := slack.PostWebhook(webhookURL, msg); err := err.(type) {
	case nil:
		return nil
	case slack.StatusCodeError:
		if err.Code == 404 {
			if err := UpdateNotifChannelState(
				ctx, h.PG, channel.Base(), NotifChannelDisabled,
			); err != nil {
				h.Logger.Error("UpdateNotifChannelState failed", zap.Error(err))
				return err
			}
			return nil
		}
		return err
	default:
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			if err := UpdateNotifChannelState(
				ctx, h.PG, channel.Base(), NotifChannelDisabled,
			); err != nil {
				h.Logger.Error("UpdateNotifChannelState failed", zap.Error(err))
				return err
			}
			return nil
		}

		h.Logger.Error("slack.PostWebhook failed",
			zap.String("webhook", webhookURL),
			zap.Error(err),
			zap.String("unwrap", fmt.Sprintf("%T", errors.Unwrap(err))))
		return err
	}
}

func slackAlertBlock(
	conf *bunconf.Config, project *org.Project, alert org.Alert,
) (*slack.SectionBlock, error) {
	baseAlert := alert.Base()
	text := slack.NewTextBlockObject("mrkdwn", "", false, false)

	switch alert := alert.(type) {
	case *ErrorAlert:
		text.Text = telegramErrorFormatter.Format(project, alert)
	case *MetricAlert:
		text.Text = telegramMetricFormatter.Format(project, alert)
	default:
		return nil, fmt.Errorf("unknown alert type: %T", alert)
	}

	viewBtnText := slack.NewTextBlockObject("plain_text", "Open", false, false)
	viewBtn := slack.NewButtonBlockElement(
		"button-action",
		fmt.Sprintf("view_%d", baseAlert.ID),
		viewBtnText,
	)
	viewBtn.URL = conf.SiteURL(baseAlert.URL())

	return slack.NewSectionBlock(text, nil, slack.NewAccessory(viewBtn)), nil
}
