package runningtexts

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
// running texts can point at
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
			repo: New(database.DatabaseClient{Querier: db}),
			db:                      db,
			userID:                  userID,
		}
	}
	return repos
}

// texts returns the texts in display order
func texts(t *testing.T, repo testRepository) []string {
	t.Helper()
	all, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("find all: %v", err)
	}
	got := make([]string, 0, len(all))
	for _, rt := range all {
		got = append(got, rt.Text)
	}
	return got
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestCreateAppendsActive(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			first, err := repo.Create(ctx, domain.RunningText{Text: "Kajian Tafsir", CreatedBy: repo.userID})
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			if first.ID == 0 || !first.Active || first.Position != 0 || first.CreatedBy != repo.userID || first.CreatedByName != "Ahmad" || first.DeactivatedAt != nil || first.CreatedAt.IsZero() {
				t.Errorf("created = %+v", first)
			}

			second, err := repo.Create(ctx, domain.RunningText{Text: "Infaq via QRIS"})
			if err != nil {
				t.Fatalf("create second: %v", err)
			}
			if second.Position != 1 || second.CreatedBy != 0 {
				t.Errorf("second = %+v, want position 1 and no creator", second)
			}

			if got := texts(t, repo); !equal(got, []string{"Kajian Tafsir", "Infaq via QRIS"}) {
				t.Errorf("texts = %v", got)
			}
		})
	}
}

func TestUpdateAndDelete(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			created, err := repo.Create(ctx, domain.RunningText{Text: "Kajian Tafsir"})
			if err != nil {
				t.Fatalf("create: %v", err)
			}

			updated, err := repo.Update(ctx, domain.RunningText{ID: created.ID, Text: "Kajian Fiqih", Active: false})
			if err != nil {
				t.Fatalf("update: %v", err)
			}
			if updated.Text != "Kajian Fiqih" || updated.Active || updated.Position != created.Position || !updated.CreatedAt.Equal(created.CreatedAt) {
				t.Errorf("updated = %+v", updated)
			}

			if _, err := repo.Update(ctx, domain.RunningText{ID: created.ID + 100, Text: "x"}); !errors.Is(err, domain.ErrRunningTextNotFound) {
				t.Errorf("update unknown: err = %v, want ErrRunningTextNotFound", err)
			}

			if err := repo.Delete(ctx, created.ID); err != nil {
				t.Fatalf("delete: %v", err)
			}
			if err := repo.Delete(ctx, created.ID); !errors.Is(err, domain.ErrRunningTextNotFound) {
				t.Errorf("delete again: err = %v, want ErrRunningTextNotFound", err)
			}
		})
	}
}

func TestDeactivatedAt(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			created, err := repo.Create(ctx, domain.RunningText{Text: "Kajian Tafsir", CreatedBy: repo.userID})
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			update := func(text string, active bool) domain.RunningText {
				t.Helper()
				rt, err := repo.Update(ctx, domain.RunningText{ID: created.ID, Text: text, Active: active})
				if err != nil {
					t.Fatalf("update: %v", err)
				}
				if rt.CreatedByName != "Ahmad" {
					t.Errorf("created by name = %q, want Ahmad", rt.CreatedByName)
				}
				return rt
			}

			off := update("Kajian Tafsir", false)
			if off.DeactivatedAt == nil {
				t.Fatal("turned off: want deactivated_at")
			}

			time.Sleep(10 * time.Millisecond)
			edited := update("Kajian Fiqih", false)
			if edited.DeactivatedAt == nil || !edited.DeactivatedAt.Equal(*off.DeactivatedAt) {
				t.Errorf("edited while off: deactivated_at = %v, want kept at %v", edited.DeactivatedAt, off.DeactivatedAt)
			}

			if on := update("Kajian Fiqih", true); on.DeactivatedAt != nil {
				t.Errorf("turned on: deactivated_at = %v, want nil", on.DeactivatedAt)
			}

			all, err := repo.FindAll(ctx)
			if err != nil || len(all) != 1 || all[0].CreatedByName != "Ahmad" {
				t.Errorf("find all = %+v, %v", all, err)
			}
		})
	}
}

func TestReorder(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			var ids []int64
			for _, text := range []string{"a", "b", "c"} {
				rt, err := repo.Create(ctx, domain.RunningText{Text: text})
				if err != nil {
					t.Fatalf("create %s: %v", text, err)
				}
				ids = append(ids, rt.ID)
			}

			if err := repo.Reorder(ctx, []int64{ids[2], ids[0], ids[1]}); err != nil {
				t.Fatalf("reorder: %v", err)
			}
			if got := texts(t, repo); !equal(got, []string{"c", "a", "b"}) {
				t.Errorf("texts = %v, want c a b", got)
			}

			// a new text still goes last
			if _, err := repo.Create(ctx, domain.RunningText{Text: "d"}); err != nil {
				t.Fatalf("create d: %v", err)
			}
			if got := texts(t, repo); !equal(got, []string{"c", "a", "b", "d"}) {
				t.Errorf("texts = %v, want c a b d", got)
			}
		})
	}
}

func TestSetting(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			if _, err := repo.GetSetting(ctx); !errors.Is(err, domain.ErrRunningTextSettingNotFound) {
				t.Errorf("never saved: err = %v, want ErrRunningTextSettingNotFound", err)
			}

			for _, speed := range []domain.ScrollSpeed{domain.ScrollSpeedFast, domain.ScrollSpeedSlow} {
				if err := repo.UpdateSetting(ctx, domain.RunningTextSetting{Speed: speed}, repo.userID); err != nil {
					t.Fatalf("save %s: %v", speed, err)
				}
				if got, err := repo.GetSetting(ctx); err != nil || got.Speed != speed {
					t.Errorf("after saving %s = %+v, %v", speed, got, err)
				}
			}
		})
	}
}
