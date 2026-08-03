package stats

import (
	"fmt"
	"strconv"

	"github.com/onsi/gomega/types"
)

type StatItem struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type Stats struct {
	Stats []StatItem `json:"stats"`
}

// SingleValue returns the numeric value of the one stat these Stats are
// expected to contain. A filter that matches nothing reads as 0 (the counter
// Envoy has not created yet), and matching more than one stat is an error
// because the filter was meant to identify exactly one counter - the same
// exactly-one invariant statMatcher enforces.
func (s *Stats) SingleValue() (float64, error) {
	switch len(s.Stats) {
	case 0:
		return 0, nil
	case 1:
		return s.Stats[0].floatValue()
	default:
		return 0, fmt.Errorf("stats filter matched %d stats, expected 1: %+q", len(s.Stats), s)
	}
}

func (i StatItem) floatValue() (float64, error) {
	switch v := i.Value.(type) {
	case float64:
		return v, nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("stat %q has non-numeric value %q: %w", i.Name, v, err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("stat %q has unexpected value type %T", i.Name, i.Value)
	}
}

func BeEqual(expected any) types.GomegaMatcher {
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

func BeGreaterThan(other any) types.GomegaMatcher {
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

func (m *statMatcher) Match(actual any) (bool, error) {
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

func (m *statMatcher) FailureMessage(actual any) string {
	actualStats := actual.(*Stats)
	actualStat := actualStats.Stats[0]
	return fmt.Sprintf(": %v %s to be: %s", actualStat.Name, actualStat.Value, m.msg())
}

func (m *statMatcher) NegatedFailureMessage(actual any) string {
	actualStats := actual.(*Stats)
	actualStat := actualStats.Stats[0]
	return fmt.Sprintf(": %v %s not to be: %s", actualStat.Name, actualStat.Value, m.msg())
}
