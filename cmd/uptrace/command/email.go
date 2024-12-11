package command

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/wneessen/go-mail"
	"go.uber.org/fx"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconf"
)

func NewEmailCommand() *cli.Command {
	return &cli.Command{
		Name:  "email-test",
		Usage: "send test email",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "to",
				Usage:    "recipient email address",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			return runSubcommand(c, chEmailTest)
		},
	}
}

func chEmailTest(lc fx.Lifecycle, c *cli.Context, conf *bunconf.Config) {
	lc.Append(fx.StartHook(func(_ context.Context) error {
		client, err := bunapp.NewMailer(conf)
		if err != nil {
			return fmt.Errorf("failed to initialize mailer: %w", err)
		}

		msg := mail.NewMsg()
		msg.Subject("[Uptrace] Test email")
		msg.SetBodyString(mail.TypeTextPlain, "This is a test email")

		_ = msg.From(conf.SMTPMailer.From)
		_ = msg.AddTo(c.String("to"))

		err = client.DialAndSend(msg)
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}

		fmt.Println("Test email sent successfully to", c.String("to"))
		return nil
	}))
}
