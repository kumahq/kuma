package linter_test

import (
    "github.com/kumahq/kuma/tools/ci/api-linter/linter"
    "testing"

    "golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
    testdata := analysistest.TestData()
    analysistest.Run(t, testdata, linter.Analyzer, "valid", "invalid_mergeable", "invalid_nonmergeable")
}
