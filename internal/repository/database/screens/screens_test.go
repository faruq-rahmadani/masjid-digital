package screens

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	"github.com/faruq-rahmadani/masjid-digital/migrations"
	pkgdb "github.com/faruq-rahmadani/masjid-digital/pkg/database"
)

// testRepository is a repository on a freshly migrated database, with the ID of a user that
// screens can point at
type testRepository struct {
	*repo
	db     *pkgdb.DB
	userID int64
}

// newTestRepositories returns a repository on a freshly migrated database for every engine
// available
func newTestRepositories(t *testing.T) map[pkgdb.Driver]testRepository {
	t.Helper()
	ctx := context.Background()

	cfgs := map[pkgdb.Driver]pkgdb.Config{
		pkgdb.DriverSQLite: {Driver: pkgdb.DriverSQLite, DSN: filepath.Join(t.TempDir(), "test.db")},
	}
	if dsn := os.Getenv("TEST_POSTGRES_DSN"); dsn != "" {
		cfgs[pkgdb.DriverPostgres] = pkgdb.Config{Driver: pkgdb.DriverPostgres, DSN: dsn}
	}

	repos := make(map[pkgdb.Driver]testRepository, len(cfgs))
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
			"Ahmad", "+6281200000001", "hash", "takmir", now, now,
		); err != nil {
			t.Fatalf("insert user on %s: %v", driver, err)
		}

		repos[driver] = testRepository{
			repo:   New(database.DatabaseClient{Querier: db}),
			db:     db,
			userID: userID,
		}
	}
	return repos
}

func TestRegisterOncePerDevice(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			first, err := repo.Register(ctx, domain.Screen{DeviceID: "device-0001", NameHint: "teras", UserAgent: "Chrome Android"})
			if err != nil {
				t.Fatalf("register: %v", err)
			}
			if first.ID == 0 || first.Name != "" || first.NameHint != "teras" || first.ForcedMode != domain.ScreenModeAuto || first.BlockedAt != nil {
				t.Errorf("registered = %+v", first)
			}

			// a second connect of the same device keeps its row
			again, err := repo.Register(ctx, domain.Screen{DeviceID: "device-0001", NameHint: "other"})
			if err != nil || again.ID != first.ID || again.NameHint != "teras" {
				t.Errorf("again = %+v, %v; want the first row unchanged", again, err)
			}

			if _, err := repo.Register(ctx, domain.Screen{DeviceID: "device-0002"}); err != nil {
				t.Fatalf("second device: %v", err)
			}
			all, err := repo.FindAll(ctx)
			if err != nil || len(all) != 2 || all[0].DeviceID != "device-0001" {
				t.Errorf("all = %+v, %v; want 2, oldest first", all, err)
			}

			if _, err := repo.FindByDeviceID(ctx, "device-9999"); !errors.Is(err, domain.ErrScreenNotFound) {
				t.Errorf("unknown device: err = %v, want ErrScreenNotFound", err)
			}
		})
	}
}

func TestUpdateScreen(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()
			screen, err := repo.Register(ctx, domain.Screen{DeviceID: "device-0001"})
			if err != nil {
				t.Fatalf("register: %v", err)
			}

			until := time.Date(2026, 11, 2, 17, 0, 0, 0, time.UTC)
			blocked := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
			screen.Name, screen.ForcedMode, screen.ForcedUntil, screen.BlockedAt = "TV Teras", domain.ScreenModeDark, &until, &blocked
			got, err := repo.Update(ctx, screen)
			if err != nil {
				t.Fatalf("update: %v", err)
			}
			if got.Name != "TV Teras" || got.ForcedMode != domain.ScreenModeDark || got.ForcedUntil == nil || !got.ForcedUntil.Equal(until) || got.BlockedAt == nil || !got.BlockedAt.Equal(blocked) {
				t.Errorf("updated = %+v", got)
			}

			// back to automatic and unblocked clears both times
			got.ForcedMode, got.ForcedUntil, got.BlockedAt = domain.ScreenModeAuto, nil, nil
			if cleared, err := repo.Update(ctx, got); err != nil || cleared.ForcedUntil != nil || cleared.BlockedAt != nil {
				t.Errorf("cleared = %+v, %v", cleared, err)
			}

			if _, err := repo.Update(ctx, domain.Screen{ID: 999, ForcedMode: domain.ScreenModeAuto}); !errors.Is(err, domain.ErrScreenNotFound) {
				t.Errorf("missing: err = %v, want ErrScreenNotFound", err)
			}
		})
	}
}
