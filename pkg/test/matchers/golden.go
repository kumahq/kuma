package matchers

import (
	"os"
	"path/filepath"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/test/matchers/golden"
)

func MatchGoldenYAML(goldenFilePath ...string) types.GomegaMatcher {
	return MatchGolden(gomega.MatchYAML, goldenFilePath...)
}

func MatchGoldenJSON(goldenFilePath ...string) types.GomegaMatcher {
	return MatchGolden(gomega.MatchJSON, goldenFilePath...)
}

func MatchGoldenXML(goldenFilePath ...string) types.GomegaMatcher {
	return MatchGolden(gomega.MatchXML, goldenFilePath...)
}

func MatchGoldenEqual(goldenFilePath ...string) types.GomegaMatcher {
	return MatchGolden(func(expected interface{}) types.GomegaMatcher {
		if expectedBytes, ok := expected.([]byte); ok {
			expected = string(expectedBytes)
		}
		return gomega.Equal(expected)
	}, goldenFilePath...)
}

type MatcherFn = func(expected interface{}) types.GomegaMatcher

// MatchGolden matches Golden file overriding it with actual content if UPDATE_GOLDEN_FILES is set to true
func MatchGolden(matcherFn MatcherFn, goldenFilePath ...string) types.GomegaMatcher {
	return &GoldenMatcher{
		MatcherFactory: matcherFn,
		GoldenFilePath: filepath.Join(goldenFilePath...),
	}
}

type GoldenMatcher struct {
	MatcherFactory MatcherFn
	Matcher        types.GomegaMatcher
	GoldenFilePath string
}

var _ types.GomegaMatcher = &GoldenMatcher{}

func (g *GoldenMatcher) Match(actual interface{}) (bool, error) {
	actualContent, err := g.actualString(actual)
	if err != nil {
		return false, err
	}
	if golden.UpdateGoldenFiles() {
		if actualContent == "" { // If there's no actual content, we remove the golden file
			_ = os.Remove(g.GoldenFilePath)
		} else {
			if actualContent != "" && actualContent[len(actualContent)-1] != '\n' {
				actualContent += "\n"
			}
			err := os.WriteFile(g.GoldenFilePath, []byte(actualContent), 0o600)
			if err != nil {
				return false, errors.Wrap(err, "could not update golden file")
			}
		}
	}
	// No content means that there shouldn't be a golden file (this avoids keep empty golden files)
	if actualContent == "" {
		return gomega.Not(gomega.BeAnExistingFile()).Match(g.GoldenFilePath)
	}
	expected, err := os.ReadFile(g.GoldenFilePath)
	if err != nil {
		return false, errors.Wrapf(err, "could not read golden file to compare with: '%v'", actualContent)
	}

	// Generate a new instance of the matcher for this match. Since
	// the matcher might keep internal state, we want to keep the same
	// instance for subsequent message calls.
	g.Matcher = g.MatcherFactory(expected)

	return g.Matcher.Match(actualContent)
}

func (g *GoldenMatcher) FailureMessage(actual interface{}) string {
	actualContent, err := g.actualString(actual)
	if err != nil {
		return err.Error()
	}
	if actualContent == "" {
		return golden.RerunMsg(g.GoldenFilePath) + "\n\n exists but output is empty, which means that the golden file should not exist."
	}
	return golden.RerunMsg(g.GoldenFilePath) + "\n\n" + g.Matcher.FailureMessage(actualContent)
}

func (g *GoldenMatcher) NegatedFailureMessage(actual interface{}) string {
	actualContent, err := g.actualString(actual)
	if err != nil {
		return err.Error()
	}
	return g.Matcher.NegatedFailureMessage(actualContent)
}

func (g *GoldenMatcher) actualString(actual interface{}) (string, error) {
	switch actual := actual.(type) {
	case nil:
		return "", nil
	case error:
		return actual.Error(), nil
	case []byte:
		return string(actual), nil
	case string:
		return actual, nil
	default:
		return "", errors.Errorf("not supported type %T for MatchGolden", actual)
	}
}
