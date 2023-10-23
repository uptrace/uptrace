package chmigrate

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/uptrace/go-clickhouse/ch"
)

type MigratorOption func(m *Migrator)

func WithTableName(table string) MigratorOption {
	return func(m *Migrator) {
		m.migrationsTable = table
	}
}

func WithLocksTableName(table string) MigratorOption {
	return func(m *Migrator) {
		m.locksTable = table
	}
}

func WithReplicated(on bool) MigratorOption {
	return func(m *Migrator) {
		m.replicated = on
	}
}

func WithDistributed(on bool) MigratorOption {
	return func(m *Migrator) {
		m.distributed = on
	}
}

func WithOnCluster(cluster string) MigratorOption {
	return func(m *Migrator) {
		m.cluster = cluster
	}
}

// WithMarkAppliedOnSuccess sets the migrator to only mark migrations as applied/unapplied
// when their up/down is successful.
func WithMarkAppliedOnSuccess(enabled bool) MigratorOption {
	return func(m *Migrator) {
		m.markAppliedOnSuccess = enabled
	}
}

type Migrator struct {
	db         *ch.DB
	migrations *Migrations

	ms MigrationSlice

	migrationsTable      string
	locksTable           string
	replicated           bool
	distributed          bool
	cluster              string
	markAppliedOnSuccess bool
}

