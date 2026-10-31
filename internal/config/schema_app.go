package config

import "time"

type App struct {
	Name       string     `mapstructure:"name"`
	HTTPServer HTTPServer `mapstructure:"http_server"`
	Logger     Logger     `mapstructure:"logger"`
	Database   Database   `mapstructure:"database"`
	Auth       Auth       `mapstructure:"auth"`
	Cache      Cache      `mapstructure:"cache"`
}

type HTTPServer struct {
	ListenAddress   string        `mapstructure:"listen_address"`
	Port            string        `mapstructure:"port"`
	GracefulTimeout time.Duration `mapstructure:"graceful_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
}

type Logger struct {
	Level string `mapstructure:"level"`
	Debug bool   `mapstructure:"debug"`
}

// Database selects the storage engine, see pkg/database
type Database struct {
	Driver          string        `mapstructure:"driver"` // sqlite | postgres
	DSN             string        `mapstructure:"dsn"`
	MigrateDSN      string        `mapstructure:"migrate_dsn"` // optional, falls back to dsn
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	AutoMigrate     bool          `mapstructure:"auto_migrate"` // apply pending migrations on startup
}

// Auth configures access tokens, see pkg/token
type Auth struct {
	JWTSecret string        `mapstructure:"jwt_secret"` // at least 32 bytes
	TokenTTL  time.Duration `mapstructure:"token_ttl"`
}

// Cache selects the key-value store, see pkg/cache
type Cache struct {
	Driver string `mapstructure:"driver"` // memory | redis
	DSN    string `mapstructure:"dsn"`    // redis only
}
