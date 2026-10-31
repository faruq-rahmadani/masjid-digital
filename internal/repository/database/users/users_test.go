package users

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

// newTestRepositories returns a repository on a freshly migrated database
// for every engine available
func newTestRepositories(t *testing.T) map[pkgdb.Driver]*UserDBRepository {
	t.Helper()
	ctx := context.Background()

	cfgs := map[pkgdb.Driver]pkgdb.Config{
		pkgdb.DriverSQLite: {Driver: pkgdb.DriverSQLite, DSN: filepath.Join(t.TempDir(), "test.db")},
	}
	if dsn := os.Getenv("TEST_POSTGRES_DSN"); dsn != "" {
		cfgs[pkgdb.DriverPostgres] = pkgdb.Config{Driver: pkgdb.DriverPostgres, DSN: dsn}
	}

	repos := make(map[pkgdb.Driver]*UserDBRepository, len(cfgs))
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

		repos[driver] = NewUserDBRepository(database.DatabaseClient{Querier: db})
	}
	return repos
}

func newUser(phone, email string) domain.User {
	return domain.User{
		Name:         "Ahmad",
		Phone:        phone,
		Email:        email,
		PasswordHash: "hash",
		Role:         domain.UserRoleTakmir,
		Status:       domain.UserStatusActive,
	}
}

func TestCreate(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			user := newUser("+6281200000001", "ahmad@example.com")
			before := time.Now()
			id, err := repo.Create(ctx, user)
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			if id <= 0 {
				t.Errorf("id = %d, want a generated positive id", id)
			}

			var got struct {
				Phone     string    `db:"phone"`
				Email     *string   `db:"email"`
				Role      string    `db:"role"`
				CreatedAt time.Time `db:"created_at"`
			}
			if err := repo.client.Get(ctx, &got, "SELECT phone, email, role, created_at FROM users WHERE id = ?", id); err != nil {
				t.Fatalf("read back: %v", err)
			}
			if got.Phone != user.Phone || got.Email == nil || *got.Email != user.Email || got.Role != string(user.Role) {
				t.Errorf("stored user = %+v, want phone=%s email=%s role=%s", got, user.Phone, user.Email, user.Role)
			}
			if got.CreatedAt.Before(before.Truncate(time.Millisecond)) || got.CreatedAt.After(time.Now()) {
				t.Errorf("created_at = %v, want the time of Create", got.CreatedAt)
			}
			if _, offset := got.CreatedAt.Zone(); offset != 0 {
				t.Errorf("created_at stored with offset %ds, want UTC", offset)
			}

			nextID, err := repo.Create(ctx, newUser("+6281200000002", ""))
			if err != nil {
				t.Fatalf("create second user: %v", err)
			}
			if nextID <= id {
				t.Errorf("second id = %d, want greater than %d", nextID, id)
			}
		})
	}
}

func TestCreateDuplicate(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			if _, err := repo.Create(ctx, newUser("+6281200000001", "ahmad@example.com")); err != nil {
				t.Fatalf("create: %v", err)
			}

			_, err := repo.Create(ctx, newUser("+6281200000001", ""))
			if !errors.Is(err, domain.ErrUserAlreadyExists) {
				t.Errorf("duplicate phone: err = %v, want ErrUserAlreadyExists", err)
			}

			_, err = repo.Create(ctx, newUser("+6281200000003", "ahmad@example.com"))
			if !errors.Is(err, domain.ErrUserAlreadyExists) {
				t.Errorf("duplicate email: err = %v, want ErrUserAlreadyExists", err)
			}
		})
	}
}

func TestCreateWithoutEmail(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			// empty email is stored as NULL, so several users may omit it
			for i, phone := range []string{"+6281200000001", "+6281200000002"} {
				if _, err := repo.Create(ctx, newUser(phone, "")); err != nil {
					t.Fatalf("create user %d without email: %v", i, err)
				}
			}
		})
	}
}

