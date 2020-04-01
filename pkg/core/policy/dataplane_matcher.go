package policy

import (
	"sort"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
)

type MatchingTags interface {
	MatchTags(selector mesh_proto.TagSelector) bool
}

// SelectDataplanePolicy given a something that could be matched (mesh_proto.Dataplane,
// mesh_proto.Dataplane_Networking_Inbound, etc.) and a list of DataplanePolicy returns the "best matching" DataplanePolicy.
// Every matching DataplanePolicy gets a rank (score) defined as a maximum number of tags in a matching selector.
// DataplanePolicy with an empty list of selectors is considered a match with a rank (score) of 0.
// DataplanePolicy with an empty selector (one that has no tags) is considered a match with a rank (score) of 0.
// In case if there are multiple DataplanePolicies with the same rank (score), the policy created last is chosen.
func SelectDataplanePolicy(matching MatchingTags, policies []DataplanePolicy) DataplanePolicy {
	sort.Stable(DataplanePolicyByName(policies)) // sort to avoid flakiness

	var bestPolicy DataplanePolicy
	var bestRank mesh_proto.TagSelectorRank
	sameRankCreatedLater := func(policy DataplanePolicy, rank mesh_proto.TagSelectorRank) bool {
		return rank.CompareTo(bestRank) == 0 && policy.GetMeta().GetCreationTime().After(bestPolicy.GetMeta().GetCreationTime())
	}

	for _, policy := range policies {
		if 0 == len(policy.Selectors()) { // match everything
			if bestPolicy == nil || sameRankCreatedLater(policy, mesh_proto.TagSelectorRank{}) {
				bestPolicy = policy
			}
			continue
		}
		for _, selector := range policy.Selectors() {
			if 0 == len(selector.Match) { // match everything
				if bestPolicy == nil || sameRankCreatedLater(policy, mesh_proto.TagSelectorRank{}) {
					bestPolicy = policy
				}
				continue
			}
			tagSelector := mesh_proto.TagSelector(selector.Match)
			if matching.MatchTags(tagSelector) {
				rank := tagSelector.Rank()
				if rank.CompareTo(bestRank) > 0 || sameRankCreatedLater(policy, rank) {
					bestRank = rank
					bestPolicy = policy
				}
			}
		}
	}
	return bestPolicy
}

type DataplanePolicyByName []DataplanePolicy

func (a DataplanePolicyByName) Len() int      { return len(a) }
func (a DataplanePolicyByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a DataplanePolicyByName) Less(i, j int) bool {
	return a[i].GetMeta().GetName() < a[j].GetMeta().GetName()
}
