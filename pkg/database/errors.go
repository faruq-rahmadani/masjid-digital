package database

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// Engine-agnostic errors. Repositories check these with errors.Is and map them
// to domain errors, e.g. ErrNotFound -> domain.ErrMasjidNotFound.
var (
	ErrNotFound            = errors.New("database: record not found")
	ErrUniqueViolation     = errors.New("database: unique constraint violation")
	ErrForeignKeyViolation = errors.New("database: foreign key constraint violation")
)

// PostgreSQL SQLSTATE codes, see https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

// translate maps an engine-specific error to one of the errors above,
// keeping the original in the chain for logging
func (d *DB) translate(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return errors.Join(ErrUniqueViolation, err)
		case pgForeignKeyViolation:
			return errors.Join(ErrForeignKeyViolation, err)
		}
	}

	var liteErr *sqlite.Error
	if errors.As(err, &liteErr) {
		switch liteErr.Code() {
		case sqlite3.SQLITE_CONSTRAINT_UNIQUE, sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY:
			return errors.Join(ErrUniqueViolation, err)
		case sqlite3.SQLITE_CONSTRAINT_FOREIGNKEY:
			return errors.Join(ErrForeignKeyViolation, err)
		}
	}

	return err
}
