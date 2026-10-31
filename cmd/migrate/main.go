package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/faruq-rahmadani/masjid-digital/internal/config"
	"github.com/faruq-rahmadani/masjid-digital/migrations"
	"github.com/faruq-rahmadani/masjid-digital/pkg/database"
	"github.com/faruq-rahmadani/masjid-digital/pkg/env"
)

const usage = `usage: migrate <command>

commands:
  up      apply all pending migrations
  down    roll back the latest migration
  status  list migrations and whether they are applied
  version print the current schema version`

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, usage)
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

	ctx := context.Background()

	db, err := database.Open(ctx, cfg.App.Database.MigrateConnConfig())
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()

	m, err := db.NewMigrator(migrations.FS)
	if err != nil {
		log.Fatalf("error creating migrator: %v", err)
	}

	if err := run(ctx, m, os.Args[1]); err != nil {
		log.Fatalf("error running migrate %s: %v", os.Args[1], err)
	}
}

func run(ctx context.Context, m *database.Migrator, command string) error {
	switch command {
	case "up":
		results, err := m.Up(ctx)
		if err != nil {
			return err
		}
		if len(results) == 0 {
			log.Print("no pending migrations")
		}
		for _, r := range results {
			log.Printf("applied %s (%s)", r.Source.Path, r.Duration)
		}

	case "down":
		r, err := m.Down(ctx)
		if err != nil {
			return err
		}
		log.Printf("rolled back %s (%s)", r.Source.Path, r.Duration)

	case "status":
		statuses, err := m.Status(ctx)
		if err != nil {
			return err
		}
		for _, s := range statuses {
			appliedAt := "-"
			if !s.AppliedAt.IsZero() {
				appliedAt = s.AppliedAt.Format("2006-01-02 15:04:05")
			}
			fmt.Printf("%-8s %-20s %s\n", s.State, appliedAt, s.Source.Path)
		}

	case "version":
		version, err := m.Version(ctx)
		if err != nil {
			return err
		}
		fmt.Println(version)

	default:
		return fmt.Errorf("unknown command %q\n\n%s", command, usage)
	}

	return nil
}
