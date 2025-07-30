// Package staticlint check go files with all standard analyzers from analysis/passes all SA analyzers and one analyzer
// from Simple, Style and Quickfix classes. Added custom analyzer to forbid call os.Exit directly from the main
// func of the main pkg.
//
// Using:
// build staticlint and call it for target file.go or whole project as ./staticlint ./..
//
// Stuff happens. If you're young and modern enough to use the latest version of GoLang,
// you might run into some issues like this: "staticlint: invoking "go tool vet" directly is unsupported;
// use "go vet". Do this "go vet -vettool=./multichecker" and fell good.
//
// Enjoy!
package main

import (
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/unitchecker"
	"honnef.co/go/tools/simple"

	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
)

func main() {
	// Add all standard analyzers from analysis/passes.
	analyzers := addPassesAnalyzers()

	// Add staticcheck.
	analyzers = append(analyzers, addStaticcheckAnalyzers()...)

	// Some additional analyzers.
	analyzers = append(analyzers, errcheck.Analyzer)  // Add errcheck analyzer. You know.
	analyzers = append(analyzers, bodyclose.Analyzer) // Add bodyclose analyzer. You surprised? I'm not.
	analyzers = append(analyzers, exitAnalyzer)       // Add custom noosexit analyzer.

	unitchecker.Main(analyzers...)
	os.Exit(1)
}

func addPassesAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}
}

func addStaticcheckAnalyzers() []*analysis.Analyzer {
	var staticcheckAnalyzers []*analysis.Analyzer

	// Add SA analyzers.
	for _, a := range staticcheck.Analyzers {
		if strings.HasPrefix(a.Analyzer.Name, "SA") {
			staticcheckAnalyzers = append(staticcheckAnalyzers, a.Analyzer)
		}
	}

	// Add another analyzer from other classes.
	// Simple
	for _, a := range simple.Analyzers {
		if strings.HasPrefix(a.Analyzer.Name, "S1") {
			staticcheckAnalyzers = append(staticcheckAnalyzers, a.Analyzer)
			break
		}
	}

	// Style
	for _, a := range stylecheck.Analyzers {
		if strings.HasPrefix(a.Analyzer.Name, "ST") {
			staticcheckAnalyzers = append(staticcheckAnalyzers, a.Analyzer)
			break
		}
	}

	// Quickfix
	for _, a := range quickfix.Analyzers {
		if strings.HasPrefix(a.Analyzer.Name, "QF") {
			staticcheckAnalyzers = append(staticcheckAnalyzers, a.Analyzer)
			break
		}
	}

	return staticcheckAnalyzers
}
