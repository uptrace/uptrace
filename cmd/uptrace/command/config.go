package command

import (
	"fmt"
	"os"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
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
					_, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					conf := app.Config()
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
				},
			},
		},
	}
}
