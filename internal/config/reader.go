package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/faruq-rahmadani/masjid-digital/pkg/env"
	"github.com/spf13/viper"
)

// configFiles maps each environment to its hardcoded config file location
var configFiles = map[string]string{
	env.EnvDevelopment: "config.yaml", // project root
	env.EnvProduction:  "/etc/masjid-digital-config/config.yaml",
}

// envPrefix starts every environment variable that overrides a config key. The
// rest is the key path in upper case with dots as underscores, e.g.
// MASJID_APP_AUTH_JWT_SECRET overrides app.auth.jwt_secret.
const envPrefix = "MASJID"

// readYaml reads the configuration from the config file of the given environment.
// An environment variable overrides a key present in the file, so a container
// can ship a full config and take secrets from its environment.
func readYaml(environment string) (*Schema, error) {
	configFile, ok := configFiles[environment]
	if !ok {
		return nil, fmt.Errorf("no config file for environment %q", environment)
	}

	log.Printf("Reading configuration for environment %s: %s", environment, configFile)

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var schema Schema
	if err := v.Unmarshal(&schema); err != nil {
		return nil, err
	}

	return &schema, nil
}
