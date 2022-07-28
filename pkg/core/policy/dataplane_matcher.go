package policy

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

// SelectDataplanePolicy given a Dataplane definition and a list of DataplanePolicy returns the "best matching" DataplanePolicy.
// A DataplanePolicy is considered a match if one of the inbound interfaces of a Dataplane or tag section on Gateway Dataplane has all tags of DataplanePolicy's selector.
// Every matching DataplanePolicy gets a rank (score) defined as a maximum number of tags in a matching selector.
// DataplanePolicy with an empty list of selectors is considered a match with a rank (score) of 0.
// DataplanePolicy with an empty selector (one that has no tags) is considered a match with a rank (score) of 0.
// In case if there are multiple DataplanePolicies with the same rank (score), the policy created last is chosen.
func SelectDataplanePolicy(dataplane *mesh.DataplaneResource, policies []DataplanePolicy) DataplanePolicy {
	return SelectDataplanePolicyWithMatcher(dataplane.Spec.Matches, policies)
}

// A TagMatcher test whether a tag selector matches some predicate.
type TagMatcher func(selector mesh_proto.TagSelector) bool

// SelectDataplanePolicyWithMatcher tries to match a DataplanePolicy with the given matches function.
// Every matching DataplanePolicy gets a rank (score) defined as a maximum number of tags in a matching selector.
// DataplanePolicy with an empty list of selectors is considered a match with a rank (score) of 0.
// DataplanePolicy with an empty selector (one that has no tags) is considered a match with a rank (score) of 0.
// In case if there are multiple DataplanePolicies with the same rank (score), the policy created last is chosen.
func SelectDataplanePolicyWithMatcher(matches TagMatcher, policies []DataplanePolicy) DataplanePolicy {
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
			if matches(tagSelector) {
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

// SelectInboundDataplanePolicies given a Dataplane definition and a list of DataplanePolicy returns the "best matching" DataplanePolicy for each inbound separately.
// A DataplanePolicy for an inbound is considered a match if the inbound matches all the tags listed in selector of the DataplanePolicy.
// Every matching DataplanePolicy gets a rank (score) defined as a maximum number of tags in a matching selector.
// DataplanePolicy with an empty list of selectors is considered a match with a rank (score) of 0.
// DataplanePolicy with an empty selector (one that has no tags) is considered a match with a rank (score) of 0.
// In case if there are multiple DataplanePolicies with the same rank (score), the policy created last is chosen.
func SelectInboundDataplanePolicies(dataplane *mesh.DataplaneResource, policies []DataplanePolicy) InboundDataplanePolicyMap {
	sort.Stable(DataplanePolicyByName(policies)) // sort to avoid flakiness

	match := InboundDataplanePolicyMap{}

	for _, inbound := range dataplane.Spec.Networking.GetInbound() {
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
				if inbound.MatchTags(tagSelector) {
					rank := tagSelector.Rank()
					if rank.CompareTo(bestRank) > 0 || sameRankCreatedLater(policy, rank) {
						bestRank = rank
						bestPolicy = policy
					}
				}
			}
		}
		if bestPolicy != nil {
			iface := dataplane.Spec.GetNetworking().ToInboundInterface(inbound)
			match[iface] = bestPolicy
		}
	}

	return match
}

type DataplanePolicyByName []DataplanePolicy

func (a DataplanePolicyByName) Len() int      { return len(a) }
func (a DataplanePolicyByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a DataplanePolicyByName) Less(i, j int) bool {
	return a[i].GetMeta().GetName() < a[j].GetMeta().GetName()
}
