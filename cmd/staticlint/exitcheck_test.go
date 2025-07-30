package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNoOsExit(t *testing.T) {
	testdata := analysistest.TestData()

	analysistest.Run(t, testdata, exitAnalyzer, "bad", "good")
}
