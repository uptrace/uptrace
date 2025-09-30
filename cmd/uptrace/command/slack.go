package command

import (
	"fmt"

	"github.com/slack-go/slack"
	"github.com/urfave/cli/v2"
)

func NewSlackCommand() *cli.Command {
	return &cli.Command{
		Name:  "slack-test",
		Usage: "send test Slack message",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "method",
				Usage:    "authentication method: webhook or token",
				Value:    "webhook",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "webhook-url",
				Usage:    "Slack webhook URL (for webhook method)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "token",
				Usage:    "Slack bot token (for token method)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "channel",
				Usage:    "Slack channel ID, name (#channel), or user (@user) (for token method)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "message",
				Usage:    "test message to send",
				Value:    "Test message from Uptrace",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			method := c.String("method")
			message := c.String("message")

			switch method {
			case "webhook":
				return sendSlackWebhookTest(c.String("webhook-url"), message)
			case "token":
				return sendSlackTokenTest(c.String("token"), c.String("channel"), message)
			default:
				return fmt.Errorf("unsupported method: %s. Use 'webhook' or 'token'", method)
			}
		},
	}
}

func sendSlackWebhookTest(webhookURL, message string) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL is required for webhook method")
	}

	msg := &slack.WebhookMessage{
		Text: message,
	}

	if err := slack.PostWebhook(webhookURL, msg); err != nil {
		return fmt.Errorf("failed to send webhook message: %w", err)
	}

	fmt.Println("Webhook message sent successfully!")
	return nil
}

func sendSlackTokenTest(token, channel, message string) error {
	if token == "" {
		return fmt.Errorf("token is required for token method")
	}
	if channel == "" {
		return fmt.Errorf("channel is required for token method")
	}

	client := slack.New(token)

	_, _, err := client.PostMessage(
		channel,
		slack.MsgOptionText(message, false),
	)

	if err != nil {
		return fmt.Errorf("failed to send token-based message: %w", err)
	}

	fmt.Printf("Token-based message sent successfully to channel: %s\n", channel)
	return nil
}