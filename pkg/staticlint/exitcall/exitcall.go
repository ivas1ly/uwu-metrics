// Package exitcall defines an Analyzer that checks for direct calls
// of the os.Exit function in main function of main package.
package exitcall

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	funcOrFileName = "main"
)

var Analyzer = &analysis.Analyzer{
	Name: "exitcall",
	Doc:  "check direct os.Exit calls in function main of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	checkExit := func(x *ast.CallExpr) {
		if s, ok := x.Fun.(*ast.SelectorExpr); ok {
			if id, ok := s.X.(*ast.Ident); ok {
				if id.Name == "os" && s.Sel.Name == "Exit" {
					pass.Reportf(id.NamePos, "direct call of the os.Exit function "+
						"in the main function of the main package")
				}
			}
		}
	}

	for _, file := range pass.Files {
		if !strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, ".go") {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.File:
				if x.Name.Name != funcOrFileName {
					return false
				}
			case *ast.FuncDecl:
				if x.Name.Name != funcOrFileName {
					return false
				}
			case *ast.CallExpr:
				checkExit(x)
			}
			return true
		})
	}

	return nil, nil
}
