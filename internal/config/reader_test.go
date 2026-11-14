package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/pkg/env"
)

const testConfig = `app:
  name: masjid-digital
  http_server:
    port: "3000"
  auth:
    jwt_secret: from-file
    token_ttl: 24h
`

// writeDevConfig writes testConfig as the development config.yaml in a fresh working directory
func writeDevConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(testConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
}

func TestReadYamlUsesFileValues(t *testing.T) {
	writeDevConfig(t)

	cfg, err := readYaml(env.EnvDevelopment)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.App.Auth.JWTSecret != "from-file" {
		t.Errorf("jwt_secret = %q, want %q", cfg.App.Auth.JWTSecret, "from-file")
	}
}

func TestReadYamlEnvOverridesFileValues(t *testing.T) {
	writeDevConfig(t)
	t.Setenv("MASJID_APP_AUTH_JWT_SECRET", "from-env")
	t.Setenv("MASJID_APP_HTTP_SERVER_PORT", "8080")

	cfg, err := readYaml(env.EnvDevelopment)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.App.Auth.JWTSecret != "from-env" {
		t.Errorf("jwt_secret = %q, want %q", cfg.App.Auth.JWTSecret, "from-env")
	}
	if cfg.App.HTTPServer.Port != "8080" {
		t.Errorf("port = %q, want %q", cfg.App.HTTPServer.Port, "8080")
	}
	if cfg.App.Name != "masjid-digital" {
		t.Errorf("name = %q, want the file value kept", cfg.App.Name)
	}
}
