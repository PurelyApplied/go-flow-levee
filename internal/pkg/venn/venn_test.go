package venn

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestVenn(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "example.com/foo")
}