package match

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/policy"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
)

// ToConnectionPolicies casts a ResourceList to a slice of ConnectionPolicy.
func ToConnectionPolicies(policies model.ResourceList) []policy.ConnectionPolicy {
	items := policies.GetItems()
	c := make([]policy.ConnectionPolicy, 0, len(items))

	for _, i := range items {
		c = append(c, i.(policy.ConnectionPolicy))
	}

	return c
}

// RankedPolicy is a policy that matches some set of tags, together
// with the rank of the match.
type RankedPolicy struct {
	Rank   mesh_proto.TagSelectorRank
	Policy policy.ConnectionPolicy
}

// ConnectionPoliciesBySource finds all the connection policies that have a
// matching `Sources` selector. The resulting matches are not ordered.
func ConnectionPoliciesBySource(
	sourceTags map[string]string,
	policies []policy.ConnectionPolicy,
) []RankedPolicy {
	var matches []RankedPolicy

	for _, p := range policies {
		if rank, ok := policy.MatchSelector(sourceTags, p.Sources()); ok {
			matches = append(matches, RankedPolicy{rank, p})
		}
	}

	return matches
}

// OldestPolicy returns the resource that has the earliest creation time.
func OldestPolicy(policies []model.Resource) model.Resource {
	if len(policies) == 0 {
		return nil
	}

	// Copy to avoid reordering the input argument.
	sorted := make([]model.Resource, len(policies))
	copy(sorted, policies)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetMeta().GetCreationTime().Before(sorted[j].GetMeta().GetCreationTime())
	})

	return sorted[0]
}

// BestConnectionPolicyForDestination returns the retry policy for the collection
// of forwarding target. This is conceptually a bit subtle because a
// forwarding target can have multiple destinations, each of which is a
// distinct service.  However, there are some relatively obvious rules that
// we can use to determine policy.
//
// 1. If all the destinations are the same service, use that policy.
// 2. If there are multiple destinations, prefer a wildcard policy.
// 3. Everything else being equal, older policies are preferred.
func BestConnectionPolicyForDestination(
	destinations []route.Destination,
	policyType model.ResourceType,
) model.Resource {
	seenNames := map[string]bool{}
	servicePolicies := map[string][]model.Resource{}

	// Index all the policies by service name.
	for _, d := range destinations {
		p, ok := d.Policies[policyType]
		if !ok {
			continue
		}

		if seenNames[p.GetMeta().GetName()] {
			continue
		}

		// Index this policy by its destination service.
		c := p.(policy.ConnectionPolicy)
		for _, selector := range c.Destinations() {
			svc := selector.GetMatch()[mesh_proto.ServiceTag]
			if svc == d.Destination[mesh_proto.ServiceTag] || svc == mesh_proto.MatchAllTag {
				servicePolicies[svc] = append(servicePolicies[svc], p)
			}
		}

		seenNames[p.GetMeta().GetName()] = true
	}

	var candidates []model.Resource

	// If we are forwarding to multiple services, no one service
	// would be the most specific match, so we should choose the
	// wildcard policy. Otherwise, we can just take the oldest of
	// all the matches, since there's no better way to discriminate.
	candidates = append(candidates, servicePolicies[mesh_proto.MatchAllTag]...)
	if len(candidates) == 0 {
		for _, p := range servicePolicies {
			candidates = append(candidates, p...)
		}
	}

	return OldestPolicy(candidates)
}
