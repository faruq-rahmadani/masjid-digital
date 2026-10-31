package main

import (
	"context"
	"fmt"
	"log"

	"github.com/faruq-rahmadani/masjid-digital/internal/config"
	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/controller"
	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/cache/sessions"
	repodb "github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/mosqueprofiles"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/users"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/authentication"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/mosqueprofile"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/user"
	"github.com/faruq-rahmadani/masjid-digital/migrations"
	"github.com/faruq-rahmadani/masjid-digital/pkg/cache"
	"github.com/faruq-rahmadani/masjid-digital/pkg/database"
	"github.com/faruq-rahmadani/masjid-digital/pkg/hasher"
	"github.com/faruq-rahmadani/masjid-digital/pkg/token"
)

type repositories struct {
	users          usecase.UserRepositoryItf
	sessions       usecase.SessionRepositoryItf
	mosqueProfiles usecase.MosqueProfileRepositoryItf
}

type usecases struct {
	user          controller.UserUsecaseItf
	auth          *authentication.AuthenticationUseCase // used by both the controller and the auth middleware
	mosqueProfile controller.MosqueProfileUsecaseItf
}

// handlers is what the HTTP server serves
type handlers struct {
	controller *controller.Controller
	auth       *middleware.Auth
}

// newHandlers wires database and cache -> repositories -> usecases -> controller and middleware.
// Call cleanup on shutdown to release what was opened here.
func newHandlers(ctx context.Context, cfg *config.Schema) (h handlers, cleanup func(), err error) {
	db, err := initDatabase(ctx, cfg.App.Database)
	if err != nil {
		return handlers{}, nil, fmt.Errorf("error initializing database: %w", err)
	}

	c, err := cache.Open(ctx, cfg.App.Cache.ConnConfig())
	if err != nil {
		db.Close()
		return handlers{}, nil, fmt.Errorf("error initializing cache: %w", err)
	}
	log.Printf("cache connected: driver=%s", cfg.App.Cache.Driver)

	cleanup = func() {
		if err := c.Close(); err != nil {
			log.Printf("error closing cache: %v", err)
		}
		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}

	repos := newRepositories(db, c)
	ucs, err := newUsecases(cfg, repos)
	if err != nil {
		cleanup()
		return handlers{}, nil, err
	}

	return handlers{
		controller: controller.NewController(ucs.user, ucs.auth, ucs.mosqueProfile),
		auth:       middleware.NewAuth(ucs.auth),
	}, cleanup, nil
}

func newRepositories(db *database.DB, c cache.Cache) repositories {
	dbClient := repodb.DatabaseClient{Querier: db}

	return repositories{
		users:          users.NewUserDBRepository(dbClient),
		sessions:       sessions.NewSessionCacheRepository(c),
		mosqueProfiles: mosqueprofiles.NewMosqueProfileDBRepository(dbClient),
	}
}

func newUsecases(cfg *config.Schema, repos repositories) (usecases, error) {
	passwordHasher := hasher.NewArgon2(hasher.DefaultArgon2Params)

	tokenIssuer, err := token.NewJWT(cfg.App.Auth.JWTSecret, cfg.App.Auth.TokenTTL, cfg.App.Name)
	if err != nil {
		return usecases{}, fmt.Errorf("error configuring auth tokens (app.auth): %w", err)
	}

	return usecases{
		user:          newUserUsecase(repos, passwordHasher),
		auth:          authentication.New(repos.users, passwordHasher, tokenIssuer, repos.sessions),
		mosqueProfile: mosqueprofile.NewMosqueProfileUsecase(repos.mosqueProfiles),
	}, nil
}

// newUserUsecase builds the user usecase with its AfterUpdate and AfterPasswordReset hooks
func newUserUsecase(repos repositories, passwordHasher hasher.PasswordHasher) *user.UserUsecase {
	uc := user.NewUserUsecase(repos.users, passwordHasher)
	uc.AddAfterUpdate(
		user.NewRevokeSessionOnDisableHook(repos.sessions),    // a disabled user is logged out right away
		user.NewRevokeSessionOnRoleChangeHook(repos.sessions), // the token carries the role, so a new role needs a new login
	)
	uc.AddAfterPasswordReset(
		user.NewRevokeSessionOnPasswordResetHook(repos.sessions), // the old session must log in with the new password
	)
	return uc
}

// initDatabase applies pending migrations, then opens the connection pool used by the app
func initDatabase(ctx context.Context, cfg config.Database) (*database.DB, error) {
	if err := migrateUp(ctx, cfg); err != nil {
		return nil, fmt.Errorf("error migrating database: %w", err)
	}

	db, err := database.Open(ctx, cfg.ConnConfig())
	if err != nil {
		return nil, err
	}

	log.Printf("database connected: driver=%s", db.Driver())

	return db, nil
}

// migrateUp applies every pending migration embedded in the binary over its own
// connection, closed once done. It does nothing when auto_migrate is off.
func migrateUp(ctx context.Context, cfg config.Database) error {
	if !cfg.AutoMigrate {
		return nil
	}

	db, err := database.Open(ctx, cfg.MigrateConnConfig())
	if err != nil {
		return err
	}
	defer db.Close()

	m, err := db.NewMigrator(migrations.FS)
	if err != nil {
		return err
	}

	results, err := m.Up(ctx)
	if err != nil {
		return err
	}

	for _, r := range results {
		log.Printf("migration applied: %s (%s)", r.Source.Path, r.Duration)
	}

	version, err := m.Version(ctx)
	if err != nil {
		return err
	}
	log.Printf("database schema at version %d", version)

	return nil
}
