package migrations_test

import (
	"bytes"
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

			// 00006: one petugas Jumat row per date
			insertDuty := "INSERT INTO friday_duties (date, khatib, updated_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?)"
			if _, err := db.Exec(ctx, insertDuty, "2026-11-06", "Ustadz Abdullah", id, now, now); err != nil {
				t.Fatalf("insert friday duty: %v", err)
			}
			if _, err := db.Exec(ctx, insertDuty, "2026-11-06", "Ustadz Hasan", id, now, now); !errors.Is(err, database.ErrUniqueViolation) {
				t.Errorf("duplicate friday duty: err = %v, want ErrUniqueViolation", err)
			}

			// 00007: a running text is active unless told otherwise
			var active bool
			if err := db.Get(ctx, &active, "INSERT INTO running_texts (text, position, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?) RETURNING active", "Kajian Tafsir", 0, id, now, now); err != nil || !active {
				t.Errorf("insert running text: active = %v, %v; want true", active, err)
			}

			// 00009: a slide keeps its image bytes as is
			image := []byte{0x89, 'P', 'N', 'G', 0x00, 0xff}
			var slideID int64
			if err := db.Get(ctx, &slideID, "INSERT INTO slides (title, duration_seconds, position, created_by, created_at, updated_at, image_type, image_data) VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING id", "Kajian", 10, 0, id, now, now, "image/png", image); err != nil {
				t.Fatalf("insert slide: %v", err)
			}
			var stored []byte
			if err := db.Get(ctx, &stored, "SELECT image_data FROM slides WHERE id = ?", slideID); err != nil || !bytes.Equal(stored, image) {
				t.Errorf("slide image = %v, %v; want %v", stored, err, image)
			}

			// 00010: slides saved before it keep showing their caption
			var showCaption bool
			if err := db.Get(ctx, &showCaption, "SELECT show_caption FROM slides WHERE id = ?", slideID); err != nil || !showCaption {
				t.Errorf("show_caption = %v, %v; want true", showCaption, err)
			}

			// 00011: one screen row per device, in automatic mode until an admin forces one
			insertScreen := "INSERT INTO screens (device_id, first_seen_at, updated_at) VALUES (?, ?, ?)"
			if _, err := db.Exec(ctx, insertScreen, "tv-1", now, now); err != nil {
				t.Fatalf("insert screen: %v", err)
			}
			if _, err := db.Exec(ctx, insertScreen, "tv-1", now, now); !errors.Is(err, database.ErrUniqueViolation) {
				t.Errorf("duplicate device: err = %v, want ErrUniqueViolation", err)
			}
			var forcedMode string
			if err := db.Get(ctx, &forcedMode, "SELECT forced_mode FROM screens WHERE device_id = ?", "tv-1"); err != nil || forcedMode != "auto" {
				t.Errorf("forced_mode = %q, %v; want auto", forcedMode, err)
			}

			// 00013: the built-in pockets, Tunai and the system categories are seeded
			var pocketKinds []string
			if err := db.Select(ctx, &pocketKinds, "SELECT kind FROM pockets ORDER BY id"); err != nil || len(pocketKinds) != 4 {
				t.Errorf("seeded pockets = %v, %v; want umum, zakat_mal, zakat_fitrah, fidyah", pocketKinds, err)
			}
			insertPocket := "INSERT INTO pockets (kind, name, created_at, updated_at) VALUES (?, ?, ?, ?)"
			if _, err := db.Exec(ctx, insertPocket, "zakat_mal", "Zakat Mal 2", now, now); !errors.Is(err, database.ErrUniqueViolation) {
				t.Errorf("second zakat_mal pocket: err = %v, want ErrUniqueViolation", err)
			}
			if _, err := db.Exec(ctx, insertPocket, "terikat", "Renovasi Atap", now, now); err != nil {
				t.Errorf("terikat pocket: %v", err)
			}
			var asnaf int
			if err := db.Get(ctx, &asnaf, "SELECT COUNT(*) FROM categories WHERE code LIKE 'asnaf_%'"); err != nil || asnaf != 8 {
				t.Errorf("asnaf categories = %d, %v; want 8", asnaf, err)
			}
			var tunaiID, umumID, infaqID int64
			if err := db.Get(ctx, &tunaiID, "SELECT id FROM accounts WHERE name = 'Tunai'"); err != nil {
				t.Fatalf("seeded Tunai: %v", err)
			}
			_ = db.Get(ctx, &umumID, "SELECT id FROM pockets WHERE kind = 'umum'")
			_ = db.Get(ctx, &infaqID, "SELECT id FROM categories WHERE name = 'Infaq Jumat'")
			insertTx := "INSERT INTO transactions (kind, amount, date, account_id, pocket_id, category_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
			if _, err := db.Exec(ctx, insertTx, "in", 0, "2026-11-06", tunaiID, umumID, infaqID, now, now); err == nil {
				t.Error("zero amount accepted")
			}
			if _, err := db.Exec(ctx, insertTx, "in", 2000000, "2026-11-06", tunaiID, umumID, infaqID, now, now); err != nil {
				t.Errorf("insert transaction: %v", err)
			}
			insertClosing := "INSERT INTO book_closings (month, kind, closing_balance, closed_at) VALUES (?, ?, ?, ?)"
			if _, err := db.Exec(ctx, insertClosing, "2026-10", "opening", 0, now); err != nil {
				t.Fatalf("insert closing: %v", err)
			}
			if _, err := db.Exec(ctx, insertClosing, "2026-10", "closing", 0, now); !errors.Is(err, database.ErrUniqueViolation) {
				t.Errorf("second active closing of a month: err = %v, want ErrUniqueViolation", err)
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
