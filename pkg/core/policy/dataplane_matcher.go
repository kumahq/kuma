package policy

import (
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"sort"
)

// FindBestMatch given a Dataplane definition and a list of DataplanePolicy returns the "best matching" DataplanePolicy.
// A DataplanePolicy is considered a match if one of the inbound interfaces of a Dataplane has all tags of DataplanePolicy's selector.
// Every matching DataplanePolicy gets a rank (score) defined as a maximum number of tags in a matching selector.
// DataplanePolicy with an empty list of selectors is considered a match with a rank (score) of 0.
// DataplanePolicy with an empty selector (one that has no tags) is considered a match with a rank (score) of 0.
// In case if there are multiple DataplanePolicies with the same rank (score), policies are sorted alphabetically by Name
// and the first one is considered the "best match".
func FindBestMatch(dataplane *mesh.DataplaneResource, policies []DataplanePolicy) DataplanePolicy {
	sort.Stable(DataplanePolicyByName(policies)) // sort to avoid flakiness

	var bestMatch DataplanePolicy
	var bestScore int
	for _, policy := range policies {
		if 0 == len(policy.Selectors()) { // match everything
			if bestMatch == nil {
				bestMatch = policy
			}
			continue
		}
		for _, selector := range policy.Selectors() {
			if 0 == len(selector.Match) { // match everything
				if bestMatch == nil {
					bestMatch = policy
				}
				continue
			}
			for _, inbound := range dataplane.Spec.Networking.GetInbound() {
				if matches, score := ScoreMatch(selector.Match, inbound.Tags); matches && bestScore < score {
					bestMatch = policy
					bestScore = score
				}
			}
		}
	}
	return bestMatch
}

func ScoreMatch(selector map[string]string, target map[string]string) (bool, int) {
	for key, requiredValue := range selector {
		if actualValue, hasKey := target[key]; !hasKey || actualValue != requiredValue {
			return false, 0
		}
	}
	return true, len(selector)
}

type DataplanePolicyByName []DataplanePolicy

func (a DataplanePolicyByName) Len() int      { return len(a) }
func (a DataplanePolicyByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a DataplanePolicyByName) Less(i, j int) bool {
	return a[i].GetMeta().GetName() < a[j].GetMeta().GetName()
}
