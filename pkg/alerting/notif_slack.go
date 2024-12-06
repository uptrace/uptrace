package alerting

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/slack-go/slack"
	"go.uber.org/zap"

	"github.com/uptrace/bun"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/uptrace/uptrace/pkg/org"
)

func (h *NotifChannelHandler) notifyBySlackHandler(ctx context.Context, eventID, channelID uint64) error {
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

	channel, err := SelectSlackNotifChannel(ctx, h.PG, channelID)
	if err != nil {
		return err
	}

	return notifyBySlackChannel(ctx, h.Logger, h.Conf, h.PG, project, alert, channel)
}

func notifyBySlackChannel(
	ctx context.Context,
	logger *otelzap.Logger,
	conf *bunconf.Config,
	pg *bun.DB,
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
			ctx, pg, channel.Base(), NotifChannelDisabled,
		); err != nil {
			logger.Error("UpdateNotifChannelState failed", zap.Error(err))
			return err
		}
		return nil
	}

	block, err := slackAlertBlock(conf, project, alert)
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
				ctx, pg, channel.Base(), NotifChannelDisabled,
			); err != nil {
				logger.Error("UpdateNotifChannelState failed", zap.Error(err))
				return err
			}
			return nil
		}
		return err
	default:
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			if err := UpdateNotifChannelState(
				ctx, pg, channel.Base(), NotifChannelDisabled,
			); err != nil {
				logger.Error("UpdateNotifChannelState failed", zap.Error(err))
				return err
			}
			return nil
		}

		logger.Error("slack.PostWebhook failed",
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
