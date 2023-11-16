package web

import "embed"

// Templates -
// Issue - https://github.com/golang/go/issues/46056
//
//go:embed templates
var Templates embed.FS
