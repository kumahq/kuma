package policy

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

type ServiceIterator interface {
	Next() (core_xds.ServiceName, bool)
}

type ServiceIteratorFunc func() (core_xds.ServiceName, bool)

func (f ServiceIteratorFunc) Next() (core_xds.ServiceName, bool) {
	return f()
}

func ToOutboundServicesOf(dataplane *core_mesh.DataplaneResource) ServiceIterator {
	idx := 0
	outbounds := dataplane.Spec.Networking.GetOutbounds(mesh_proto.NonBackendRefFilter)
	return ServiceIteratorFunc(func() (core_xds.ServiceName, bool) {
		if len(outbounds) < idx {
			return "", false
		}
		if len(outbounds) == idx { // add additional implicit pass through service
			idx++
			return core_meta.PassThroughServiceName, true
		}
		oface := outbounds[idx]
		idx++
		return oface.GetService(), true
	})
}

func ToServicesOf(destinations core_xds.DestinationMap) ServiceIterator {
	services := make([]core_xds.ServiceName, 0, len(destinations))
	for service := range destinations {
		services = append(services, service)
	}
	return ToServices(services)
}

func ToServices(services []core_xds.ServiceName) ServiceIterator {
	idx := 0
	return ServiceIteratorFunc(func() (core_xds.ServiceName, bool) {
		if len(services) <= idx {
			return "", false
		}
		service := services[idx]
		idx++
		return service, true
	})
}

// SelectOutboundConnectionPolicies picks a single the most specific policy for each outbound interface of a given Dataplane.
func SelectOutboundConnectionPolicies(dataplane *core_mesh.DataplaneResource, policies []ConnectionPolicy) OutboundConnectionPolicyMap {
	return SelectConnectionPolicies(dataplane, ToOutboundServicesOf(dataplane), policies)
}

// SelectConnectionPolicies picks a single the most specific policy applicable to a connection between a given dataplane and given destination services.
func SelectConnectionPolicies(dataplane *core_mesh.DataplaneResource, destinations ServiceIterator, policies []ConnectionPolicy) OutboundConnectionPolicyMap {
	sort.Stable(ConnectionPolicyByName(policies)) // sort to avoid flakiness

	// First, select only those ConnectionPolicies that have a `source` selector matching a given Dataplane.
	// If a ConnectionPolicy has multiple matching `source` selectors, we need to choose the most specific one.
	// Technically, we give a rank to every matching selector. The more specific selector is, the higher rank it gets.

	type candidateBySource struct {
		policy         ConnectionPolicy
		bestSourceRank mesh_proto.TagSelectorRank
	}

	candidatesBySource := []candidateBySource{}
	for _, policy := range policies {
		candidate := candidateBySource{policy: policy}
		matches := false
		for _, source := range policy.Sources() {
			sourceSelector := mesh_proto.TagSelector(source.Match)
			if dataplane.Spec.Matches(sourceSelector) {
				sourceRank := sourceSelector.Rank()
				if !matches || sourceRank.CompareTo(candidate.bestSourceRank) > 0 {
					candidate.bestSourceRank = sourceRank
				}
				matches = true
			}
		}
		if matches {
			candidatesBySource = append(candidatesBySource, candidate)
		}
	}

	// Then, for each outbound interface consider all ConnectionPolicies that match it by a `destination` selector.
	// If a ConnectionPolicy has multiple matching `destination` selectors, we need to choose the most specific one.
	//
	// It's possible that there will be multiple ConnectionPolicies that match a given outbound interface.
	// To choose between them, we need to compute an aggregate rank of the most specific selector by `source`
	// with the most specific selector by `destination`.

	type candidateByDestination struct {
		candidateBySource
		bestAggregateRank mesh_proto.TagSelectorRank
	}

	candidatesByDestination := map[core_xds.ServiceName]candidateByDestination{}
	for service, ok := destinations.Next(); ok; service, ok = destinations.Next() {
		if _, ok := candidatesByDestination[service]; ok {
			// apparently, multiple outbound interfaces of a given Dataplane refer to the same service
			continue
		}
		outboundTags := mesh_proto.SingleValueTagSet{mesh_proto.ServiceTag: service}
		for _, candidateBySource := range candidatesBySource {
			for _, destination := range candidateBySource.policy.Destinations() {
				destinationSelector := mesh_proto.TagSelector(destination.Match)
				if destinationSelector.Matches(outboundTags) {
					aggregateRank := destinationSelector.Rank().CombinedWith(candidateBySource.bestSourceRank)

					candidateByDestination, exists := candidatesByDestination[service]

					if !exists ||
						aggregateRank.CompareTo(candidateByDestination.bestAggregateRank) > 0 ||
						(aggregateRank.CompareTo(candidateByDestination.bestAggregateRank) == 0 && candidateBySource.policy.GetMeta().GetCreationTime().After(candidateByDestination.policy.GetMeta().GetCreationTime())) {
						candidateByDestination.candidateBySource = candidateBySource
						candidateByDestination.bestAggregateRank = aggregateRank

						candidatesByDestination[service] = candidateByDestination
					}
				}
			}
		}
	}

	policyMap := OutboundConnectionPolicyMap{}
	for service, candidate := range candidatesByDestination {
		policyMap[service] = candidate.policy
	}
	return policyMap
}

