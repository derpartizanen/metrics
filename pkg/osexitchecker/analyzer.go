// Package osexitchecker provides detection of usage os.Exit() call in main func
package osexitchecker

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

const (
	mainPkg    = "main"
	mainFunc   = "main"
	osPkg      = "os"
	osExitFunc = "Exit"
)

var Analyzer = &analysis.Analyzer{
	Name: "osexitchecker",
	Doc:  "checks os.Exit() call in main func",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		foundPos := FindOsExitInMain(file)
		if foundPos != nil {
			pos := *foundPos
			pass.Reportf(pos, "detected os.Exit() call in main function")
		}
	}
	return nil, nil
}

// FindOsExitInMain returns the position where the os.Exit call was found
func FindOsExitInMain(file ast.Node) *token.Pos {
	var pos *token.Pos
	ast.Inspect(file, func(n1 ast.Node) bool {
		pkg, okFile := n1.(*ast.File)
		if !okFile || pkg.Name.Name != mainPkg {
			return true
		}

		ast.Inspect(pkg, func(n2 ast.Node) bool {
			a, okFuncDecl := n2.(*ast.FuncDecl)
			if !okFuncDecl || a.Name.Name != mainFunc {
				return true
			}

			ast.Inspect(a.Body, func(bodyNode ast.Node) bool {
				callExpr, ok := bodyNode.(*ast.CallExpr)
				if !ok {
					return true
				}

				selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := selectorExpr.X.(*ast.Ident)
				if !ok || ident.Name != osPkg || selectorExpr.Sel.Name != osExitFunc {
					return true
				}

				pos = &ident.NamePos
				return false
			})
			return false
		})
		return false
	})

	return pos
}
