package matchers

import (
	"io/ioutil"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/test/golden"
)

func MatchGoldenYAML(goldenFilePath string) types.GomegaMatcher {
	return MatchGolden(gomega.MatchYAML, goldenFilePath)
}

func MatchGoldenJSON(goldenFilePath string) types.GomegaMatcher {
	return MatchGolden(gomega.MatchJSON, goldenFilePath)
}

func MatchGoldenEqual(goldenFilePath string) types.GomegaMatcher {
	return MatchGolden(func(expected interface{}) types.GomegaMatcher {
		if expectedBytes, ok := expected.([]byte); ok {
			expected = string(expectedBytes)
		}
		return gomega.Equal(expected)
	}, goldenFilePath)
}

type MatcherFn = func(expected interface{}) types.GomegaMatcher

// MatchGolden matches Golden file overriding it with actual content if UPDATE_GOLDEN_FILES is set to true
func MatchGolden(matcherFn MatcherFn, goldenFilePath string) types.GomegaMatcher {
	return &GoldenYAMLMatcher{
		MatcherFn:      matcherFn,
		GoldenFilePath: goldenFilePath,
	}
}

type GoldenYAMLMatcher struct {
	MatcherFn      MatcherFn
	GoldenFilePath string
}

var _ types.GomegaMatcher = &GoldenYAMLMatcher{}

func (g *GoldenYAMLMatcher) Match(actual interface{}) (success bool, err error) {
	actualContent, err := g.actualBytes(actual)
	if err != nil {
		return false, err
	}
	if golden.UpdateGoldenFiles() {
		err := ioutil.WriteFile(g.GoldenFilePath, []byte(actualContent), 0644)
		if err != nil {
			return false, errors.Wrap(err, "could not update golden file")
		}
	}
	expected, err := ioutil.ReadFile(g.GoldenFilePath)
	if err != nil {
		return false, errors.Wrap(err, "could not read golden file")
	}
	return g.MatcherFn(expected).Match(actualContent)
}

func (g *GoldenYAMLMatcher) FailureMessage(actual interface{}) (message string) {
	actualContent, err := g.actualBytes(actual)
	if err != nil {
		return err.Error()
	}
	expected, err := ioutil.ReadFile(g.GoldenFilePath)
	if err != nil {
		return errors.Wrap(err, "could not read golden file").Error()
	}
	return golden.RerunMsg + "\n\n" + g.MatcherFn(expected).FailureMessage(actualContent)
}

func (g *GoldenYAMLMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	actualContent, err := g.actualBytes(actual)
	if err != nil {
		return err.Error()
	}
	expected, err := ioutil.ReadFile(g.GoldenFilePath)
	if err != nil {
		return errors.Wrap(err, "could not read golden file").Error()
	}
	return g.MatcherFn(expected).NegatedFailureMessage(actualContent)
}

func (g *GoldenYAMLMatcher) actualBytes(actual interface{}) (string, error) {
	switch actual := actual.(type) {
	case []byte:
		return string(actual), nil
	case string:
		return actual, nil
	default:
		return "", errors.Errorf("not supported type %T for MatchGolden", actual)
	}
}