func NewMigrator(db *ch.DB, migrations *Migrations, opts ...MigratorOption) *Migrator {
	m := &Migrator{
		db:         db,
		migrations: migrations,

		ms: migrations.ms,

		migrationsTable: "ch_migrations",
		locksTable:      "ch_migration_locks",
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Migrator) DB() *ch.DB {
	return m.db
}

// MigrationsWithStatus returns migrations with status in ascending order.
func (m *Migrator) MigrationsWithStatus(ctx context.Context) (MigrationSlice, error) {
	sorted, _, err := m.migrationsWithStatus(ctx)
	return sorted, err
}

func (m *Migrator) migrationsWithStatus(ctx context.Context) (MigrationSlice, int64, error) {
	sorted := m.migrations.Sorted()

	applied, err := m.AppliedMigrations(ctx)
	if err != nil {
		return nil, 0, err
	}

	appliedMap := migrationMap(applied)
	for i := range sorted {
		m1 := &sorted[i]
		if m2, ok := appliedMap[m1.Name]; ok {
			m1.GroupID = m2.GroupID
			m1.MigratedAt = m2.MigratedAt
		}
	}

	return sorted, applied.LastGroupID(), nil
}

func (m *Migrator) Init(ctx context.Context) error {
	if m.distributed {
		if m.cluster == "" {
			return errors.New("chmigrate: distributed requires a cluster name")
		}
	}

	if _, err := m.db.NewCreateTable().
		Model((*Migration)(nil)).
		Apply(func(q *ch.CreateTableQuery) *ch.CreateTableQuery {
			if m.replicated {
				return q.Engine("ReplicatedCollapsingMergeTree(sign)")
			}
			return q.Engine("CollapsingMergeTree(sign)")
		}).
		ModelTable(m.migrationsTable).
		OnCluster(m.cluster).
		IfNotExists().
		Exec(ctx); err != nil {
		return err
	}

	if _, err := m.db.NewCreateTable().
		Model((*migrationLock)(nil)).
		Apply(func(q *ch.CreateTableQuery) *ch.CreateTableQuery {
			if m.replicated {
				return q.Engine("ReplicatedMergeTree")
			}
			return q.Engine("MergeTree")
		}).
		ModelTable(m.locksTable).
		OnCluster(m.cluster).
		IfNotExists().
		Exec(ctx); err != nil {
		return err
	}

	if m.distributed {
		if _, err := m.db.NewCreateTable().
			Table(m.distTable(m.migrationsTable)).
			As(m.migrationsTable).
			Engine("Distributed(?, currentDatabase(), ?, rand())",
				ch.Ident(m.cluster), ch.Ident(m.migrationsTable)).
			OnCluster(m.cluster).
			IfNotExists().
			Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) Reset(ctx context.Context) error {
	tables := []string{
		m.migrationsTable,
		m.locksTable,
	}
	if m.distributed {
		tables = append(tables,
			m.distTable(m.migrationsTable),
		)
	}
	for _, tableName := range tables {
		if _, err := m.db.NewDropTable().
			Table(tableName).
			OnCluster(m.cluster).
			IfExists().
			Exec(ctx); err != nil {
			return err
		}
	}
	return m.Init(ctx)
}

// Migrate runs unapplied migrations. If a migration fails, migrate immediately exits.
func (m *Migrator) Migrate(ctx context.Context, opts ...MigrationOption) (*MigrationGroup, error) {
	cfg := newMigrationConfig(opts)

	if err := m.validate(); err != nil {
		return nil, err
	}

	migrations, lastGroupID, err := m.migrationsWithStatus(ctx)
	if err != nil {
		return nil, err
	}

	// Get the unapplied migrations we will range over and apply
	migrations = migrations.Unapplied()
	if len(migrations) == 0 {
		return &MigrationGroup{}, nil
	}

	// Create empty group to track applied migrations
	applied := new(MigrationGroup)
	applied.ID = lastGroupID + 1

	for i := range migrations {
		migration := &migrations[i]
		migration.GroupID = applied.ID

		if !m.markAppliedOnSuccess {
			if err := m.MarkApplied(ctx, migration); err != nil {
				return applied, err
			}

			// Add to the group before upping the migration when always marking applied
			applied.Migrations = append(applied.Migrations, *migration)
		}

		if !cfg.nop && migration.Up != nil {
			if err := migration.Up(ctx, m.db); err != nil {
				return applied, err
			}
		}

		if m.markAppliedOnSuccess {
			if err := m.MarkApplied(ctx, migration); err != nil {
				return applied, err
			}

			// Add to the group after upping the migration when marking applied on success
			applied.Migrations = append(applied.Migrations, *migration)
		}
	}

	return applied, nil
}

func (m *Migrator) Rollback(ctx context.Context, opts ...MigrationOption) (*MigrationGroup, error) {
	cfg := newMigrationConfig(opts)

	if err := m.validate(); err != nil {
		return nil, err
	}

	migrations, err := m.MigrationsWithStatus(ctx)
	if err != nil {
		return nil, err
	}

	lastGroup := migrations.LastGroup()

	// Create empty group to track unapplied migrations
	unapplied := new(MigrationGroup)
	unapplied.ID = lastGroup.ID

	for i := len(lastGroup.Migrations) - 1; i >= 0; i-- {
		migration := &lastGroup.Migrations[i]

		if !m.markAppliedOnSuccess {
			if err := m.MarkUnapplied(ctx, migration); err != nil {
				return unapplied, err
			}

			// Add to the group before downing the migration when always marking unapplied
			unapplied.Migrations = append(unapplied.Migrations, *migration)
		}

		if !cfg.nop && migration.Down != nil {
			if err := migration.Down(ctx, m.db); err != nil {
				return unapplied, err
			}
		}

		if m.markAppliedOnSuccess {
			if err := m.MarkUnapplied(ctx, migration); err != nil {
				return unapplied, err
			}

			// Add to the group after downing the migration when marking unapplied on success
			unapplied.Migrations = append(unapplied.Migrations, *migration)
		}
	}

	return lastGroup, nil
}

type goMigrationConfig struct {
	packageName string
}

type GoMigrationOption func(cfg *goMigrationConfig)

func WithPackageName(name string) GoMigrationOption {
	return func(cfg *goMigrationConfig) {
		cfg.packageName = name
	}
}

// CreateGoMigration creates a Go migration file.
func (m *Migrator) CreateGoMigration(
	ctx context.Context, name string, opts ...GoMigrationOption,
) (*MigrationFile, error) {
	cfg := &goMigrationConfig{
		packageName: "migrations",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	name, err := m.genMigrationName(name)
	if err != nil {
		return nil, err
	}

	fname := name + ".go"
	fpath := filepath.Join(m.migrations.getDirectory(), fname)
	content := fmt.Sprintf(goTemplate, cfg.packageName)

	if err := ioutil.WriteFile(fpath, []byte(content), 0o644); err != nil {
		return nil, err
	}

	mf := &MigrationFile{
		Name:    fname,
		Path:    fpath,
		Content: content,
	}
	return mf, nil
}

// CreateSQLMigrations creates an up and down SQL migration files.
func (m *Migrator) CreateSQLMigrations(ctx context.Context, name string) ([]*MigrationFile, error) {
	name, err := m.genMigrationName(name)
	if err != nil {
		return nil, err
	}

	up, err := m.createSQL(ctx, name+".up.sql")
	if err != nil {
		return nil, err
	}

	down, err := m.createSQL(ctx, name+".down.sql")
	if err != nil {
		return nil, err
	}

	return []*MigrationFile{up, down}, nil
}

func (m *Migrator) createSQL(ctx context.Context, fname string) (*MigrationFile, error) {
	fpath := filepath.Join(m.migrations.getDirectory(), fname)

	if err := ioutil.WriteFile(fpath, []byte(sqlTemplate), 0o644); err != nil {
		return nil, err
	}

	mf := &MigrationFile{
		Name:    fname,
		Path:    fpath,
		Content: goTemplate,
	}
	return mf, nil
}

var nameRE = regexp.MustCompile(`^[0-9a-z_\-]+$`)

func (m *Migrator) genMigrationName(name string) (string, error) {
	const timeFormat = "20060102150405"

	if name == "" {
		return "", errors.New("chmigrate: migration name can't be empty")
	}
	if !nameRE.MatchString(name) {
		return "", fmt.Errorf("chmigrate: invalid migration name: %q", name)
	}

	version := time.Now().UTC().Format(timeFormat)
	return fmt.Sprintf("%s_%s", version, name), nil
}

// MarkApplied marks the migration as applied (completed).
func (m *Migrator) MarkApplied(ctx context.Context, migration *Migration) error {
	migration.Sign = 1
	migration.MigratedAt = time.Now()
	_, err := m.db.NewInsert().
		Model(migration).
		ModelTable(m.distTable(m.migrationsTable)).
		Exec(ctx)
	return err
}

// MarkUnapplied marks the migration as unapplied (new).
func (m *Migrator) MarkUnapplied(ctx context.Context, migration *Migration) error {
	migration.Sign = -1
	_, err := m.db.NewInsert().
		Model(migration).
		ModelTable(m.distTable(m.migrationsTable)).
		Exec(ctx)
	return err
}

func (m *Migrator) TruncateTable(ctx context.Context) error {
	_, err := m.db.Exec("TRUNCATE TABLE ?", ch.Ident(m.distTable(m.migrationsTable)))
	return err
}

// MissingMigrations returns applied migrations that can no longer be found.
func (m *Migrator) MissingMigrations(ctx context.Context) (MigrationSlice, error) {
	applied, err := m.AppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	existing := migrationMap(m.migrations.ms)
	for i := len(applied) - 1; i >= 0; i-- {
		m := &applied[i]
		if _, ok := existing[m.Name]; ok {
			applied = append(applied[:i], applied[i+1:]...)
		}
	}

	return applied, nil
}

// AppliedMigrations selects applied migrations in descending order.
func (m *Migrator) AppliedMigrations(ctx context.Context) (MigrationSlice, error) {
	var ms MigrationSlice
	if err := m.db.NewSelect().
		ColumnExpr("*").
		Model(&ms).
		ModelTable(m.distTable(m.migrationsTable)).
		Final().
		Scan(ctx); err != nil {
		return nil, err
	}
	return ms, nil
}

func (m *Migrator) validate() error {
	if len(m.ms) == 0 {
		return errors.New("chmigrate: there are no any migrations")
	}
	return nil
}

func (m *Migrator) distTable(table string) string {
	if m.distributed {
		return table + "_dist"
	}
	return table
}

//------------------------------------------------------------------------------

type migrationLock struct {
	A int8
}

func (m *Migrator) Lock(ctx context.Context) error {
	if _, err := m.db.ExecContext(
		ctx,
		"ALTER TABLE ? ADD COLUMN ? Int8",
		ch.Safe(m.locksTable), ch.Safe("lock"),
	); err != nil {
		return fmt.Errorf("chmigrate: migrations table is already locked (%w)", err)
	}
	return nil
}

func (m *Migrator) Unlock(ctx context.Context) error {
	if _, err := m.db.ExecContext(
		ctx,
		"ALTER TABLE ? DROP COLUMN ?",
		ch.Safe(m.locksTable), ch.Safe("lock"),
	); err != nil && !strings.Contains(err.Error(), "Cannot find column") {
		return fmt.Errorf("chmigrate: migrations table is already unlocked (%w)", err)
	}
	return nil
}

func migrationMap(ms MigrationSlice) map[string]*Migration {
	mp := make(map[string]*Migration)
	for i := range ms {
		m := &ms[i]
		mp[m.Name] = m
	}
	return mp
}
