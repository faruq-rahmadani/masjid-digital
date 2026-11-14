package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/faruq-rahmadani/masjid-digital/internal/config"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	repodb "github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/activitylogs"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/users"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/realtime/activitylogbus"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/activitylog"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase/user"
	"github.com/faruq-rahmadani/masjid-digital/pkg/database"
	"github.com/faruq-rahmadani/masjid-digital/pkg/env"
	"github.com/faruq-rahmadani/masjid-digital/pkg/hasher"
)

// passwordEnv lets scripts pass the password without a prompt
const passwordEnv = "ADMIN_PASSWORD"

const usage = `usage: create-admin -name <name> -phone <phone> [-email <email>]

Creates a super_admin, e.g. the first user of a new deployment, who can then
add everyone else from the admin panel. The password is read from $ADMIN_PASSWORD
or, when unset, asked on stdin.

The database must be migrated: start the app once (auto_migrate) or run migrate up.`

func main() {
	name := flag.String("name", "", "full name (required)")
	phone := flag.String("phone", "", "phone number used to log in (required)")
	email := flag.String("email", "", "email (optional)")
	flag.Usage = func() { fmt.Fprintln(os.Stderr, usage) }
	flag.Parse()

	if *name == "" || *phone == "" {
		flag.Usage()
		os.Exit(2)
	}

	// Init ENV, empty falls back to development
	if err := env.Init(os.Getenv("ENV")); err != nil {
		log.Fatalf("error initializing env: %v", err)
	}

	cfg, err := config.LoadAll(env.GetEnv())
	if err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}

	password, err := readPassword()
	if err != nil {
		log.Fatalf("error reading password: %v", err)
	}

	ctx := context.Background()

	db, err := database.Open(ctx, cfg.App.Database.ConnConfig())
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()

	dbClient := repodb.DatabaseClient{Querier: db}
	activityLogBus := activitylogbus.New(activitylog.NewActivityLogUsecase(activitylogs.New(dbClient)).Save)
	uc := user.NewUserUsecase(users.NewUserDBRepository(dbClient), hasher.NewArgon2(hasher.DefaultArgon2Params), activityLogBus)

	// CreatedBy stays 0: nobody is logged in, so created_by and the log's actor are NULL
	err = uc.Create(ctx, domain.CreateUserRequest{
		Name:     *name,
		Phone:    *phone,
		Email:    *email,
		Password: password,
		Role:     domain.UserRoleSuperAdmin,
	})
	activityLogBus.Close() // save the queued log before the database closes
	if err != nil {
		log.Fatalf("error creating super admin: %v", err)
	}

	log.Printf("super admin %q created, log in with phone %s", *name, *phone)
}

// readPassword returns $ADMIN_PASSWORD, or the first line of stdin when it is unset
func readPassword() (string, error) {
	if password, ok := os.LookupEnv(passwordEnv); ok {
		return password, nil
	}

	fmt.Fprint(os.Stderr, "Password: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
