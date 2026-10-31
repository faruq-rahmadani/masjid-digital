package config

import (
	"github.com/faruq-rahmadani/masjid-digital/pkg/cache"
	"github.com/faruq-rahmadani/masjid-digital/pkg/database"
)

// ConnConfig converts the YAML database settings into a database.Config
func (d Database) ConnConfig() database.Config {
	return database.Config{
		Driver:          database.Driver(d.Driver),
		DSN:             d.DSN,
		MaxOpenConns:    d.MaxOpenConns,
		MaxIdleConns:    d.MaxIdleConns,
		ConnMaxLifetime: d.ConnMaxLifetime,
	}
}

// MigrateConnConfig is the connection used to run migrations. It uses
// migrate_dsn when set, so migrations can run as a user allowed to change the
// schema while the app connects as a user limited to reading and writing data.
func (d Database) MigrateConnConfig() database.Config {
	dsn := d.MigrateDSN
	if dsn == "" {
		dsn = d.DSN
	}

	return database.Config{
		Driver: database.Driver(d.Driver),
		DSN:    dsn,
	}
}

// ConnConfig converts the YAML cache settings into a cache.Config
func (c Cache) ConnConfig() cache.Config {
	return cache.Config{
		Driver: cache.Driver(c.Driver),
		DSN:    c.DSN,
	}
}
