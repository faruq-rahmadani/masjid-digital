package database

import (
	"fmt"
	"net/url"
	"strings"

	// Register the database/sql drivers
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// sqliteDefaults are applied to a SQLite DSN unless it already sets them
var sqliteDefaults = []struct {
	key    string
	pragma string // name inside _pragma=name(value), empty for plain params
	value  string
}{
	{key: "_pragma", pragma: "journal_mode", value: "WAL"},  // readers do not block the writer
	{key: "_pragma", pragma: "foreign_keys", value: "1"},    // off by default in SQLite
	{key: "_pragma", pragma: "busy_timeout", value: "5000"}, // wait for the write lock instead of failing
	{key: "_txlock", value: "immediate"},                    // take the write lock at BEGIN to avoid SQLITE_BUSY on upgrade
	{key: "_time_format", value: "sqlite"},                  // store time as sortable text with timezone
}

// sqliteDSN normalizes a SQLite DSN ("masjid.db" or "file:masjid.db?...")
// and adds the defaults it does not set yet
func sqliteDSN(dsn string) (string, error) {
	if dsn == "" {
		return "", fmt.Errorf("sqlite dsn must not be empty")
	}

	path, rawQuery, _ := strings.Cut(dsn, "?")
	if !strings.HasPrefix(path, "file:") {
		path = "file:" + path
	}

	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", fmt.Errorf("error parsing sqlite dsn: %w", err)
	}

	for _, def := range sqliteDefaults {
		if def.pragma == "" {
			if !query.Has(def.key) {
				query.Set(def.key, def.value)
			}
			continue
		}
		if !hasPragma(query[def.key], def.pragma) {
			query.Add(def.key, def.pragma+"("+def.value+")")
		}
	}

	return path + "?" + query.Encode(), nil
}

func hasPragma(pragmas []string, name string) bool {
	for _, p := range pragmas {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(p)), name) {
			return true
		}
	}
	return false
}
