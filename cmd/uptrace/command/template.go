package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/uptrace/uptrace/pkg/bunconf"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

func NewTemplateCommand() *cli.Command {
	return &cli.Command{
		Name:  "tpl",
		Usage: "Uptrace dashboard templates commands",
		Subcommands: []*cli.Command{
			{
				Name:  "validate",
				Usage: "validate dashboard templates",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "dir",
						Value: "config/dashboard-templates",
						Usage: "path to the dir containing dashboard templates",
					},
				},
				Action: func(c *cli.Context) error {
					dir, err := filepath.Abs(c.String("dir"))
					if err != nil {
						return err
					}

					entries, err := os.ReadDir(dir)
					if err != nil {
						return err
					}

					for _, e := range entries {
						if e.IsDir() {
							continue
						}

						b, err := os.ReadFile(filepath.Join(dir, e.Name()))
						if err != nil {
							return err
						}

						if err := validateYAML(b); err != nil {
							return fmt.Errorf("%s: %w", e.Name(), err)
						}
					}

					return nil
				},
			},
		},
	}
}

func validateYAML(b []byte) error {
	dashboard := new(bunconf.Dashboard)

	dec := yaml.NewDecoder(bytes.NewReader(b))
	for {
		if err := dec.Decode(&dashboard); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err := dashboard.Validate(); err != nil {
			return err
		}
	}

	return nil
}
