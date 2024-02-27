package web

import "embed"

// Templates variable for accessing HTML files for website handlers.
// Issue - https://github.com/golang/go/issues/46056
//
//go:embed templates
var Templates embed.FS
