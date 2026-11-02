package slides

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	"github.com/faruq-rahmadani/masjid-digital/migrations"
	pkgdb "github.com/faruq-rahmadani/masjid-digital/pkg/database"
)

// testRepository is a repository on a freshly migrated database, with the ID of a user that
// slides can point at
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

var png = domain.SlideImage{ContentType: "image/png", Data: []byte{0x89, 'P', 'N', 'G', 0x00, 0xff}}

// titles returns the slide titles in display order
func titles(t *testing.T, repo testRepository) []string {
	t.Helper()
	all, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("find all: %v", err)
	}
	got := make([]string, 0, len(all))
	for _, s := range all {
		got = append(got, s.Title)
	}
	return got
}

func TestCreateAppendsActiveWithImage(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			first, err := repo.Create(ctx, domain.Slide{Title: "Kajian", Caption: "Ahad pagi", DurationSeconds: 10, ShowCaption: true, CreatedBy: repo.userID}, png)
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			if first.ID == 0 || !first.Active || first.Position != 0 || first.Caption != "Ahad pagi" || first.DurationSeconds != 10 || !first.ShowCaption || first.CreatedByName != "Ahmad" || first.CreatedAt.IsZero() {
				t.Errorf("created = %+v", first)
			}

			second, err := repo.Create(ctx, domain.Slide{Title: "Infaq", DurationSeconds: 15}, domain.SlideImage{ContentType: "image/jpeg", Data: []byte{0xff, 0xd8}})
			if err != nil {
				t.Fatalf("create second: %v", err)
			}
			if second.Position != 1 || second.ShowCaption || second.CreatedBy != 0 || second.CreatedByName != "" {
				t.Errorf("second = %+v, want position 1 and no creator", second)
			}

			if got := titles(t, repo); !slices.Equal(got, []string{"Kajian", "Infaq"}) {
				t.Errorf("titles = %v", got)
			}

			image, err := repo.GetImage(ctx, first.ID)
			if err != nil || image.ContentType != png.ContentType || !bytes.Equal(image.Data, png.Data) {
				t.Errorf("image = %+v, %v; want %+v", image, err, png)
			}
			if _, err := repo.GetImage(ctx, second.ID+100); !errors.Is(err, domain.ErrSlideNotFound) {
				t.Errorf("image of unknown slide: err = %v, want ErrSlideNotFound", err)
			}
		})
	}
}

func TestUpdateAndDelete(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			created, err := repo.Create(ctx, domain.Slide{Title: "Kajian", DurationSeconds: 10}, png)
			if err != nil {
				t.Fatalf("create: %v", err)
			}

			updated, err := repo.Update(ctx, domain.Slide{ID: created.ID, Title: "Kajian Fiqih", Caption: "Kamis", DurationSeconds: 30, ShowCaption: true, Active: false})
			if err != nil {
				t.Fatalf("update: %v", err)
			}
			if updated.Title != "Kajian Fiqih" || updated.Caption != "Kamis" || updated.DurationSeconds != 30 || !updated.ShowCaption || updated.Active || updated.Position != created.Position || !updated.CreatedAt.Equal(created.CreatedAt) {
				t.Errorf("updated = %+v", updated)
			}

			if _, err := repo.Update(ctx, domain.Slide{ID: created.ID + 100, Title: "x", DurationSeconds: 10}); !errors.Is(err, domain.ErrSlideNotFound) {
				t.Errorf("update unknown: err = %v, want ErrSlideNotFound", err)
			}

			if err := repo.Delete(ctx, created.ID); err != nil {
				t.Fatalf("delete: %v", err)
			}
			if err := repo.Delete(ctx, created.ID); !errors.Is(err, domain.ErrSlideNotFound) {
				t.Errorf("delete again: err = %v, want ErrSlideNotFound", err)
			}
			if _, err := repo.GetImage(ctx, created.ID); !errors.Is(err, domain.ErrSlideNotFound) {
				t.Errorf("image after delete: err = %v, want ErrSlideNotFound", err)
			}
		})
	}
}

func TestReorder(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			var ids []int64
			for _, title := range []string{"A", "B", "C"} {
				s, err := repo.Create(ctx, domain.Slide{Title: title, DurationSeconds: 10}, png)
				if err != nil {
					t.Fatalf("create %s: %v", title, err)
				}
				ids = append(ids, s.ID)
			}

			if err := repo.Reorder(ctx, []int64{ids[2], ids[0], ids[1]}); err != nil {
				t.Fatalf("reorder: %v", err)
			}
			if got := titles(t, repo); !slices.Equal(got, []string{"C", "A", "B"}) {
				t.Errorf("titles = %v, want [C A B]", got)
			}

			// a slide added after a reorder still goes last
			if _, err := repo.Create(ctx, domain.Slide{Title: "D", DurationSeconds: 10}, png); err != nil {
				t.Fatalf("create D: %v", err)
			}
			if got := titles(t, repo); !slices.Equal(got, []string{"C", "A", "B", "D"}) {
				t.Errorf("titles = %v, want [C A B D]", got)
			}
		})
	}
}
