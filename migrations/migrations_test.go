package migrations_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/migrations"
	"github.com/faruq-rahmadani/masjid-digital/pkg/database"
)

// TestMigrations applies every migration up, checks the schema, then rolls
// everything back, on each engine available
func TestMigrations(t *testing.T) {
	cfgs := map[database.Driver]database.Config{
		database.DriverSQLite: {Driver: database.DriverSQLite, DSN: filepath.Join(t.TempDir(), "test.db")},
	}
	if dsn := os.Getenv("TEST_POSTGRES_DSN"); dsn != "" {
		cfgs[database.DriverPostgres] = database.Config{Driver: database.DriverPostgres, DSN: dsn}
	}

	for driver, cfg := range cfgs {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			db, err := database.Open(ctx, cfg)
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			defer db.Close()

			m, err := db.NewMigrator(migrations.FS)
			if err != nil {
				t.Fatalf("new migrator: %v", err)
			}

			// schema as of version 2, with a user still on the old "user" role
			if _, err := m.UpTo(ctx, 2); err != nil {
				t.Fatalf("up to 2: %v", err)
			}

			now := time.Now().UTC()
			insert := "INSERT INTO users (name, phone, password_hash, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?) RETURNING id"

			var oldID int64
			if err := db.Get(ctx, &oldID, insert, "Lama", "+6281200000009", "hash", "user", now, now); err != nil {
				t.Fatalf("insert user before 00003: %v", err)
			}

			if _, err := m.Up(ctx); err != nil {
				t.Fatalf("up: %v", err)
			}

			var role string
			if err := db.Get(ctx, &role, "SELECT role FROM users WHERE id = ?", oldID); err != nil {
				t.Fatalf("get role: %v", err)
			}
			if role != "takmir" {
				t.Errorf("role after 00003 = %q, want takmir", role)
			}

			var id int64
			if err := db.Get(ctx, &id, insert, "Ahmad", "+6281234567890", "hash", "bendahara", now, now); err != nil {
				t.Fatalf("insert user: %v", err)
			}
			var dupID int64
			if err := db.Get(ctx, &dupID, insert, "Budi", "+6281234567890", "hash", "takmir", now, now); !errors.Is(err, database.ErrUniqueViolation) {
				t.Errorf("duplicate phone: err = %v, want ErrUniqueViolation", err)
			}

			// 00005: one override per prayer per date
			insertOverride := "INSERT INTO prayer_time_overrides (date, prayer, time, remarks, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
			if _, err := db.Exec(ctx, insertOverride, "2026-11-06", "dzuhur", "11:58", "Jumat", id, now, now); err != nil {
				t.Fatalf("insert override: %v", err)
			}
			if _, err := db.Exec(ctx, insertOverride, "2026-11-06", "dzuhur", "12:00", "", id, now, now); !errors.Is(err, database.ErrUniqueViolation) {
				t.Errorf("duplicate override: err = %v, want ErrUniqueViolation", err)
			}

			for {
				version, err := m.Version(ctx)
				if err != nil {
					t.Fatalf("version: %v", err)
				}
				if version == 0 {
					break
				}
				if _, err := m.Down(ctx); err != nil {
					t.Fatalf("down from %d: %v", version, err)
				}
			}
		})
	}
}
