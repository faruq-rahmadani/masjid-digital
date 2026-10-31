// Package database wraps PostgreSQL and SQLite behind one API so repositories
// never need to know which engine is running.
//
// Repositories write portable SQL with `?` placeholders and depend on the
// Querier interface. The wrapper takes care of:
//   - placeholder rebinding (`?` -> `$1` on PostgreSQL)
//   - engine-specific errors -> ErrNotFound, ErrUniqueViolation, ErrForeignKeyViolation
//   - transactions carried in context.Context (see Transactor)
//   - SQLite defaults: WAL, foreign keys, busy timeout, immediate transactions
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// Driver is a supported database engine
type Driver string

const (
	DriverPostgres Driver = "postgres"
	DriverSQLite   Driver = "sqlite"
)

// driverNames maps a Driver to the database/sql driver name it registers
var driverNames = map[Driver]string{
	DriverPostgres: "pgx",
	DriverSQLite:   "sqlite",
}

func init() {
	// sqlx does not know modernc's "sqlite" driver name yet
	sqlx.BindDriver(driverNames[DriverSQLite], sqlx.QUESTION)
}

// Config holds the connection settings of a database
type Config struct {
	Driver          Driver
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Querier is the only database dependency a repository needs.
// Every method joins the transaction stored in ctx, if any.
type Querier interface {
	// Exec runs a statement that returns no rows
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	// Get scans a single row into dest, returning ErrNotFound when there is none
	Get(ctx context.Context, dest any, query string, args ...any) error
	// Select scans all rows into dest, which must be a pointer to a slice
	Select(ctx context.Context, dest any, query string, args ...any) error
}

// Transactor runs fn inside a transaction. It is meant for the usecase layer,
// which groups several repository calls into one atomic unit.
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// DB is a connection pool to either PostgreSQL or SQLite
type DB struct {
	db     *sqlx.DB
	driver Driver
}

var (
	_ Querier    = (*DB)(nil)
	_ Transactor = (*DB)(nil)
)

// Open connects to the database described by cfg and verifies the connection
func Open(ctx context.Context, cfg Config) (*DB, error) {
	driverName, ok := driverNames[cfg.Driver]
	if !ok {
		return nil, fmt.Errorf("unsupported database driver %q, must be one of: %s, %s", cfg.Driver, DriverPostgres, DriverSQLite)
	}

	dsn := cfg.DSN
	if cfg.Driver == DriverSQLite {
		var err error
		if dsn, err = sqliteDSN(dsn); err != nil {
			return nil, err
		}
	}

	db, err := sqlx.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening %s database: %w", cfg.Driver, err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to %s database: %w", cfg.Driver, err)
	}

	return &DB{db: db, driver: cfg.Driver}, nil
}

// Driver returns the engine in use. Repositories should not branch on it;
// it exists for wiring code such as picking a migrations directory.
func (d *DB) Driver() Driver {
	return d.driver
}

// Close closes the connection pool
func (d *DB) Close() error {
	return d.db.Close()
}

// Ping verifies the database is reachable, e.g. for a health check
func (d *DB) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	res, err := d.conn(ctx).ExecContext(ctx, d.db.Rebind(query), args...)
	return res, d.translate(err)
}

func (d *DB) Get(ctx context.Context, dest any, query string, args ...any) error {
	return d.translate(sqlx.GetContext(ctx, d.conn(ctx), dest, d.db.Rebind(query), args...))
}

func (d *DB) Select(ctx context.Context, dest any, query string, args ...any) error {
	return d.translate(sqlx.SelectContext(ctx, d.conn(ctx), dest, d.db.Rebind(query), args...))
}

// conn returns the transaction stored in ctx, or the pool when there is none
func (d *DB) conn(ctx context.Context) sqlx.ExtContext {
	if tx := txFromContext(ctx); tx != nil {
		return tx
	}
	return d.db
}
