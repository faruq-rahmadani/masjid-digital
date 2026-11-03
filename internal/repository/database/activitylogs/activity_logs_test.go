package activitylogs

import (
	"context"
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
// logs can point at
type testRepository struct {
	*repo
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
			"Hasan Basri", "+6281200000001", "hash", "takmir", now, now,
		); err != nil {
			t.Fatalf("insert user on %s: %v", driver, err)
		}

		repos[driver] = testRepository{repo: New(database.DatabaseClient{Querier: db}), userID: userID}
	}
	return repos
}

func TestCreateAndFindRecent(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()
			base := time.Date(2026, 11, 3, 8, 0, 0, 0, time.UTC)

			logs := []domain.ActivityLog{
				{ActorID: repo.userID, Action: domain.ActivityIqamahSettingsUpdated, Summary: "mengubah pengaturan iqamah", CreatedAt: base},
				{ActorID: repo.userID, Action: domain.ActivitySlideCreated, EntityID: 4, Summary: "menambah slide Kajian Ahad", CreatedAt: base.Add(time.Minute)},
				{Action: domain.ActivitySlideDeleted, EntityID: 4, Summary: "menghapus slide", CreatedAt: base.Add(2 * time.Minute)},
			}
			for _, l := range logs {
				if err := repo.Create(ctx, l); err != nil {
					t.Fatalf("create: %v", err)
				}
			}

			got, err := repo.FindRecent(ctx, 2)
			if err != nil {
				t.Fatalf("find recent: %v", err)
			}
			if len(got) != 2 {
				t.Fatalf("got %d logs, want 2", len(got))
			}

			// newest first; no actor reads back as 0 and ""
			if got[0].Action != domain.ActivitySlideDeleted || got[0].ActorID != 0 || got[0].ActorName != "" || got[0].EntityID != 4 {
				t.Errorf("newest = %+v", got[0])
			}
			second := got[1]
			if second.ID == 0 || second.ActorID != repo.userID || second.ActorName != "Hasan Basri" ||
				second.Summary != "menambah slide Kajian Ahad" || !second.CreatedAt.Equal(base.Add(time.Minute)) {
				t.Errorf("second = %+v", second)
			}
		})
	}
}

func TestFindRecentSettingHasNoEntity(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()
			if err := repo.Create(ctx, domain.ActivityLog{ActorID: repo.userID, Action: domain.ActivityIqamahSettingsUpdated, Summary: "mengubah pengaturan iqamah", CreatedAt: time.Now()}); err != nil {
				t.Fatalf("create: %v", err)
			}

			got, err := repo.FindRecent(ctx, 10)
			if err != nil || len(got) != 1 || got[0].EntityID != 0 {
				t.Errorf("got %+v, %v; want one log without entity", got, err)
			}
		})
	}
}
