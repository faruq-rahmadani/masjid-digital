// Command admin manages user accounts from the server, for what the app itself cannot do:
// creating the first super admin of an empty database, and resetting a password nobody
// can reset from the app. It reads the same configuration as cmd/http (ENV, config file,
// APP_* variables). The password is asked on the terminal, or read from stdin when piped.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/faruq-rahmadani/masjid-digital/internal/config"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	repodb "github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database/users"
	"github.com/faruq-rahmadani/masjid-digital/pkg/database"
	"github.com/faruq-rahmadani/masjid-digital/pkg/env"
	"github.com/faruq-rahmadani/masjid-digital/pkg/hasher"
)

const usage = `usage: admin <command> [flags]

commands:
  create-super-admin -name NAME -phone PHONE [-email EMAIL]
        create a super admin; the database must already be migrated
  reset-password (-phone PHONE | -email EMAIL)
        set a new password for an existing user`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	if err := env.Init(os.Getenv("ENV")); err != nil {
		log.Fatalf("error initializing env: %v", err)
	}
	cfg, err := config.LoadAll(env.GetEnv())
	if err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}

	ctx := context.Background()
	db, err := database.Open(ctx, cfg.App.Database.ConnConfig())
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()

	a := admin{
		users:  users.NewUserDBRepository(repodb.DatabaseClient{Querier: db}),
		hasher: hasher.NewArgon2(hasher.DefaultArgon2Params),
		stdin:  bufio.NewReader(os.Stdin),
	}

	commands := map[string]func(context.Context, []string) error{
		"create-super-admin": a.createSuperAdmin,
		"reset-password":     a.resetPassword,
	}
	run, ok := commands[os.Args[1]]
	if !ok {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	if err := run(ctx, os.Args[2:]); err != nil {
		log.Fatalf("%s: %v", os.Args[1], err)
	}
}

type admin struct {
	users  *users.UserDBRepository
	hasher hasher.PasswordHasher
	stdin  *bufio.Reader
}

func (a admin) createSuperAdmin(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("create-super-admin", flag.ExitOnError)
	name := fs.String("name", "", "full name")
	phone := fs.String("phone", "", "phone, e.g. 081234567890 or +6281234567890")
	email := fs.String("email", "", "email (optional)")
	_ = fs.Parse(args)

	password, err := a.password()
	if err != nil {
		return err
	}
	req := domain.CreateUserRequest{
		Name: *name, Phone: *phone, Email: *email, Password: password, Role: domain.UserRoleSuperAdmin,
	}
	if err := req.ValidateAndNormalize(); err != nil {
		return err
	}
	hash, err := a.hasher.Hash(req.Password)
	if err != nil {
		return err
	}

	id, err := a.users.Create(ctx, req.ToUser(hash))
	if err != nil {
		return err
	}
	log.Printf("created super admin %s (id %d, login with %s)", req.Name, id, req.Phone)
	return nil
}

func (a admin) resetPassword(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("reset-password", flag.ExitOnError)
	phone := fs.String("phone", "", "the user's phone")
	email := fs.String("email", "", "the user's email")
	_ = fs.Parse(args)

	user, err := a.findUser(ctx, *phone, *email)
	if err != nil {
		return err
	}
	password, err := a.password()
	if err != nil {
		return err
	}
	req := domain.ResetPasswordRequest{UserID: user.ID, NewPassword: password}
	if err := req.Validate(); err != nil {
		return err
	}
	hash, err := a.hasher.Hash(req.NewPassword)
	if err != nil {
		return err
	}

	if err := a.users.UpdatePasswordHash(ctx, user.ID, hash); err != nil {
		return err
	}
	log.Printf("password reset for %s (id %d); existing sessions stay until they expire", user.Name, user.ID)
	return nil
}

// findUser looks the user up by phone, or by email when no phone is given
func (a admin) findUser(ctx context.Context, phone, email string) (domain.User, error) {
	if phone != "" {
		normalized, err := domain.NormalizePhone(phone)
		if err != nil {
			return domain.User{}, err
		}
		return a.users.GetByPhone(ctx, normalized)
	}
	if email == "" {
		return domain.User{}, errors.New("give -phone or -email")
	}
	normalized, err := domain.NormalizeEmail(email)
	if err != nil {
		return domain.User{}, err
	}
	return a.users.GetByEmail(ctx, normalized)
}

// password asks twice on a terminal without echo, or reads one line from piped stdin
func (a admin) password() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		line, err := a.stdin.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	first, err := prompt(fd, "New password: ")
	if err != nil {
		return "", err
	}
	second, err := prompt(fd, "Repeat password: ")
	if err != nil {
		return "", err
	}
	if first != second {
		return "", errors.New("passwords do not match")
	}
	return first, nil
}

func prompt(fd int, label string) (string, error) {
	fmt.Fprint(os.Stderr, label)
	b, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	return string(b), err
}
