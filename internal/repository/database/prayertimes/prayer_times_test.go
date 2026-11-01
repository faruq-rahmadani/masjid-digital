package prayertimes

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	"github.com/faruq-rahmadani/masjid-digital/migrations"
	pkgdb "github.com/faruq-rahmadani/masjid-digital/pkg/database"
)

// testRepository is a repository on a freshly migrated database, with the ID of a user that
// settings and overrides can point at
type testRepository struct {
	*PrayerTimeDBRepository
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
			PrayerTimeDBRepository: NewPrayerTimeDBRepository(database.DatabaseClient{Querier: db}),
			db:                     db,
			userID:                 userID,
		}
	}
	return repos
}

var settings = domain.PrayerSettings{
	Corrections:     domain.PrayerCorrections{Subuh: 2, Dzuhur: 2, Ashar: 2, Maghrib: 2, Isya: 2},
	ImsakOffsetMin:  10,
	WeeklyOverrides: []domain.WeeklyPrayerOverride{{Weekday: "friday", Prayer: domain.PrayerDzuhur, Time: "12:00", Remarks: "Sholat Jumat"}},
	HijriOffsetDays: -1,
	HijriOffsetFrom: "2027-02-08",
}

func TestGetSettingsNeverSaved(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			if _, err := repo.GetSettings(context.Background()); !errors.Is(err, domain.ErrPrayerSettingsNotFound) {
				t.Errorf("err = %v, want ErrPrayerSettingsNotFound", err)
			}
		})
	}
}

func TestUpdateSettingsInsertsThenReplaces(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			if err := repo.UpdateSettings(ctx, domain.DefaultPrayerSettings(), 0); err != nil {
				t.Fatalf("insert: %v", err)
			}
			if got, err := repo.GetSettings(ctx); err != nil || !reflect.DeepEqual(got, domain.DefaultPrayerSettings()) {
				t.Errorf("after insert = %+v, %v", got, err)
			}

			if err := repo.UpdateSettings(ctx, settings, repo.userID); err != nil {
				t.Fatalf("replace: %v", err)
			}
			if got, err := repo.GetSettings(ctx); err != nil || !reflect.DeepEqual(got, settings) {
				t.Errorf("after replace = %+v, %v; want %+v", got, err, settings)
			}

			var by *int64
			if err := repo.db.Get(ctx, &by, "SELECT updated_by FROM configurations WHERE key = ?", settingsKey); err != nil || by == nil || *by != repo.userID {
				t.Errorf("updated_by = %v, %v; want %d", by, err, repo.userID)
			}
		})
	}
}

func TestOverrides(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			create := func(date string, prayer domain.Prayer, at string) domain.PrayerTimeOverride {
				t.Helper()
				o, err := repo.CreateOverride(ctx, domain.PrayerTimeOverride{Date: date, Prayer: prayer, Time: at, Remarks: "Jumat", CreatedBy: repo.userID})
				if err != nil {
					t.Fatalf("create %s %s: %v", date, prayer, err)
				}
				return o
			}
			isya := create("2026-11-06", domain.PrayerIsya, "19:05")
			dzuhur := create("2026-11-06", domain.PrayerDzuhur, "11:58")
			create("2026-12-01", domain.PrayerSubuh, "04:00")

			if dzuhur.ID == 0 || dzuhur.CreatedAt.IsZero() {
				t.Errorf("created = %+v, want id and timestamps", dzuhur)
			}

			if _, err := repo.CreateOverride(ctx, domain.PrayerTimeOverride{Date: "2026-11-06", Prayer: domain.PrayerDzuhur, Time: "12:00"}); !errors.Is(err, domain.ErrPrayerTimeOverrideExists) {
				t.Errorf("duplicate: err = %v, want ErrPrayerTimeOverrideExists", err)
			}

			got, err := repo.FindOverridesBetween(ctx, "2026-11-01", "2026-11-30")
			if err != nil {
				t.Fatalf("find: %v", err)
			}
			if len(got) != 2 || got[0].ID != dzuhur.ID || got[1].ID != isya.ID {
				t.Errorf("found = %+v, want dzuhur then isya of 2026-11-06", got)
			}
			if got[0].Time != "11:58" || got[0].Remarks != "Jumat" || got[0].CreatedBy != repo.userID {
				t.Errorf("found[0] = %+v", got[0])
			}

			if err := repo.DeleteOverride(ctx, dzuhur.ID); err != nil {
				t.Fatalf("delete: %v", err)
			}
			if err := repo.DeleteOverride(ctx, dzuhur.ID); !errors.Is(err, domain.ErrPrayerTimeOverrideNotFound) {
				t.Errorf("delete again: err = %v, want ErrPrayerTimeOverrideNotFound", err)
			}
		})
	}
}
