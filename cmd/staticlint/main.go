// Package main provides a custom static checker.
//
// This package performs static code analysis.
// It consists of custom analyzers and analyzers from external packages.
package main

import (
	"github.com/gostaticanalysis/funcstat"
	"github.com/gostaticanalysis/zapvet/passes/fieldtype"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"

	"github.com/derpartizanen/metrics/pkg/osexitchecker"
)

// main function registers and runs a set of analyzers
func main() {
	var analyzers []*analysis.Analyzer

	// Add all analyzers from staticcheck package
	for _, v := range staticcheck.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}
	// Add analyzer from quickfix package
	for _, qf := range quickfix.Analyzers {
		if qf.Analyzer.Name == "QF1006" { // Lift if+break into loop condition
			analyzers = append(analyzers, qf.Analyzer)
		}
	}

	// Add custom analyzer
	analyzers = append(analyzers, osexitchecker.Analyzer)

	// Add external analyzers
	analyzers = append(analyzers, fieldtype.Analyzer)
	analyzers = append(analyzers, funcstat.Analyzer)

	// Add passes analyzers
	analyzers = append(analyzers, assign.Analyzer)
	analyzers = append(analyzers, findcall.Analyzer)
	analyzers = append(analyzers, inspect.Analyzer)
	analyzers = append(analyzers, printf.Analyzer)
	analyzers = append(analyzers, shadow.Analyzer)
	analyzers = append(analyzers, shift.Analyzer)
	analyzers = append(analyzers, structtag.Analyzer)

	// Run all analyzers
	multichecker.Main(analyzers...)
}
