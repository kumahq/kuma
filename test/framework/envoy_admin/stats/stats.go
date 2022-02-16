package stats

import (
	"fmt"

	"github.com/onsi/gomega/types"
)

type StatItem struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type Stats struct {
	Stats []StatItem `json:"stats"`
}

func BeEqual(expected interface{}) types.GomegaMatcher {
	return &statMatcher{
		expected:  expected,
		predicate: &equalPredicate,
	}
}

func BeGreaterThanZero() types.GomegaMatcher {
	return &statMatcher{
		predicate: &greaterThanZeroPredicate,
	}
}

func BeEqualZero() types.GomegaMatcher {
	return &statMatcher{
		predicate: &equalZeroPredicate,
	}
}

var equalZero = func(stat StatItem) bool {
	return stat.Value.(float64) == 0
}

var equalZeroPredicate = func(interface{}) func(item StatItem) bool {
	return equalZero
}

var greaterThanZero = func(stat StatItem) bool {
	return stat.Value.(float64) > 0
}

var greaterThanZeroPredicate = func(interface{}) func(item StatItem) bool {
	return greaterThanZero
}

var equalPredicate = func(expected interface{}) func(item StatItem) bool {
	return func(stat StatItem) bool {
		return int(stat.Value.(float64)) == expected.(int)
	}
}

type statMatcher struct {
	expected  interface{}
	predicate *func(interface{}) func(StatItem) bool
}

func (m *statMatcher) Match(actual interface{}) (success bool, err error) {
	stats, ok := actual.(Stats)
	if !ok {
		return false, fmt.Errorf("BeEqual matcher expects a Stats")
	}

	if len(stats.Stats) == 0 {
		return false, fmt.Errorf("no stat found: %+q", stats)
	}

	if len(stats.Stats) > 1 {
		return false, fmt.Errorf("actual stats have more items than 1: %+q", stats)
	}

	return (*m.predicate)(m.expected)(stats.Stats[0]), nil
}

func (m *statMatcher) genFailureMessage(toBeOrNotToBe string, actual interface{}) (message string) {
	actualStats := actual.(Stats)
	actualStat := actualStats.Stats[0]

	var expectation string
	switch m.predicate {
	case &equalPredicate:
		expectation = fmt.Sprintf("%s: %v %s equal %v", toBeOrNotToBe, actualStat.Name, actualStat.Value, m.expected)
	case &equalZeroPredicate:
		expectation = fmt.Sprintf("%s: %v %s equal 0", toBeOrNotToBe, actualStat.Name, actualStat.Value)
	case &greaterThanZeroPredicate:
		expectation = fmt.Sprintf("%s: %v %s greater than 0", actualStat.Name, actualStat.Value, toBeOrNotToBe)
	default:
		panic("unknown predicate")
	}

	return fmt.Sprintf("Expected %s", expectation)
}

func (m *statMatcher) FailureMessage(actual interface{}) (message string) {
	return m.genFailureMessage("to be", actual)
}

func (m *statMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return m.genFailureMessage("not to be", actual)
}
