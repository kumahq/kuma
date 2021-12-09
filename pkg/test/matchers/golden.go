package matchers

import (
	"os"
	"path/filepath"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/test/golden"
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
	return &GoldenYAMLMatcher{
		MatcherFactory: matcherFn,
		GoldenFilePath: filepath.Join(goldenFilePath...),
	}
}

type GoldenYAMLMatcher struct {
	MatcherFactory MatcherFn
	Matcher        types.GomegaMatcher
	GoldenFilePath string
}

var _ types.GomegaMatcher = &GoldenYAMLMatcher{}

func (g *GoldenYAMLMatcher) Match(actual interface{}) (success bool, err error) {
	actualContent, err := g.actualString(actual)
	if err != nil {
		return false, err
	}
	if golden.UpdateGoldenFiles() {
		if actualContent[len(actualContent)-1] != '\n' {
			actualContent += "\n"
		}
		err := os.WriteFile(g.GoldenFilePath, []byte(actualContent), 0644)
		if err != nil {
			return false, errors.Wrap(err, "could not update golden file")
		}
	}
	expected, err := os.ReadFile(g.GoldenFilePath)
	if err != nil {
		return false, errors.Wrap(err, "could not read golden file")
	}

	// Generate a new instance of the matcher for this match. Since
	// the matcher might keep internal state, we want to keep the same
	// instance for subsequent message calls.
	g.Matcher = g.MatcherFactory(expected)

	return g.Matcher.Match(actualContent)
}

func (g *GoldenYAMLMatcher) FailureMessage(actual interface{}) (message string) {
	actualContent, err := g.actualString(actual)
	if err != nil {
		return err.Error()
	}
	return golden.RerunMsg + "\n\n" + g.Matcher.FailureMessage(actualContent)
}

func (g *GoldenYAMLMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	actualContent, err := g.actualString(actual)
	if err != nil {
		return err.Error()
	}
	return g.Matcher.NegatedFailureMessage(actualContent)
}

func (g *GoldenYAMLMatcher) actualString(actual interface{}) (string, error) {
	switch actual := actual.(type) {
	case []byte:
		return string(actual), nil
	case string:
		return actual, nil
	default:
		return "", errors.Errorf("not supported type %T for MatchGolden", actual)
	}
}
