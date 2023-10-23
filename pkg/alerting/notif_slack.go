package alerting

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/slack-go/slack"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"go.uber.org/zap"
)

func notifyBySlackHandler(ctx context.Context, eventID, channelID uint64) error {
	app := bunapp.AppFromContext(ctx)

	alert, err := selectAlertWithEvent(ctx, app, eventID)
	if err != nil {
		return err
	}
	baseAlert := alert.Base()

	project, err := org.SelectProject(ctx, app, baseAlert.ProjectID)
	if err != nil {
		return err
	}

	channel, err := SelectSlackNotifChannel(ctx, app, channelID)
	if err != nil {
		return err
	}

	return notifyBySlackChannel(ctx, app, project, alert, channel)
}

func notifyBySlackChannel(
	ctx context.Context,
	app *bunapp.App,
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
			ctx, app, channel.Base(), NotifChannelDisabled,
		); err != nil {
			app.Zap(ctx).Error("UpdateNotifChannelState failed", zap.Error(err))
			return err
		}
		return nil
	}

	block, err := slackAlertBlock(app, project, alert)
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
				ctx, app, channel.Base(), NotifChannelDisabled,
			); err != nil {
				app.Zap(ctx).Error("UpdateNotifChannelState failed", zap.Error(err))
				return err
			}
			return nil
		}
		return err
	default:
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			if err := UpdateNotifChannelState(
				ctx, app, channel.Base(), NotifChannelDisabled,
			); err != nil {
				app.Zap(ctx).Error("UpdateNotifChannelState failed", zap.Error(err))
				return err
			}
			return nil
		}

		app.Zap(ctx).Error("slack.PostWebhook failed",
			zap.String("webhook", webhookURL),
			zap.Error(err),
			zap.String("unwrap", fmt.Sprintf("%T", errors.Unwrap(err))))
		return err
	}
}

func slackAlertBlock(
	app *bunapp.App, project *org.Project, alert org.Alert,
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
	viewBtn.URL = app.SiteURL(baseAlert.URL())

	return slack.NewSectionBlock(text, nil, slack.NewAccessory(viewBtn)), nil
}
