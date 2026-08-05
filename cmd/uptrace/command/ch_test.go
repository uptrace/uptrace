package command

import (
	"testing"

	"github.com/uptrace/pkg/clickhouse/ch"
	"github.com/uptrace/pkg/clickhouse/chmigrate"
	"github.com/uptrace/uptrace/pkg/bunconf"
)

func TestNewCHMigratorLogsSchemaArgs(t *testing.T) {
	t.Run("uses logs settings", func(t *testing.T) {
		conf := new(bunconf.Config)
		conf.CHSchema.Spans.TTLDelete = "30 DAY"
		conf.CHSchema.Spans.StoragePolicy = "spans"
		conf.CHSchema.Logs.TTLDelete = "3 DAY"
		conf.CHSchema.Logs.StoragePolicy = "logs"

		got := formatLogsSchemaArgs(t, conf)
		if want := "3 DAY|'logs'"; got != want {
			t.Fatalf("formatted logs schema args = %q, want %q", got, want)
		}
	})

}

func formatLogsSchemaArgs(t *testing.T, conf *bunconf.Config) string {
	t.Helper()

	chdb := ch.Connect()
	t.Cleanup(func() {
		if err := chdb.Close(); err != nil {
			t.Errorf("closing ClickHouse client: %v", err)
		}
	})

	migrator := NewCHMigrator(conf, chdb, chmigrate.NewMigrations())
	return migrator.DB().Formatter().FormatQuery("?LOGS_TTL|?LOGS_STORAGE")
}
