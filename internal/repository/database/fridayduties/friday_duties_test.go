package fridayduties

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
// the duties can point at
type testRepository struct {
	*FridayDutyDBRepository
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
			FridayDutyDBRepository: NewFridayDutyDBRepository(database.DatabaseClient{Querier: db}),
			userID:                 userID,
		}
	}
	return repos
}

func TestFridayDuties(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			first, err := repo.Save(ctx, domain.FridayDuty{Date: "2026-11-06", Khatib: "Ustadz Abdullah", UpdatedBy: repo.userID})
			if err != nil {
				t.Fatalf("insert: %v", err)
			}
			if first.Khatib != "Ustadz Abdullah" || first.CreatedAt.IsZero() || first.UpdatedBy != repo.userID {
				t.Errorf("inserted = %+v", first)
			}

			// same date replaces the names but keeps created_at
			second, err := repo.Save(ctx, domain.FridayDuty{Date: "2026-11-06", Imam: "Ustadz Hasan", Theme: "Amanah"})
			if err != nil {
				t.Fatalf("replace: %v", err)
			}
			if second.Khatib != "" || second.Imam != "Ustadz Hasan" || second.Theme != "Amanah" || !second.CreatedAt.Equal(first.CreatedAt) {
				t.Errorf("replaced = %+v, first = %+v", second, first)
			}

			if _, err := repo.Save(ctx, domain.FridayDuty{Date: "2026-11-20", Khatib: "Ustadz Yusuf"}); err != nil {
				t.Fatalf("second friday: %v", err)
			}
			got, err := repo.FindBetween(ctx, "2026-11-01", "2026-11-13")
			if err != nil || len(got) != 1 || got[0].Date != "2026-11-06" {
				t.Errorf("between = %+v, %v; want only 2026-11-06", got, err)
			}

			if err := repo.Delete(ctx, "2026-11-06"); err != nil {
				t.Fatalf("delete: %v", err)
			}
			if err := repo.Delete(ctx, "2026-11-06"); !errors.Is(err, domain.ErrFridayDutyNotFound) {
				t.Errorf("delete again: err = %v, want ErrFridayDutyNotFound", err)
			}
		})
	}
}