// SelectInboundConnectionPolicies picks a single the most specific policy for each inbound interface of a given Dataplane.
// For each inbound we pick a policy that matches the most destination tags with inbound tags
// Sources part of matched policies are later used in Envoy config to apply it only for connection that matches sources
func SelectInboundConnectionPolicies(dataplane *core_mesh.DataplaneResource, inbounds []*mesh_proto.Dataplane_Networking_Inbound, policies []ConnectionPolicy) InboundConnectionPolicyMap {
	sort.Stable(ConnectionPolicyByName(policies)) // sort to avoid flakiness
	policiesMap := make(InboundConnectionPolicyMap)
	for _, inbound := range inbounds {
		if bestPolicy := SelectInboundConnectionPolicy(inbound.Tags, policies); bestPolicy != nil {
			iface := dataplane.Spec.GetNetworking().ToInboundInterface(inbound)
			policiesMap[iface] = bestPolicy
		}
	}

	return policiesMap
}

// SelectInboundConnectionMatchingPolicies picks all matching policies for each inbound interface of a given Dataplane.
func SelectInboundConnectionMatchingPolicies(dataplane *core_mesh.DataplaneResource, inbounds []*mesh_proto.Dataplane_Networking_Inbound, policies []ConnectionPolicy) InboundConnectionPoliciesMap {
	sort.Stable(ConnectionPolicyByName(policies)) // sort to avoid flakiness
	policiesMap := make(InboundConnectionPoliciesMap)
	for _, inbound := range inbounds {
		if matchnigPolicies := SelectInboundConnectionAllPolicies(inbound.Tags, policies); matchnigPolicies != nil {
			iface := dataplane.Spec.GetNetworking().ToInboundInterface(inbound)
			policiesMap[iface] = matchnigPolicies
		}
	}

	return policiesMap
}

// SelectInboundConnectionPolicy picks a single the most specific policy for given inbound tags.
func SelectInboundConnectionPolicy(inboundTags map[string]string, policies []ConnectionPolicy) ConnectionPolicy {
	var bestPolicy ConnectionPolicy
	var bestRank mesh_proto.TagSelectorRank
	sameRankCreatedLater := func(policy ConnectionPolicy, rank mesh_proto.TagSelectorRank) bool {
		return rank.CompareTo(bestRank) == 0 && policy.GetMeta().GetCreationTime().After(bestPolicy.GetMeta().GetCreationTime())
	}

	for _, policy := range policies {
		for _, selector := range policy.Destinations() {
			tagSelector := mesh_proto.TagSelector(selector.Match)
			if tagSelector.Matches(inboundTags) {
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

// SelectInboundConnectionAllPolicies picks polices for given inbound tags.
func SelectInboundConnectionAllPolicies(inboundTags map[string]string, policies []ConnectionPolicy) []ConnectionPolicy {
	matchingPolicies := []ConnectionPolicy{}

	for _, policy := range policies {
		for _, selector := range policy.Destinations() {
			tagSelector := mesh_proto.TagSelector(selector.Match)
			if tagSelector.Matches(inboundTags) {
				matchingPolicies = append(matchingPolicies, policy)
			}
		}
	}

	// Make sure more specific policies get top priority
	sort.Stable(ConnectionPolicyBySourceRank(matchingPolicies))
	return matchingPolicies
}

type ConnectionPolicyByName []ConnectionPolicy

func (a ConnectionPolicyByName) Len() int      { return len(a) }
func (a ConnectionPolicyByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ConnectionPolicyByName) Less(i, j int) bool {
	return a[i].GetMeta().GetName() < a[j].GetMeta().GetName()
}

type ConnectionPolicyBySourceRank []ConnectionPolicy

func (a ConnectionPolicyBySourceRank) Len() int      { return len(a) }
func (a ConnectionPolicyBySourceRank) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ConnectionPolicyBySourceRank) Less(i, j int) bool {
	tagSelectorI := mesh_proto.TagSelector(a[i].Sources()[0].Match)
	tagSelectorJ := mesh_proto.TagSelector(a[j].Sources()[0].Match)

	tagComparison := tagSelectorI.Rank().CompareTo(tagSelectorJ.Rank())

	if tagComparison == 0 {
		return a[i].GetMeta().GetCreationTime().After(a[j].GetMeta().GetCreationTime())
	}

	return tagComparison > 0
}
