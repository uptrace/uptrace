package chmigrate

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/uptrace/pkg/clickhouse/ch"
	"io"
	"io/fs"
	"sort"
	"strings"
	"time"
)

type Migration struct {
	Name       string `ch:",pk"`
	Comment    string `ch:"-"`
	GroupID    int64
	MigratedAt time.Time
	Sign       int8
	Up         MigrationFunc `ch:"-"`
	Down       MigrationFunc `ch:"-"`
}

func (m *Migration) String() string  { return fmt.Sprintf("%s_%s", m.Name, m.Comment) }
func (m *Migration) IsApplied() bool { return !m.MigratedAt.IsZero() }

type MigrationFunc func(ctx context.Context, db *ch.DB) error

func NewSQLMigrationFunc(fsys fs.FS, name string) MigrationFunc {
	return func(ctx context.Context, db *ch.DB) error {
		f, err := fsys.Open(name)
		if err != nil {
			return err
		}
		return Exec(ctx, db, f)
	}
}
func Exec(ctx context.Context, db *ch.DB, f io.Reader) error {
	scanner := bufio.NewScanner(f)
	var queries []string
	var query []byte
	for scanner.Scan() {
		b := scanner.Bytes()
		const prefix = "--migration:"
		if bytes.HasPrefix(b, []byte(prefix)) {
			b = b[len(prefix):]
			if bytes.Equal(b, []byte("split")) {
				queries = append(queries, string(query))
				query = query[:0]
				continue
			}
			return fmt.Errorf("ch: unknown directive: %q", b)
		}
		query = append(query, b...)
		query = append(query, '\n')
	}
	if len(query) > 0 {
		queries = append(queries, string(query))
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	for _, q := range queries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

const goTemplate = `package %s

import (
	"context"
	"fmt"

	"github.com/uptrace/pkg/clickhouse/ch"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *ch.DB) error {
		fmt.Print(" [up migration] ")
		return nil
	}, func(ctx context.Context, db *ch.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}
`
const sqlTemplate = `SELECT 1

--migration:split

SELECT 2
`

type MigrationSlice []Migration

func (ms MigrationSlice) String() string {
	if len(ms) == 0 {
		return "empty"
	}
	if len(ms) > 5 {
		return fmt.Sprintf("%d migrations (%s ... %s)", len(ms), ms[0].Name, ms[len(ms)-1].Name)
	}
	var sb strings.Builder
	for i := range ms {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(ms[i].String())
	}
	return sb.String()
}
func (ms MigrationSlice) Applied() MigrationSlice {
	var applied MigrationSlice
	for i := range ms {
		if ms[i].IsApplied() {
			applied = append(applied, ms[i])
		}
	}
	sortDesc(applied)
	return applied
}
func (ms MigrationSlice) Unapplied() MigrationSlice {
	var unapplied MigrationSlice
	for i := range ms {
		if !ms[i].IsApplied() {
			unapplied = append(unapplied, ms[i])
		}
	}
	sortAsc(unapplied)
	return unapplied
}
func (ms MigrationSlice) LastGroupID() int64 {
	var lastGroupID int64
	for i := range ms {
		groupID := ms[i].GroupID
		if groupID > lastGroupID {
			lastGroupID = groupID
		}
	}
	return lastGroupID
}
func (ms MigrationSlice) LastGroup() *MigrationGroup {
	group := &MigrationGroup{ID: ms.LastGroupID()}
	if group.ID == 0 {
		return group
	}
	for i := range ms {
		if ms[i].GroupID == group.ID {
			group.Migrations = append(group.Migrations, ms[i])
		}
	}
	return group
}

type MigrationGroup struct {
	ID         int64
	Migrations MigrationSlice
}

func (g *MigrationGroup) IsZero() bool { return g.ID == 0 && len(g.Migrations) == 0 }
func (g *MigrationGroup) String() string {
	if g.IsZero() {
		return "nil"
	}
	return fmt.Sprintf("group #%d (%s)", g.ID, g.Migrations)
}

type MigrationFile struct {
	Name    string
	Path    string
	Content string
}
type migrationConfig struct{ nop bool }

func newMigrationConfig(opts []MigrationOption) *migrationConfig {
	cfg := new(migrationConfig)
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

type MigrationOption func(cfg *migrationConfig)

func WithNopMigration() MigrationOption { return func(cfg *migrationConfig) { cfg.nop = true } }
func sortAsc(ms MigrationSlice) {
	sort.Slice(ms, func(i, j int) bool { return ms[i].Name < ms[j].Name })
}
func sortDesc(ms MigrationSlice) {
	sort.Slice(ms, func(i, j int) bool { return ms[i].Name > ms[j].Name })
}
