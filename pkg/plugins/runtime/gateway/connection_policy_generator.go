package gateway

import (
	"sort"

	"github.com/kumahq/kuma/pkg/core/policy"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

func PopulatePolicies(info GatewayListenerInfo, host GatewayHost, routes []route.Entry) []route.Entry {
	var routesWithPolicies []route.Entry

	for _, e := range routes {
		for i, destination := range e.Action.Forward {
			e.Action.Forward[i].Policies = mapPoliciesForDestination(destination.Destination, info, host)
		}
		if e.Mirror != nil {
			e.Mirror.Forward.Policies = mapPoliciesForDestination(e.Mirror.Forward.Destination, info, host)
		}

		routesWithPolicies = append(routesWithPolicies, e)
	}

	return routesWithPolicies
}

func mapPoliciesForDestination(destination envoy.Tags, info GatewayListenerInfo, host GatewayHost) map[model.ResourceType]model.Resource {
	policies := map[model.ResourceType]model.Resource{}

	for _, policyType := range ConnectionPolicyTypes {
		if policy := matchConnectionPolicy(host.Policies[policyType], destination); policy != nil {
			policies[policyType] = policy
		}
	}

	return policies
}

func matchConnectionPolicy(candidates []match.RankedPolicy, destination envoy.Tags) model.Resource {
	var matches []match.RankedPolicy

	for _, c := range candidates {
		if rank, ok := policy.MatchSelector(destination, c.Policy.Destinations()); ok {
			// Track this match with the combined source+destination rank.
			matches = append(matches, match.RankedPolicy{
				Rank:   rank.CombinedWith(c.Rank),
				Policy: c.Policy,
			})
		}
	}

	if len(matches) == 0 {
		return nil
	}

	// Sort more specific (higher ranked) policies first.
	sort.Slice(matches, func(i, j int) bool {
		n := matches[i].Rank.CompareTo(matches[j].Rank)
		switch {
		case n < 0:
			return false
		case n > 0:
			return true
		default /* i == 0 */ :
			// If the rank is the same, the most recent
			// policy sorts to the front (i.e. takes priority).
			return matches[i].Policy.GetMeta().GetCreationTime().After(
				matches[j].Policy.GetMeta().GetCreationTime())
		}
	})

	return matches[0].Policy
}