func TestGetByPhoneAndEmail(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			user := newUser("+6281200000001", "ahmad@example.com")
			id, err := repo.Create(ctx, user)
			if err != nil {
				t.Fatalf("create: %v", err)
			}

			byPhone, err := repo.GetByPhone(ctx, "+6281200000001")
			if err != nil {
				t.Fatalf("get by phone: %v", err)
			}
			byEmail, err := repo.GetByEmail(ctx, "ahmad@example.com")
			if err != nil {
				t.Fatalf("get by email: %v", err)
			}

			for _, got := range []domain.User{byPhone, byEmail} {
				if got.ID != id || got.Name != user.Name || got.Phone != user.Phone || got.Email != user.Email ||
					got.PasswordHash != user.PasswordHash || got.Role != user.Role || got.Status != user.Status {
					t.Errorf("got %+v, want the created user %+v with id %d", got, user, id)
				}
				if got.CreatedAt.IsZero() || got.LastLoginAt != nil {
					t.Errorf("created_at=%v last_login_at=%v, want a set created_at and no last login", got.CreatedAt, got.LastLoginAt)
				}
			}

			if _, err := repo.GetByPhone(ctx, "+6281200000009"); !errors.Is(err, domain.ErrUserNotFound) {
				t.Errorf("unknown phone: err = %v, want ErrUserNotFound", err)
			}
			if _, err := repo.GetByEmail(ctx, "budi@example.com"); !errors.Is(err, domain.ErrUserNotFound) {
				t.Errorf("unknown email: err = %v, want ErrUserNotFound", err)
			}
		})
	}
}

func TestUpdateLastLoginAt(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			id, err := repo.Create(ctx, newUser("+6281200000001", ""))
			if err != nil {
				t.Fatalf("create: %v", err)
			}

			wib := time.FixedZone("WIB", 7*60*60)
			at := time.Date(2026, 10, 31, 10, 0, 0, 0, wib)
			if err := repo.UpdateLastLoginAt(ctx, id, at); err != nil {
				t.Fatalf("update last login: %v", err)
			}

			got, err := repo.GetByPhone(ctx, "+6281200000001")
			if err != nil {
				t.Fatalf("get: %v", err)
			}
			if got.LastLoginAt == nil || !got.LastLoginAt.Equal(at) {
				t.Errorf("last_login_at = %v, want %v", got.LastLoginAt, at)
			}
			if _, offset := got.LastLoginAt.Zone(); offset != 0 {
				t.Errorf("last_login_at stored with offset %ds, want UTC", offset)
			}

			if err := repo.UpdateLastLoginAt(ctx, id+100, at); !errors.Is(err, domain.ErrUserNotFound) {
				t.Errorf("unknown id: err = %v, want ErrUserNotFound", err)
			}
		})
	}
}

func TestCreatedBy(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			// the first user has no creator: 0 must be stored as NULL, not as a dangling id
			adminID, err := repo.Create(ctx, newUser("+6281200000001", ""))
			if err != nil {
				t.Fatalf("create first user: %v", err)
			}

			var raw *int64
			if err := repo.client.Get(ctx, &raw, "SELECT created_by FROM users WHERE id = ?", adminID); err != nil {
				t.Fatalf("read created_by: %v", err)
			}
			if raw != nil {
				t.Errorf("created_by of the first user = %d, want NULL", *raw)
			}

			member := newUser("+6281200000002", "")
			member.CreatedBy = adminID
			if _, err := repo.Create(ctx, member); err != nil {
				t.Fatalf("create user with creator: %v", err)
			}

			admin, err := repo.GetByPhone(ctx, "+6281200000001")
			if err != nil {
				t.Fatalf("get first user: %v", err)
			}
			got, err := repo.GetByPhone(ctx, "+6281200000002")
			if err != nil {
				t.Fatalf("get created user: %v", err)
			}
			if admin.CreatedBy != 0 || got.CreatedBy != adminID {
				t.Errorf("created_by = %d and %d, want 0 and %d", admin.CreatedBy, got.CreatedBy, adminID)
			}

			orphan := newUser("+6281200000003", "")
			orphan.CreatedBy = adminID + 100
			if _, err := repo.Create(ctx, orphan); err == nil {
				t.Error("user created by a missing user was accepted")
			}
		})
	}
}

