package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"honnef.co/go/tools/staticcheck"
)

func TestAddStaticcheckAnalyzers(t *testing.T) {
	analyzers := addStaticcheckAnalyzers()
	var sa, st, s1, qf int

	for _, a := range analyzers {
		switch {
		case strings.HasPrefix(a.Name, "SA"):
			sa++
		case strings.HasPrefix(a.Name, "S1"):
			s1++
		case strings.HasPrefix(a.Name, "ST"):
			st++
		case strings.HasPrefix(a.Name, "QF"):
			qf++
		default:
			t.Fatalf("unknown analyzer %s", a.Name)
		}
	}
	require.Len(t, staticcheck.Analyzers, sa)
	require.Equal(t, 1, s1, "analyzers must have only one Simple analyzer")
	require.Equal(t, 1, st, "analyzers must have only one Style analyzer")
	require.Equal(t, 1, qf, "analyzers must have only one Qickfix analyzer")
}
