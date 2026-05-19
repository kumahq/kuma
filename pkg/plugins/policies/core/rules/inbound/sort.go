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
		return CompareByMatches(a.Matches, b.Matches)
	})
}

// CompareByMatches compares the match keys of two rules. It assumes each rule
// has at most one match — the invariant established by buildRules, which splits
// multi-match rules into one rule per match before sorting.
func CompareByMatches(a, b []common_api.Match) int {
	var ma, mb common_api.Match
	if len(a) > 0 {
		ma = a[0]
	}
	if len(b) > 0 {
		mb = b[0]
	}
	return CompareMatch(ma, mb)
}

func CompareMatch(a, b common_api.Match) int {
	if c := compareSpiffeID(a.SpiffeID, b.SpiffeID); c != 0 {
		return c
	}
	if c := compareSNI(a.SNI, b.SNI); c != 0 {
		return c
	}
	return 0
}

func compareSpiffeID(a, b *common_api.SpiffeIDMatch) int {
	switch {
	case a != nil && b == nil:
		return -1
	case a == nil && b != nil:
		return 1
	case a == nil && b == nil:
		return 0
	}

	score := func(m *common_api.SpiffeIDMatch) int {
		switch m.Type {
		case common_api.ExactMatchType:
			return 2
		case common_api.PrefixMatchType:
			return 1
		default:
			return 0
		}
	}
	return cmp.Compare(score(b), score(a))
}

func compareSNI(a, b *common_api.SNIMatch) int {
	switch {
	case a != nil && b == nil:
		return -1
	case a == nil && b != nil:
		return 1
	}
	return 0
}
