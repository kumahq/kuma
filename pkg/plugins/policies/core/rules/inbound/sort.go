package inbound

import (
	"cmp"
	"slices"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/sort"
)

func Sort[T common.PolicyAttributes](list []T) {
	slices.SortStableFunc(list, sort.Compose(
		sort.CompareByPolicyAttributes[T],
		sort.CompareByDisplayName[T],
	))
}

func SortRules(list []*Rule) {
	slices.SortStableFunc(list, func(a, b *Rule) int {
		if less := CompareByMatches(a.Matches, b.Matches); less != 0 {
			return less
		}
		return 0
	})
}

func CompareByMatches(a, b []common_api.Match) int {
	return compareSingleMatch(firstMatch(a), firstMatch(b))
}

func firstMatch(matches []common_api.Match) *common_api.Match {
	if len(matches) == 0 {
		return nil
	}
	return &matches[0]
}

func compareSingleMatch(a, b *common_api.Match) int {
	if less := compareSpiffeID(matchSpiffeID(a), matchSpiffeID(b)); less != 0 {
		return less
	}
	if less := compareSNI(matchSNI(a), matchSNI(b)); less != 0 {
		return less
	}
	return 0
}

func matchSpiffeID(match *common_api.Match) *common_api.SpiffeIDMatch {
	if match == nil {
		return nil
	}
	return match.SpiffeID
}

func matchSNI(match *common_api.Match) *common_api.SNIMatch {
	if match == nil {
		return nil
	}
	return match.SNI
}

func compareSpiffeID(a, b *common_api.SpiffeIDMatch) int {
	score := func(m *common_api.SpiffeIDMatch) int {
		switch {
		case m == nil:
			return 0
		case m.Type == common_api.ExactMatchType:
			return 3
		case m.Type == common_api.PrefixMatchType:
			return 2
		default:
			return 0
		}
	}
	return cmp.Compare(score(b), score(a))
}

func compareSNI(a, b *common_api.SNIMatch) int {
	score := func(m *common_api.SNIMatch) int {
		if m == nil {
			return 0
		}
		return 1
	}
	return cmp.Compare(score(b), score(a))
}
