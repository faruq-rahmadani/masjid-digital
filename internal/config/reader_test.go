package config

import (
	"os"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/pkg/env"
)

// skipWithProductionFile keeps the tests off a machine that has the real production file
func skipWithProductionFile(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(configFiles[env.EnvProduction]); err == nil {
		t.Skip("production config file exists on this machine")
	}
}

func TestReadYamlProductionFromEnvOnly(t *testing.T) {
	skipWithProductionFile(t)
	t.Setenv("APP_DATABASE_DRIVER", "postgres")
	t.Setenv("APP_DATABASE_DSN", "postgres://app@db.example:5432/postgres")
	t.Setenv("APP_AUTH_JWT_SECRET", "a-secret-that-is-at-least-32-bytes-long")
	t.Setenv("APP_HTTP_SERVER_ALLOWED_ORIGINS", "https://admin.example.app,https://tv.example.app")
	t.Setenv("PORT", "10000")

	cfg, err := readYaml(env.EnvProduction)
	if err != nil {
		t.Fatalf("readYaml: %v", err)
	}

	app := cfg.App
	if app.Database.Driver != "postgres" || app.Database.DSN != "postgres://app@db.example:5432/postgres" {
		t.Errorf("database = %+v", app.Database)
	}
	if app.Auth.JWTSecret != "a-secret-that-is-at-least-32-bytes-long" || app.Auth.TokenTTL != 24*time.Hour {
		t.Errorf("auth = %+v", app.Auth)
	}
	if app.HTTPServer.Port != "10000" || app.HTTPServer.GracefulTimeout != 30*time.Second {
		t.Errorf("http server = %+v", app.HTTPServer)
	}
	origins := app.HTTPServer.AllowedOrigins
	if len(origins) != 2 || origins[0] != "https://admin.example.app" || origins[1] != "https://tv.example.app" {
		t.Errorf("allowed origins = %q", origins)
	}
	if !app.Database.AutoMigrate || app.Cache.Driver != "memory" {
		t.Errorf("defaults not applied: %+v %+v", app.Database, app.Cache)
	}
}

func TestReadYamlPortPrefersAppVariable(t *testing.T) {
	skipWithProductionFile(t)
	t.Setenv("PORT", "10000")
	t.Setenv("APP_HTTP_SERVER_PORT", "8080")

	cfg, err := readYaml(env.EnvProduction)
	if err != nil {
		t.Fatalf("readYaml: %v", err)
	}
	if cfg.App.HTTPServer.Port != "8080" {
		t.Errorf("port = %q, want 8080", cfg.App.HTTPServer.Port)
	}
}

func TestReadYamlDevelopmentNeedsTheFile(t *testing.T) {
	t.Chdir(t.TempDir())

	if _, err := readYaml(env.EnvDevelopment); err == nil {
		t.Fatal("want an error without config.yaml in development")
	}
}
