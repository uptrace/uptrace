package command

import (
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"github.com/uptrace/uptrace/pkg/bunapp"
)

func runSubcommand(c *cli.Context, f any, opts ...fx.Option) error {
	opts = append(opts, fx.Supply(c), fx.NopLogger, fx.Invoke(f))
	app, err := bunapp.New(
		c.String("config"),
		opts...,
	)
	if err != nil {
		return err
	}

	_ = app.Start(c.Context)
	return app.Stop(c.Context)
}
