package command

import (
	"fmt"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/urfave/cli/v2"
)

func NewCHSchemaCommand() *cli.Command {
	return &cli.Command{
		Name:  "ch_schema",
		Usage: "Commands to update ClickHouse schema",
		Subcommands: []*cli.Command{
			{
				Name:  "ttl_move",
				Usage: "update ClickHouse schema to move data to S3",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "table",
						Value: "",
						Usage: "table name, for example, spans_data or spans_index",
					},
					&cli.StringFlag{
						Name:  "after",
						Value: "10 DAY",
						Usage: "ttl interval value",
					},
					&cli.StringFlag{
						Name:  "volume",
						Value: "s3",
						Usage: "volume name",
					},
					&cli.StringFlag{
						Name:  "storage",
						Value: "",
						Usage: "table storage policy",
					},
				},
				Action: func(c *cli.Context) error {
					ctx, app, err := bunapp.StartCLI(c)
					if err != nil {
						return err
					}
					defer app.Stop()

					chSchema := &app.Config().CHSchema

					table := c.String("table")
					var ttlDelete string
					ttlMove := c.String("after")
					volume := c.String("volume")
					storage := c.String("storage")

					switch table {
					case "":
						return fmt.Errorf("--table flag is required")
					case "spans_data", "spans_index":
						ttlDelete = chSchema.Spans.TTLDelete
						if storage == "" {
							storage = chSchema.Spans.StoragePolicy
						}
					case "datapoint_minutes", "datapoint_hours":
						ttlDelete = chSchema.Metrics.TTLDelete
						if storage == "" {
							storage = chSchema.Metrics.StoragePolicy
						}
					default:
						return fmt.Errorf("unsupported table name: %q", table)
					}

					params := &struct {
						Table     ch.Ident
						TTLDelete ch.Safe
						TTLMove   ch.Safe
						Volume    string
						Storage   string
					}{
						Table:     ch.Ident(table),
						TTLDelete: ch.Safe(ttlDelete),
						TTLMove:   ch.Safe(ttlMove),
						Volume:    volume,
						Storage:   storage,
					}

					queries := []string{
						`ALTER TABLE ?table MODIFY SETTING storage_policy = ?storage`,
						`ALTER TABLE ?table MODIFY TTL toDate(time) + INTERVAL ?ttl_delete DELETE, ` +
							`toDate(time) + INTERVAL ?ttl_move TO VOLUME ?volume`,
					}
					for _, query := range queries {
						if _, err := app.CH.ExecContext(ctx, query, params); err != nil {
							return err
						}
					}

					return nil
				},
			},
		},
	}
}
