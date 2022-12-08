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
		predicate: func(item StatItem) bool {
			return item.Value == expected
		},
		msg: func() string {
			return fmt.Sprintf("equal %v", expected)
		},
	}
}

func BeEqualZero() types.GomegaMatcher {
	return BeEqual(float64(0))
}

func BeGreaterThan(other interface{}) types.GomegaMatcher {
	var expected float64
	if si, ok := other.(StatItem); ok {
		expected = si.Value.(float64)
	} else {
		expected = other.(float64)
	}
	return &statMatcher{
		predicate: func(item StatItem) bool {
			return item.Value.(float64) > expected
		},
		msg: func() string {
			return fmt.Sprintf("greater than %v", expected)
		},
	}
}

func BeGreaterThanZero() types.GomegaMatcher {
	return BeGreaterThan(float64(0))
}

type statMatcher struct {
	predicate func(StatItem) bool
	msg       func() string
}

func (m *statMatcher) Match(actual interface{}) (bool, error) {
	stats, ok := actual.(*Stats)
	if !ok {
		return false, fmt.Errorf("BeEqual matcher expects a Stats")
	}

	if len(stats.Stats) == 0 {
		return false, fmt.Errorf("no stat found: %+q", stats)
	}

	if len(stats.Stats) > 1 {
		return false, fmt.Errorf("actual stats have more items than 1: %+q", stats)
	}

	return m.predicate(stats.Stats[0]), nil
}

func (m *statMatcher) FailureMessage(actual interface{}) string {
	actualStats := actual.(*Stats)
	actualStat := actualStats.Stats[0]
	return fmt.Sprintf(": %v %s to be: %s", actualStat.Name, actualStat.Value, m.msg())
}

func (m *statMatcher) NegatedFailureMessage(actual interface{}) string {
	actualStats := actual.(*Stats)
	actualStat := actualStats.Stats[0]
	return fmt.Sprintf(": %v %s not to be: %s", actualStat.Name, actualStat.Value, m.msg())
}
