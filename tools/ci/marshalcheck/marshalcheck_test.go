package marshalcheck_test

import (
    "github.com/kumahq/kuma/tools/ci/marshalcheck"
    "testing"

    "golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
    testdata := analysistest.TestData()
    analysistest.Run(t, testdata, marshalcheck.Analyzer, "valid", "invalid_mergeable", "invalid_nonmergeable")
}
