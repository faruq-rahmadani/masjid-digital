package database

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// gooseDialects maps a Driver to its goose dialect
var gooseDialects = map[Driver]goose.Dialect{
	DriverPostgres: goose.DialectPostgres,
	DriverSQLite:   goose.DialectSQLite3,
}

// MigrationResult describes one applied or rolled back migration
type MigrationResult = goose.MigrationResult

// MigrationStatus describes whether a migration is applied or pending
type MigrationStatus = goose.MigrationStatus

// Migrator applies schema migrations with goose
type Migrator struct {
	provider *goose.Provider
}

// NewMigrator reads migrations from the directory of fsys named after the
// driver in use (sqlite/ or postgres/), see the migrations package
func (d *DB) NewMigrator(fsys fs.FS) (*Migrator, error) {
	dir, err := fs.Sub(fsys, string(d.driver))
	if err != nil {
		return nil, fmt.Errorf("error opening %s migrations: %w", d.driver, err)
	}

	var opts []goose.ProviderOption
	if d.driver == DriverPostgres {
		// Several app instances may start at once; only one migrates at a time
		locker, err := lock.NewPostgresSessionLocker()
		if err != nil {
			return nil, fmt.Errorf("error creating migration lock: %w", err)
		}
		opts = append(opts, goose.WithSessionLocker(locker))
	}

	provider, err := goose.NewProvider(gooseDialects[d.driver], d.db.DB, dir, opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating migrator: %w", err)
	}

	return &Migrator{provider: provider}, nil
}

// Up applies all pending migrations
func (m *Migrator) Up(ctx context.Context) ([]*MigrationResult, error) {
	return m.provider.Up(ctx)
}

// UpTo applies pending migrations up to and including version
func (m *Migrator) UpTo(ctx context.Context, version int64) ([]*MigrationResult, error) {
	return m.provider.UpTo(ctx, version)
}

// Down rolls back the latest applied migration
func (m *Migrator) Down(ctx context.Context) (*MigrationResult, error) {
	return m.provider.Down(ctx)
}

// Status lists every migration and whether it is applied
func (m *Migrator) Status(ctx context.Context) ([]*MigrationStatus, error) {
	return m.provider.Status(ctx)
}

// Version returns the latest applied migration version
func (m *Migrator) Version(ctx context.Context) (int64, error) {
	return m.provider.GetDBVersion(ctx)
}
