package chmigrate_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/go-clickhouse/chdebug"
	"github.com/uptrace/go-clickhouse/chmigrate"
)

func TestMigrate(t *testing.T) {
	type Test struct {
		run func(t *testing.T, db *ch.DB)
	}

	tests := []Test{
		{run: testChmigrateUpAndDown},
		{run: testChmigrateUpError},
	}

	db := chDB()

	for _, test := range tests {
		t.Run(funcName(test.run), func(t *testing.T) {
			test.run(t, db)
		})
	}
}

func testChmigrateUpAndDown(t *testing.T, db *ch.DB) {
	ctx := context.Background()

	var history []string

	migrations := chmigrate.NewMigrations()
	migrations.Add(chmigrate.Migration{
		Name: "20060102150405",
		Up: func(ctx context.Context, db *ch.DB) error {
			history = append(history, "up1")
			return nil
		},
		Down: func(ctx context.Context, db *ch.DB) error {
			history = append(history, "down1")
			return nil
		},
	})
	migrations.Add(chmigrate.Migration{
		Name: "20060102160405",
		Up: func(ctx context.Context, db *ch.DB) error {
			history = append(history, "up2")
			return nil
		},
		Down: func(ctx context.Context, db *ch.DB) error {
			history = append(history, "down2")
			return nil
		},
	})

	m := chmigrate.NewMigrator(db, migrations)
	err := m.Reset(ctx)
	require.NoError(t, err)

	group, err := m.Migrate(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), group.ID)
	require.Len(t, group.Migrations, 2)
	require.Equal(t, []string{"up1", "up2"}, history)

	history = nil
	group, err = m.Rollback(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), group.ID)
	require.Len(t, group.Migrations, 2)
	require.Equal(t, []string{"down2", "down1"}, history)
}

func testChmigrateUpError(t *testing.T, db *ch.DB) {
	ctx := context.Background()

	var history []string

	migrations := chmigrate.NewMigrations()
	migrations.Add(chmigrate.Migration{
		Name: "20060102150405",
		Up: func(ctx context.Context, db *ch.DB) error {
			history = append(history, "up1")
			return nil
		},
		Down: func(ctx context.Context, db *ch.DB) error {
			history = append(history, "down1")
			return nil
		},
	})
	migrations.Add(chmigrate.Migration{
		Name: "20060102160405",
		Up: func(ctx context.Context, db *ch.DB) error {
			history = append(history, "up2")
			return errors.New("failed")
		},
		Down: func(ctx context.Context, db *ch.DB) error {
			history = append(history, "down2")
			return nil
		},
	})
	migrations.Add(chmigrate.Migration{
		Name: "20060102170405",
		Up: func(ctx context.Context, db *ch.DB) error {
			history = append(history, "up3")
			return errors.New("failed")
		},
		Down: func(ctx context.Context, db *ch.DB) error {
			history = append(history, "down3")
			return nil
		},
	})

	m := chmigrate.NewMigrator(db, migrations)
	err := m.Reset(ctx)
	require.NoError(t, err)

	group, err := m.Migrate(ctx)
	require.Error(t, err)
	require.Equal(t, "failed", err.Error())
	require.Equal(t, int64(1), group.ID)
	require.Len(t, group.Migrations, 2)
	require.Equal(t, []string{"up1", "up2"}, history)

	history = nil
	group, err = m.Rollback(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), group.ID)
	require.Len(t, group.Migrations, 2)
	require.Equal(t, []string{"down2", "down1"}, history)
}

func chDB(opts ...ch.Option) *ch.DB {
	dsn := os.Getenv("CH")
	if dsn == "" {
		dsn = "clickhouse://localhost:9000/test?sslmode=disable"
	}

	opts = append(opts, ch.WithDSN(dsn))
	db := ch.Connect(opts...)
	db.AddQueryHook(chdebug.NewQueryHook(
		chdebug.WithEnabled(false),
		chdebug.FromEnv("CHDEBUG"),
	))
	return db
}

func funcName(x interface{}) string {
	s := runtime.FuncForPC(reflect.ValueOf(x).Pointer()).Name()
	if i := strings.LastIndexByte(s, '.'); i >= 0 {
		return s[i+1:]
	}
	return s
}
