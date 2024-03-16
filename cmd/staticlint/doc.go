// Package defines a multichecker static linter with checks from golang.org/x/tools/go/analysis/passes,
// staticcheck.io, other public analyzers from GitHub, and one custom analyzer to check
// the os.Exit call in the main function of the main package.
//
// # Usage
//
// Build the multichecker:
//
//	go build -o ./cmd/staticlint/multichecker ./cmd/staticlint/main.go
//
// Get a description of all the static checks:
//
//	./cmd/staticlint/multichecker help
//
// Run multichecker:
//
//	./cmd/staticlint/multichecker ./...
//
// # Source
//
//   - Standart analyzers — https://golang.org/x/tools/go/analysis/passes
//
// Analyzers from staticcheck.io:
//   - SA — staticcheck https://staticcheck.dev/docs/checks/#SA (all checks).
//   - S — simple https://staticcheck.dev/docs/checks/#S (S1000, S1002, S1005).
//   - ST — stylecheck https://staticcheck.dev/docs/checks/#ST (ST1013, ST1015, ST1017).
//   - QF — quickfix https://staticcheck.dev/docs/checks/#QF (QF1004, QF1011, QF1012).
//
// Other public analyzers:
//   - smrcptr — https://github.com/nikolaydubina/smrcptr
//   - bodyclose — https://github.com/timakin/bodyclose
//
// Custom analyzers:
//   - See pkg/staticlint/exitcall package.
package main
