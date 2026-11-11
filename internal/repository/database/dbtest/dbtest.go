// Package dbtest gives repository tests a freshly migrated database on every engine available:
// SQLite always, PostgreSQL when TEST_POSTGRES_DSN is set.
package dbtest

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	"github.com/faruq-rahmadani/masjid-digital/migrations"
	pkgdb "github.com/faruq-rahmadani/masjid-digital/pkg/database"
)

// DB is a migrated database, its client for repositories, and a user that rows can point at
type DB struct {
	*pkgdb.DB
	Client database.DatabaseClient
	UserID int64
}

// Open returns a migrated database per engine; each is rolled back and closed when the test ends
func Open(t *testing.T) map[pkgdb.Driver]DB {
	t.Helper()
	ctx := context.Background()

	cfgs := map[pkgdb.Driver]pkgdb.Config{
		pkgdb.DriverSQLite: {Driver: pkgdb.DriverSQLite, DSN: filepath.Join(t.TempDir(), "test.db")},
	}
	if dsn := os.Getenv("TEST_POSTGRES_DSN"); dsn != "" {
		cfgs[pkgdb.DriverPostgres] = pkgdb.Config{Driver: pkgdb.DriverPostgres, DSN: dsn}
	}

	dbs := make(map[pkgdb.Driver]DB, len(cfgs))
	for driver, cfg := range cfgs {
		db, err := pkgdb.Open(ctx, cfg)
		if err != nil {
			t.Fatalf("open %s: %v", driver, err)
		}
		m, err := db.NewMigrator(migrations.FS)
		if err != nil {
			t.Fatalf("new migrator on %s: %v", driver, err)
		}
		if _, err := m.Up(ctx); err != nil {
			t.Fatalf("migrate %s: %v", driver, err)
		}
		t.Cleanup(func() {
			for {
				if v, err := m.Version(ctx); err != nil || v == 0 {
					break
				}
				if _, err := m.Down(ctx); err != nil {
					break
				}
			}
			db.Close()
		})

		now := time.Now().UTC()
		var userID int64
		if err := db.Get(ctx, &userID,
			"INSERT INTO users (name, phone, password_hash, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?) RETURNING id",
			"Bendahara", "+6281200000001", "hash", "bendahara", now, now,
		); err != nil {
			t.Fatalf("insert user on %s: %v", driver, err)
		}

		dbs[driver] = DB{DB: db, Client: database.DatabaseClient{Querier: db}, UserID: userID}
	}
	return dbs
}
