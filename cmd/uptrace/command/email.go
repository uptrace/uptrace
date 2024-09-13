package command

import (
	"fmt"

	"github.com/uptrace/uptrace/pkg/uptracebundle"
	"github.com/urfave/cli/v2"
	"github.com/wneessen/go-mail"
)

// UPTRACE_CONFIG=config/uptrace.yml go run cmd/uptrace/main.go email-test --to uptrace@localhost

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
			_, app, err := uptracebundle.StartCLI(c)
			if err != nil {
				return fmt.Errorf("failed to start app: %w", err)
			}
			defer app.Stop()

			client, err := app.InitMailer()
			if err != nil {
				return fmt.Errorf("failed to initialize mailer: %w", err)
			}

			msg := mail.NewMsg()
			msg.Subject("[Uptrace] Test email")
			msg.SetBodyString(mail.TypeTextPlain, "This is a test email")

			msg.From(app.Config().SMTPMailer.From)
			msg.AddTo(c.String("to"))

			err = client.DialAndSend(msg)
			if err != nil {
				return fmt.Errorf("failed to send email: %w", err)
			}

			fmt.Println("Test email sent successfully to", c.String("to"))
			return nil
		},
	}
}
