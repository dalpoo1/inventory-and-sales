// Package migrations embeds the SQL migration files so they can be applied
// by the in-process migration runner (used by tests and local development).
//
// The same files are mounted into the Postgres container's
// /docker-entrypoint-initdb.d directory by docker-compose so a fresh database
// is migrated and seeded automatically.
package migrations

import "embed"

// FS holds the ordered SQL migration files.
//
//go:embed *.sql
var FS embed.FS
