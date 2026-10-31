package database

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// schema is portable across both engines: IDs are generated in Go,
// money is an integer and time is passed in from Go
const schema = `
CREATE TABLE mosques (
	id         TEXT PRIMARY KEY,
	slug       TEXT NOT NULL UNIQUE,
	name       TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL
);
CREATE TABLE transactions (
	id        TEXT PRIMARY KEY,
	mosque_id TEXT NOT NULL REFERENCES mosques(id),
	amount    BIGINT NOT NULL,
	is_income BOOLEAN NOT NULL
);`

type mosque struct {
	ID        string    `db:"id"`
	Slug      string    `db:"slug"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

// openTestDBs returns a fresh database for every engine available
func openTestDBs(t *testing.T) map[Driver]*DB {
	t.Helper()
	ctx := context.Background()

	cfgs := map[Driver]Config{
		DriverSQLite: {Driver: DriverSQLite, DSN: filepath.Join(t.TempDir(), "test.db")},
	}
	if dsn := os.Getenv("TEST_POSTGRES_DSN"); dsn != "" {
		cfgs[DriverPostgres] = Config{Driver: DriverPostgres, DSN: dsn}
	}

	dbs := make(map[Driver]*DB, len(cfgs))
	for driver, cfg := range cfgs {
		db, err := Open(ctx, cfg)
		if err != nil {
			t.Fatalf("open %s: %v", driver, err)
		}
		t.Cleanup(func() {
			_, _ = db.Exec(ctx, "DROP TABLE IF EXISTS transactions")
			_, _ = db.Exec(ctx, "DROP TABLE IF EXISTS mosques")
			db.Close()
		})
		_, _ = db.Exec(ctx, "DROP TABLE IF EXISTS transactions")
		_, _ = db.Exec(ctx, "DROP TABLE IF EXISTS mosques")
		if _, err := db.Exec(ctx, schema); err != nil {
			t.Fatalf("create schema on %s: %v", driver, err)
		}
		dbs[driver] = db
	}
	return dbs
}

func insertMosque(ctx context.Context, q Querier, m mosque) error {
	_, err := q.Exec(ctx,
		"INSERT INTO mosques (id, slug, name, created_at) VALUES (?, ?, ?, ?)",
		m.ID, m.Slug, m.Name, m.CreatedAt,
	)
	return err
}

func TestQuerier(t *testing.T) {
	for driver, db := range openTestDBs(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()
			now := time.Now().UTC().Truncate(time.Millisecond)

			want := mosque{ID: "m1", Slug: "al-ikhlas", Name: "Masjid Al-Ikhlas", CreatedAt: now}
			if err := insertMosque(ctx, db, want); err != nil {
				t.Fatalf("insert: %v", err)
			}

			var got mosque
			if err := db.Get(ctx, &got, "SELECT * FROM mosques WHERE slug = ?", "al-ikhlas"); err != nil {
				t.Fatalf("get: %v", err)
			}
			if got.ID != want.ID || got.Name != want.Name || !got.CreatedAt.Equal(want.CreatedAt) {
				t.Errorf("get = %+v, want %+v", got, want)
			}

			if err := db.Get(ctx, &got, "SELECT * FROM mosques WHERE slug = ?", "missing"); !errors.Is(err, ErrNotFound) {
				t.Errorf("get missing: err = %v, want ErrNotFound", err)
			}

			dup := mosque{ID: "m2", Slug: "al-ikhlas", Name: "Duplicate", CreatedAt: now}
			if err := insertMosque(ctx, db, dup); !errors.Is(err, ErrUniqueViolation) {
				t.Errorf("insert duplicate slug: err = %v, want ErrUniqueViolation", err)
			}

			_, err := db.Exec(ctx,
				"INSERT INTO transactions (id, mosque_id, amount, is_income) VALUES (?, ?, ?, ?)",
				"t1", "no-such-mosque", 50000, true,
			)
			if !errors.Is(err, ErrForeignKeyViolation) {
				t.Errorf("insert orphan transaction: err = %v, want ErrForeignKeyViolation", err)
			}

			var all []mosque
			if err := db.Select(ctx, &all, "SELECT * FROM mosques ORDER BY id"); err != nil {
				t.Fatalf("select: %v", err)
			}
			if len(all) != 1 {
				t.Errorf("select returned %d rows, want 1", len(all))
			}
		})
	}
}

func TestWithTx(t *testing.T) {
	for driver, db := range openTestDBs(t) {
		t.Run(string(driver), func(t *testing.T) {
			ctx := context.Background()
			now := time.Now().UTC()
			errBoom := errors.New("boom")

			err := db.WithTx(ctx, func(ctx context.Context) error {
				if err := insertMosque(ctx, db, mosque{ID: "rb", Slug: "rolled-back", Name: "x", CreatedAt: now}); err != nil {
					return err
				}
				return errBoom
			})
			if !errors.Is(err, errBoom) {
				t.Fatalf("WithTx err = %v, want errBoom", err)
			}

			var count int
			if err := db.Get(ctx, &count, "SELECT COUNT(*) FROM mosques WHERE id = ?", "rb"); err != nil {
				t.Fatalf("count: %v", err)
			}
			if count != 0 {
				t.Errorf("rolled back row still present")
			}

			err = db.WithTx(ctx, func(ctx context.Context) error {
				if err := insertMosque(ctx, db, mosque{ID: "c1", Slug: "committed", Name: "x", CreatedAt: now}); err != nil {
					return err
				}
				// nested WithTx joins the outer transaction
				return db.WithTx(ctx, func(ctx context.Context) error {
					_, err := db.Exec(ctx,
						"INSERT INTO transactions (id, mosque_id, amount, is_income) VALUES (?, ?, ?, ?)",
						"t1", "c1", 150000, true,
					)
					return err
				})
			})
			if err != nil {
				t.Fatalf("WithTx commit: %v", err)
			}

			var balance int64
			if err := db.Get(ctx, &balance, "SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE mosque_id = ?", "c1"); err != nil {
				t.Fatalf("balance: %v", err)
			}
			if balance != 150000 {
				t.Errorf("balance = %d, want 150000", balance)
			}
		})
	}
}

func TestSQLiteDSN(t *testing.T) {
	got, err := sqliteDSN("masjid.db?_pragma=busy_timeout(1000)&_txlock=deferred")
	if err != nil {
		t.Fatal(err)
	}
	want := "file:masjid.db?_pragma=busy_timeout%281000%29&_pragma=journal_mode%28WAL%29&_pragma=foreign_keys%281%29&_time_format=sqlite&_txlock=deferred"
	if got != want {
		t.Errorf("sqliteDSN = %q\nwant        %q", got, want)
	}
}

func TestOpenUnsupportedDriver(t *testing.T) {
	if _, err := Open(context.Background(), Config{Driver: "mysql", DSN: "x"}); err == nil {
		t.Error("Open with unsupported driver: want error")
	}
}
