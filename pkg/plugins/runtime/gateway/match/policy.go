package match

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/policy"
	"github.com/kumahq/kuma/pkg/core/resources/model"
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
