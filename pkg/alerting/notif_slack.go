package alerting

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/slack-go/slack"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/utf8util"
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
	text := slack.NewTextBlockObject("mrkdwn", "", false, false)

	switch alert := alert.(type) {
	case *ErrorAlert:
		text.Text = slackErrorAlertText(project, alert)
	case *MetricAlert:
		text.Text = slackMetricAlertText(project, alert)
	default:
		return nil, fmt.Errorf("unknown alert type: %T", alert)
	}

	baseAlert := alert.Base()
	viewBtnText := slack.NewTextBlockObject("plain_text", "View", false, false)
	viewBtn := slack.NewButtonBlockElement(
		"button-action",
		fmt.Sprintf("view_%d", baseAlert.ID),
		viewBtnText,
	)
	viewBtn.URL = app.SiteURL(baseAlert.URL())

	return slack.NewSectionBlock(text, nil, slack.NewAccessory(viewBtn)), nil
}

func slackErrorAlertText(project *org.Project, alert *ErrorAlert) string {
	alertName := utf8util.Trunc(alert.Name, 255)

	if alert.Event.User != nil {
		if alert.State == org.AlertOpen {
			return fmt.Sprintf(
				":red_circle: *[%s] The error is reopened by %s*\n%s",
				project.Name, alert.Event.User.Username, alertName,
			)
		}
		return fmt.Sprintf(
			":white_check_mark: *[%s] The error is closed by %s*\n%s",
			project.Name, alert.Event.User.Username, alertName,
		)
	}

	if alert.State == org.AlertClosed {
		return fmt.Sprintf(
			":white_check_mark: *[%s] The error is closed*\n%s",
			project.Name, alertName,
		)
	}

	return fmt.Sprintf(
		":beetle: *[%s] %s*\n%s",
		project.Name, alert.Summary(), alertName,
	)
}

func slackMetricAlertText(project *org.Project, alert *MetricAlert) string {
	alertName := utf8util.Trunc(alert.Name, 255)

	if alert.Event.User != nil {
		if alert.State == org.AlertOpen {
			return fmt.Sprintf(
				":red_circle: *[%s] The metric alert is reopened by %s*\n%s",
				project.Name, alert.Event.User.Username, alertName,
			)
		}
		return fmt.Sprintf(
			":white_check_mark: *[%s] The metric alert is closed by %s*\n%s",
			project.Name, alert.Event.User.Username, alertName,
		)
	}

	if alert.State == org.AlertClosed {
		return fmt.Sprintf(
			":white_check_mark: *[%s] The metric is back to normal*\n%s",
			project.Name, alertName,
		)
	}

	var title string
	switch alert.Event.Name {
	case org.AlertEventCreated:
		title = "The metric is firing"
	default:
		title = "The metric is still firing"
	}

	return fmt.Sprintf(
		"%s *[%s] %s*: %s\n%s",
		slackAnomalyEmoji(&alert.Params),
		project.Name,
		title,
		alert.Summary(),
		alertName,
	)
}

func slackAnomalyEmoji(params *MetricAlertParams) string {
	switch params.Firing {
	case -1:
		return ":arrow_lower_right:"
	case 1:
		return ":arrow_upper_right:"
	default:
		return ""
	}
}
