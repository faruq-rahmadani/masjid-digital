package mosqueprofiles

import (
	"bytes"
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

// testRepository is a repository on a freshly migrated database, with the ID
// of a user that settings can point at as their updater
type testRepository struct {
	*MosqueProfileDBRepository
	db     *pkgdb.DB
	userID int64
}

// newTestRepositories returns a repository on a freshly migrated database
// for every engine available
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
			MosqueProfileDBRepository: NewMosqueProfileDBRepository(database.DatabaseClient{Querier: db}),
			db:                        db,
			userID:                    userID,
		}
	}
	return repos
}

var profile = domain.MosqueProfile{
	Name:         "Masjid Al-Ikhlas",
	Address:      "Jl. Merdeka No. 1",
	ContactPhone: "+6281234567890",
	Latitude:     -6.2088,
	Longitude:    106.8456,
	Timezone:     domain.MosqueTimezoneWIB,
}

var logo = domain.MosqueLogo{ContentType: "image/png", Data: []byte("\x89PNG\r\n\x1a\n")}

// updatedBy reads who last saved the row under key
func (r testRepository) updatedBy(t *testing.T, key settingKey) *int64 {
	t.Helper()

	var by *int64
	if err := r.db.Get(context.Background(), &by, "SELECT updated_by FROM configurations WHERE key = ?", string(key)); err != nil {
		t.Fatalf("get updated_by of %s: %v", key, err)
	}
	return by
}

func TestGetNeverSaved(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			_, err := repo.Get(context.Background())
			if !errors.Is(err, domain.ErrMosqueProfileNotFound) {
				t.Errorf("err = %v, want ErrMosqueProfileNotFound", err)
			}
		})
	}
}

func TestUpdateInsertsThenReplaces(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			// first save by nobody, e.g. seeded
			first := profile
			first.Name = "Masjid Lama"
			if err := repo.Update(ctx, first, 0); err != nil {
				t.Fatalf("insert: %v", err)
			}
			if got, err := repo.Get(ctx); err != nil || got != first {
				t.Errorf("after insert = %+v, %v; want %+v", got, err, first)
			}
			if by := repo.updatedBy(t, keyProfile); by != nil {
				t.Errorf("updated_by after insert = %d, want NULL", *by)
			}

			if err := repo.Update(ctx, profile, repo.userID); err != nil {
				t.Fatalf("replace: %v", err)
			}
			if got, err := repo.Get(ctx); err != nil || got != profile {
				t.Errorf("after replace = %+v, %v; want %+v", got, err, profile)
			}
			if by := repo.updatedBy(t, keyProfile); by == nil || *by != repo.userID {
				t.Errorf("updated_by after replace = %v, want %d", by, repo.userID)
			}
		})
	}
}

func TestGetLogoNeverUploaded(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			_, err := repo.GetLogo(context.Background())
			if !errors.Is(err, domain.ErrMosqueLogoNotFound) {
				t.Errorf("err = %v, want ErrMosqueLogoNotFound", err)
			}
		})
	}
}

func TestUpdateLogo(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			if err := repo.UpdateLogo(ctx, domain.MosqueLogo{ContentType: "image/jpeg", Data: []byte("\xff\xd8\xff")}, repo.userID); err != nil {
				t.Fatalf("insert: %v", err)
			}
			if err := repo.UpdateLogo(ctx, logo, repo.userID); err != nil {
				t.Fatalf("replace: %v", err)
			}

			got, err := repo.GetLogo(ctx)
			if err != nil {
				t.Fatalf("get: %v", err)
			}
			if got.ContentType != logo.ContentType || !bytes.Equal(got.Data, logo.Data) {
				t.Errorf("logo = %+v, want %+v", got, logo)
			}
		})
	}
}

func TestProfileAndLogoAreApart(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			if err := repo.Update(ctx, profile, repo.userID); err != nil {
				t.Fatalf("save profile: %v", err)
			}
			if err := repo.UpdateLogo(ctx, logo, repo.userID); err != nil {
				t.Fatalf("save logo: %v", err)
			}
			if err := repo.DeleteLogo(ctx); err != nil {
				t.Fatalf("delete logo: %v", err)
			}

			if got, err := repo.Get(ctx); err != nil || got != profile {
				t.Errorf("profile = %+v, %v; the logo must not touch it", got, err)
			}
		})
	}
}

func TestDeleteLogo(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			if err := repo.UpdateLogo(ctx, logo, repo.userID); err != nil {
				t.Fatalf("save: %v", err)
			}
			if err := repo.DeleteLogo(ctx); err != nil {
				t.Fatalf("delete: %v", err)
			}
			if _, err := repo.GetLogo(ctx); !errors.Is(err, domain.ErrMosqueLogoNotFound) {
				t.Errorf("get after delete: err = %v, want ErrMosqueLogoNotFound", err)
			}

			// deleting again is fine
			if err := repo.DeleteLogo(ctx); err != nil {
				t.Errorf("delete missing: %v", err)
			}
		})
	}
}
