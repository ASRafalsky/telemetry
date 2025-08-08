package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = `check for direct os.Exit calls in main function

This analyzer reports direct calls to os.Exit in the main function of the main package.`

var exitAnalyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// This is the main pkg?
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Try to find os pkg import.
	osImported := false
	for _, imp := range pass.Pkg.Imports() {
		if imp.Path() == "os" {
			osImported = true
			break
		}
	}

	if !osImported {
		return nil, nil // No way.
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)

		// This is main func?
		if fn.Name.Name != "main" {
			return
		}

		// Try to find os.Exit call inside the main func body.
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Name == "os" && sel.Sel.Name == "Exit" {
				pass.Reportf(call.Pos(), "direct call to os.Exit in main function of main package")
			}

			return true
		})
	})

	return nil, nil
}
