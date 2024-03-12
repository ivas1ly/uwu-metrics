package migrations

import "embed"

// Migrations variable to access SQL files with migrations for the database.
//
// Issue - https://github.com/golang/go/issues/46056
//
//go:embed *.sql
var Migrations embed.FS
