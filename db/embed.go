package db

import "embed"

// Migrations -
// Issue - https://github.com/golang/go/issues/46056
//
//go:embed migrations/*.sql
var Migrations embed.FS
