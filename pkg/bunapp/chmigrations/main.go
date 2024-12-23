package chmigrations

import (
	"embed"

	"github.com/uptrace/pkg/clickhouse/chmigrate"
)

var Migrations = chmigrate.NewMigrations()

//go:embed *.sql
var sqlMigrations embed.FS

func init() {
	if err := Migrations.Discover(sqlMigrations); err != nil {
		panic(err)
	}
}
