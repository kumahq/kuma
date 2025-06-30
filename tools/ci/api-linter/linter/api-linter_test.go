package linter_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/kumahq/kuma/tools/ci/api-linter/linter"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(
		t,
		testdata,
		linter.Analyzer,
		"valid",
		"invalid_mergeable",
		"invalid_nonmergeable",
		"invalid_type",
	)
}
