package main

import (
	"context"
	"fmt"
	"log"

	"github.com/faruq-rahmadani/masjid-digital/internal/config"
	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/controller"
	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/cache/screenpresence"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/cache/sessions"
	repodb "github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/accounts"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/activitylogs"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/bookclosings"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/categories"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/financesettings"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/fridayduties"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/iqamahsettings"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/mosqueprofiles"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/pockets"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/prayertimes"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/runningtexts"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/screens"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/slides"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/transactions"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/users"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/realtime/activitylogbus"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/realtime/screenhub"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/activitylog"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/authentication"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/finance"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/fridayduty"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/iqamah"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/mosqueprofile"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/prayertime"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/runningtext"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/screen"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/slide"
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
	prayerTimes    usecase.PrayerTimeRepositoryItf
	runningTexts   usecase.RunningTextRepositoryItf
	iqamah         usecase.IqamahRepositoryItf
	fridayDuties   usecase.FridayDutyRepositoryItf
	slides         usecase.SlideRepositoryItf
	screens        usecase.ScreenRepositoryItf
	screenPresence usecase.ScreenPresenceRepositoryItf
	screenHub      usecase.ScreenHubItf
	activityLogs   usecase.ActivityLogRepositoryItf
	finance        finance.Repositories
}

type usecases struct {
	user          controller.UserUsecaseItf
	auth          *authentication.AuthenticationUseCase // used by both the controller and the auth middleware
	mosqueProfile controller.MosqueProfileUsecaseItf
	prayerTime    controller.PrayerTimeUsecaseItf
	runningText   controller.RunningTextUsecaseItf
	iqamah        controller.IqamahUsecaseItf
	fridayDuty    controller.FridayDutyUsecaseItf
	slide         controller.SlideUsecaseItf
	screen        controller.ScreenUsecaseItf
	activityLog   *activitylog.UseCase // used by both the controller and the activity log bus
	finance       controller.FinanceUsecaseItf
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

	repos := newRepositories(db, c)

	// every usecase publishes its changes here; the activity log usecase saves them
	activityLog := activitylog.NewActivityLogUsecase(repos.activityLogs)
	activityLogBus := activitylogbus.New(activityLog.Save)

	cleanup = func() {
		activityLogBus.Close() // save the queued logs while the database is still open
		if err := c.Close(); err != nil {
			log.Printf("error closing cache: %v", err)
		}
		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}

	ucs, err := newUsecases(cfg, repos, activityLogBus, activityLog)
	if err != nil {
		cleanup()
		return handlers{}, nil, err
	}

	return handlers{
		controller: controller.NewController(ucs.user, ucs.auth, ucs.mosqueProfile, ucs.prayerTime, ucs.runningText, ucs.iqamah, ucs.fridayDuty, ucs.slide, ucs.screen, ucs.activityLog, ucs.finance),
		auth:       middleware.NewAuth(ucs.auth),
	}, cleanup, nil
}

func newRepositories(db *database.DB, c cache.Cache) repositories {
	dbClient := repodb.DatabaseClient{Querier: db}

	return repositories{
		users:          users.NewUserDBRepository(dbClient),
		sessions:       sessions.NewSessionCacheRepository(c),
		mosqueProfiles: mosqueprofiles.NewMosqueProfileDBRepository(dbClient),
		prayerTimes:    prayertimes.NewPrayerTimeDBRepository(dbClient),
		runningTexts:   runningtexts.New(dbClient),
		iqamah:         iqamahsettings.NewIqamahDBRepository(dbClient),
		fridayDuties:   fridayduties.NewFridayDutyDBRepository(dbClient),
		slides:         slides.New(dbClient),
		screens:        screens.New(dbClient),
		screenPresence: screenpresence.New(c),
		screenHub:      screenhub.New(),
		activityLogs:   activitylogs.New(dbClient),
		finance: finance.Repositories{
			Transactor:     db,
			Accounts:       accounts.New(dbClient),
			Pockets:        pockets.New(dbClient),
			Categories:     categories.New(dbClient),
			Transactions:   transactions.New(dbClient),
			BookClosings:   bookclosings.New(dbClient),
			MosqueProfiles: mosqueprofiles.NewMosqueProfileDBRepository(dbClient),
			PrayerTimes:    prayertimes.NewPrayerTimeDBRepository(dbClient),
			Settings:       financesettings.New(dbClient),
		},
	}
}

func newUsecases(cfg *config.Schema, repos repositories, activityLogBus usecase.ActivityLogPublisherItf, activityLog *activitylog.UseCase) (usecases, error) {
	passwordHasher := hasher.NewArgon2(hasher.DefaultArgon2Params)

	tokenIssuer, err := token.NewJWT(cfg.App.Auth.JWTSecret, cfg.App.Auth.TokenTTL, cfg.App.Name)
	if err != nil {
		return usecases{}, fmt.Errorf("error configuring auth tokens (app.auth): %w", err)
	}

	return usecases{
		user:          newUserUsecase(repos, passwordHasher, activityLogBus),
		auth:          authentication.New(repos.users, passwordHasher, tokenIssuer, repos.sessions),
		mosqueProfile: mosqueprofile.NewMosqueProfileUsecase(repos.mosqueProfiles, activityLogBus),
		prayerTime:    prayertime.NewPrayerTimeUsecase(repos.prayerTimes, activityLogBus),
		runningText:   runningtext.New(repos.runningTexts, activityLogBus),
		iqamah:        iqamah.NewIqamahUsecase(repos.iqamah, activityLogBus),
		fridayDuty:    fridayduty.NewFridayDutyUsecase(repos.fridayDuties, activityLogBus),
		slide:         slide.New(repos.slides, activityLogBus),
		screen:        screen.New(repos.screens, repos.screenPresence, repos.screenHub, repos.mosqueProfiles, activityLogBus),
		activityLog:   activityLog,
		finance:       finance.New(repos.finance, activityLogBus),
	}, nil
}

// newUserUsecase builds the user usecase with its AfterUpdate and AfterPasswordReset hooks
func newUserUsecase(repos repositories, passwordHasher hasher.PasswordHasher, activityLogBus usecase.ActivityLogPublisherItf) *user.UserUsecase {
	uc := user.NewUserUsecase(repos.users, passwordHasher, activityLogBus)
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
