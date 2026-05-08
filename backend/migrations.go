package main

import "embed"

// migrationsFS contains the embedded database migration files.
//go:embed db/migrations/*.sql
var migrationsFS embed.FS
