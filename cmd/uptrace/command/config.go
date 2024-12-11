package command

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"

	"github.com/uptrace/uptrace/pkg/bunconf"
)

func NewConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Uptrace config commands",
		Subcommands: []*cli.Command{
			{
				Name:  "dump",
				Usage: "dumps Uptrace config in YAML format",
				Action: func(c *cli.Context) error {
					return runSubcommand(c, configDump)
				},
			},
		},
	}
}

func configDump(lc fx.Lifecycle, conf *bunconf.Config) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		fmt.Fprintf(os.Stdout, "# %s\n\n", conf.Path)

		enc := yaml.NewEncoder(os.Stdout)
		enc.SetIndent(2)

		if err := enc.Encode(conf); err != nil {
			return err
		}
		if err := enc.Close(); err != nil {
			return err
		}

		return nil
	}))
}