func TestFindAllAndFindByID(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			empty, err := repo.FindAll(ctx)
			if err != nil || empty == nil || len(empty) != 0 {
				t.Errorf("find all on empty table = %v, %v; want an empty, non-nil slice", empty, err)
			}

			firstID, _ := repo.Create(ctx, newUser("+6281200000001", "ahmad@example.com"))
			secondID, _ := repo.Create(ctx, newUser("+6281200000002", ""))

			all, err := repo.FindAll(ctx)
			if err != nil {
				t.Fatalf("find all: %v", err)
			}
			if len(all) != 2 || all[0].ID != firstID || all[1].ID != secondID {
				t.Errorf("find all = %+v, want users %d and %d in that order", all, firstID, secondID)
			}
			if all[0].Email != "ahmad@example.com" || all[1].Email != "" {
				t.Errorf("emails = %q, %q", all[0].Email, all[1].Email)
			}

			got, err := repo.FindByID(ctx, secondID)
			if err != nil || got.Phone != "+6281200000002" {
				t.Errorf("find by id = %+v, %v; want the second user", got, err)
			}

			if _, err := repo.FindByID(ctx, secondID+100); !errors.Is(err, domain.ErrUserNotFound) {
				t.Errorf("unknown id: err = %v, want ErrUserNotFound", err)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			id, _ := repo.Create(ctx, newUser("+6281200000001", "ahmad@example.com"))
			_, _ = repo.Create(ctx, newUser("+6281200000002", "budi@example.com"))

			before, _ := repo.FindByID(ctx, id)

			changed := before
			changed.Name = "Ahmad Fauzi"
			changed.Phone = "+6281200000009"
			changed.Email = ""
			changed.Role = domain.UserRoleSuperAdmin
			changed.Status = domain.UserStatusDisabled
			changed.PasswordHash = "must-not-change"
			if err := repo.Update(ctx, changed); err != nil {
				t.Fatalf("update: %v", err)
			}

			got, _ := repo.FindByID(ctx, id)
			if got.Name != "Ahmad Fauzi" || got.Phone != "+6281200000009" || got.Email != "" ||
				got.Role != domain.UserRoleSuperAdmin || got.Status != domain.UserStatusDisabled {
				t.Errorf("after update = %+v", got)
			}
			if got.PasswordHash != before.PasswordHash || !got.CreatedAt.Equal(before.CreatedAt) {
				t.Error("update changed the password hash or created_at")
			}
			if got.UpdatedAt.Before(before.UpdatedAt) {
				t.Errorf("updated_at went from %v to %v", before.UpdatedAt, got.UpdatedAt)
			}

			taken := got
			taken.Email = "budi@example.com"
			if err := repo.Update(ctx, taken); !errors.Is(err, domain.ErrUserAlreadyExists) {
				t.Errorf("email taken: err = %v, want ErrUserAlreadyExists", err)
			}

			missing := got
			missing.ID = id + 100
			if err := repo.Update(ctx, missing); !errors.Is(err, domain.ErrUserNotFound) {
				t.Errorf("unknown id: err = %v, want ErrUserNotFound", err)
			}
		})
	}
}

func TestUpdatePasswordHash(t *testing.T) {
	for driver, repo := range newTestRepositories(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()

			id, _ := repo.Create(ctx, newUser("+6281200000001", ""))
			before, _ := repo.FindByID(ctx, id)

			if err := repo.UpdatePasswordHash(ctx, id, "new-hash"); err != nil {
				t.Fatalf("update password: %v", err)
			}

			got, _ := repo.FindByID(ctx, id)
			if got.PasswordHash != "new-hash" {
				t.Errorf("password hash = %q, want new-hash", got.PasswordHash)
			}
			if got.Name != before.Name || got.Phone != before.Phone || got.UpdatedAt.Before(before.UpdatedAt) {
				t.Errorf("other fields changed: before %+v, after %+v", before, got)
			}

			if err := repo.UpdatePasswordHash(ctx, id+100, "x"); !errors.Is(err, domain.ErrUserNotFound) {
				t.Errorf("unknown id: err = %v, want ErrUserNotFound", err)
			}
		})
	}
}
