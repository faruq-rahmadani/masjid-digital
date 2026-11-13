package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"reflect"
	"strings"

	"github.com/faruq-rahmadani/masjid-digital/pkg/env"
	"github.com/spf13/viper"
)

// configFiles maps each environment to its hardcoded config file location
var configFiles = map[string]string{
	env.EnvDevelopment: "config.yaml", // project root
	env.EnvProduction:  "/etc/masjid-digital-config/config.yaml",
}

// defaults lets a host that only offers environment variables (Render, Fly, Railway) run
// without a config file; the values match config.yaml.example
var defaults = map[string]any{
	"app.name":                         "masjid-digital",
	"app.http_server.listen_address":   "0.0.0.0",
	"app.http_server.port":             "3000",
	"app.http_server.graceful_timeout": "30s",
	"app.http_server.read_timeout":     "30s",
	"app.http_server.write_timeout":    "30s",
	"app.http_server.idle_timeout":     "120s",
	"app.logger.level":                 "info",
	"app.database.driver":              "sqlite",
	"app.database.dsn":                 "masjid.db",
	"app.database.auto_migrate":        true,
	"app.auth.token_ttl":               "24h",
	"app.cache.driver":                 "memory",
}

// readYaml reads the configuration of the given environment. Every key can be set or
// overridden by an environment variable named after its path, e.g. app.database.dsn is
// APP_DATABASE_DSN. The port also follows PORT, which most hosts set. In production the
// file is optional, so the whole configuration can come from the environment.
func readYaml(environment string) (*Schema, error) {
	configFile, ok := configFiles[environment]
	if !ok {
		return nil, fmt.Errorf("no config file for environment %q", environment)
	}

	v := viper.New()
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	if err := bindEnv(v); err != nil {
		return nil, err
	}

	log.Printf("Reading configuration for environment %s: %s", environment, configFile)
	v.SetConfigFile(configFile)
	err := v.ReadInConfig()
	if err != nil && !(environment == env.EnvProduction && errors.Is(err, fs.ErrNotExist)) {
		return nil, err
	}
	if err != nil {
		log.Printf("%s not found, using environment variables only", configFile)
	}

	var schema Schema
	if err := v.Unmarshal(&schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

// bindEnv binds every key of the schema to its environment variable, so keys missing
// from the file still reach Unmarshal
func bindEnv(v *viper.Viper) error {
	for _, key := range schemaKeys(reflect.TypeFor[Schema](), "") {
		if err := v.BindEnv(key, strings.ToUpper(strings.ReplaceAll(key, ".", "_"))); err != nil {
			return err
		}
	}

	// PORT wins over the file, APP_HTTP_SERVER_PORT over PORT
	return v.BindEnv("app.http_server.port", "APP_HTTP_SERVER_PORT", "PORT")
}

// schemaKeys lists the dotted mapstructure paths of every leaf field of t
func schemaKeys(t reflect.Type, prefix string) []string {
	var keys []string
	for field := range t.Fields() {
		key := prefix + field.Tag.Get("mapstructure")
		if field.Type.Kind() != reflect.Struct || field.Type.PkgPath() != t.PkgPath() {
			keys = append(keys, key)
			continue
		}
		keys = append(keys, schemaKeys(field.Type, key+".")...)
	}

	return keys
}
