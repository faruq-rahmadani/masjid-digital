package iqamahsettings

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
// the settings can point at
type testRepository struct {
	*IqamahDBRepository
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
			IqamahDBRepository: NewIqamahDBRepository(database.DatabaseClient{Querier: db}),
			userID:             userID,
		}
	}
	return repos
}

func TestSettings(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			if _, err := repo.GetSettings(ctx); !errors.Is(err, domain.ErrIqamahSettingsNotFound) {
				t.Fatalf("never saved: err = %v, want ErrIqamahSettingsNotFound", err)
			}

			settings := domain.DefaultIqamahSettings()
			settings.SholatMin.Jumat = 45
			settings.SholatDisplay = domain.SholatDisplayBlack
			settings.Alarm.Sound = domain.AlarmSoundBell
			for _, s := range []domain.IqamahSettings{domain.DefaultIqamahSettings(), settings} {
				if err := repo.UpdateSettings(ctx, s, repo.userID); err != nil {
					t.Fatalf("save: %v", err)
				}
				if got, err := repo.GetSettings(ctx); err != nil || !reflect.DeepEqual(got, s) {
					t.Errorf("got %+v, %v; want %+v", got, err, s)
				}
			}
		})
	}
}
