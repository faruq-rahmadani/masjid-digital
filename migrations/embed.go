// Package migrations holds the schema migrations of every supported engine,
// embedded into the binary. Each engine has its own directory with the same
// version numbers, so a version means the same schema change everywhere.
package migrations

import "embed"

// FS contains one directory per database.Driver: sqlite/ and postgres/
//
//go:embed sqlite/*.sql postgres/*.sql
var FS embed.FS
