package gateway

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/policy"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

// ConnectionPolicyGenerator matches connection policies for each route table
// entry that forwards traffic.
type ConnectionPolicyGenerator struct {
}

func (*ConnectionPolicyGenerator) SupportsProtocol(p mesh_proto.MeshGateway_Listener_Protocol) bool {
	return true
}

func (g *ConnectionPolicyGenerator) GenerateHost(ctx xds_context.Context, info *GatewayListenerInfo, host gatewayHostInfo) (*core_xds.ResourceSet, error) {
	for _, e := range host.RouteTable.Entries {
		for i, destination := range e.Action.Forward {
			e.Action.Forward[i].Policies = mapPoliciesForDestination(destination.Destination, info, host.Host)
		}
		if e.Mirror != nil {
			e.Mirror.Forward.Policies = mapPoliciesForDestination(e.Mirror.Forward.Destination, info, host.Host)
		}
	}

	return nil, nil
}

func mapPoliciesForDestination(destination envoy.Tags, info *GatewayListenerInfo, host GatewayHost) map[model.ResourceType]model.Resource {
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
