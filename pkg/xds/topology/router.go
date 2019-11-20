package topology

import (
	"context"
	"sort"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

type pseudoMeta struct {
	Name string
}

func (m *pseudoMeta) GetMesh() string {
	return ""
}
func (m *pseudoMeta) GetNamespace() string {
	return ""
}
func (m *pseudoMeta) GetName() string {
	return m.Name
}
func (m *pseudoMeta) GetVersion() string {
	return ""
}

// GetRoutes picks a single the most specific route for each outbound interface of a given Dataplane.
func GetRoutes(ctx context.Context, dataplane *mesh_core.DataplaneResource, manager core_manager.ResourceManager) (core_xds.RouteMap, error) {
	if len(dataplane.Spec.Networking.GetOutbound()) == 0 {
		return nil, nil
	}
	routes := &mesh_core.TrafficRouteResourceList{}
	if err := manager.List(ctx, routes, core_store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
		return nil, err
	}
	return BuildRouteMap(dataplane, routes.Items), nil
}

// BuildRouteMap picks a single the most specific route for each outbound interface of a given Dataplane.
func BuildRouteMap(dataplane *mesh_core.DataplaneResource, routes []*mesh_core.TrafficRouteResource) core_xds.RouteMap {
	sort.Stable(TrafficRoutesByNamespacedName(routes)) // sort to avoid flakiness

	// First, select only those TrafficRoutes that have a `source` selector matching a given Dataplane.
	// If a TrafficRoute has multiple matching `source` selectors, we need to choose the most specific one.
	// Technically, we give a rank to every matching selector. The more specific selector is, the higher rank it gets.

	type candidateBySource struct {
		route          *mesh_core.TrafficRouteResource
		bestSourceRank mesh_proto.TagSelectorRank
	}

	candidatesBySource := []candidateBySource{}
	for _, route := range routes {
		candidate := candidateBySource{route: route}
		matches := false
		for _, source := range route.Spec.Sources {
			sourceSelector := mesh_proto.TagSelector(source.Match)
			if dataplane.Spec.Matches(sourceSelector) {
				sourceRank := sourceSelector.Rank()
				if !matches || sourceRank.CompareTo(candidate.bestSourceRank) > 0 {
					// TODO(yskopets): use CreationDate to resolve a conflict between 2 equal ranks
					candidate.bestSourceRank = sourceRank
				}
				matches = true
			}
		}
		if matches {
			candidatesBySource = append(candidatesBySource, candidate)
		}
	}

	// Then, for each outbound interface consider all TrafficRoutes that match it by a `destination` selector.
	// If a TrafficRoute has multiple matching `destination` selectors, we need to choose the most specific one.
	//
	// It's possible that there will be multiple TrafficRoutes that match a given outbound interface.
	// To choose between them, we need to compute an aggregate rank of the most specific selector by `source`
	// with the most specific selector by `destination`.

	type candidateByDestination struct {
		candidateBySource
		bestAggregateRank mesh_proto.TagSelectorRank
	}

	candidatesByDestination := map[core_xds.ServiceName]candidateByDestination{}
	for _, oface := range dataplane.Spec.Networking.GetOutbound() {
		if _, ok := candidatesByDestination[oface.Service]; ok {
			// apparently, multiple outbound interfaces of a given Dataplane refer to the same service
			continue
		}
		outboundTags := mesh_proto.SingleValueTagSet{mesh_proto.ServiceTag: oface.Service}
		for _, candidateBySource := range candidatesBySource {
			for _, destination := range candidateBySource.route.Spec.Destinations {
				destinationSelector := mesh_proto.TagSelector(destination.Match)
				if destinationSelector.Matches(outboundTags) {
					aggregateRank := destinationSelector.Rank().CombinedWith(candidateBySource.bestSourceRank)

					candidateByDestination, exists := candidatesByDestination[oface.Service]

					if !exists || aggregateRank.CompareTo(candidateByDestination.bestAggregateRank) > 0 {
						// TODO(yskopets): use CreationDate to resolve a conflict between 2 equal ranks
						candidateByDestination.candidateBySource = candidateBySource
						candidateByDestination.bestAggregateRank = aggregateRank

						candidatesByDestination[oface.Service] = candidateByDestination
					}
				}
			}
		}
	}

	routeMap := core_xds.RouteMap{}
	for _, oface := range dataplane.Spec.Networking.GetOutbound() {
		candidate, exists := candidatesByDestination[oface.Service]
		if exists {
			routeMap[oface.Service] = candidate.route
		} else {
			// to avoid scattering defaulting logic everywhere, let's create a pseudo TrafficRoute
			routeMap[oface.Service] = &mesh_core.TrafficRouteResource{
				Meta: &pseudoMeta{
					Name: "(implicit default route)",
				},
				Spec: mesh_proto.TrafficRoute{
					Sources: []*mesh_proto.Selector{{
						Match: mesh_proto.MatchAnyService(),
					}},
					Destinations: []*mesh_proto.Selector{{
						Match: mesh_proto.MatchService(oface.Service),
					}},
					Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
						Weight:      100,
						Destination: mesh_proto.MatchService(oface.Service),
					}},
				},
			}
		}
	}
	return routeMap
}

// BuildDestinationMap creates a map of selectors to match other dataplanes reachable from a given one
// via given routes.
func BuildDestinationMap(dataplane *mesh_core.DataplaneResource, routes core_xds.RouteMap) core_xds.DestinationMap {
	destinations := core_xds.DestinationMap{}
	for _, oface := range dataplane.Spec.Networking.GetOutbound() {
		route, ok := routes[oface.Service]
		if ok {
			for _, destination := range route.Spec.Conf {
				service, ok := destination.Destination[mesh_proto.ServiceTag]
				if !ok {
					// ignore destinations without a `service` tag
					// TODO(yskopets): consider adding a metric for this
					continue
				}
				destinations[service] = destinations[service].Add(mesh_proto.MatchTags(destination.Destination))
			}
		} else {
			destinations[oface.Service] = destinations[oface.Service].Add(mesh_proto.MatchService(oface.Service))
		}
	}
	return destinations
}

// todo fix namespace
type TrafficRoutesByNamespacedName []*mesh_core.TrafficRouteResource

func (a TrafficRoutesByNamespacedName) Len() int      { return len(a) }
func (a TrafficRoutesByNamespacedName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a TrafficRoutesByNamespacedName) Less(i, j int) bool {
	return a[i].Meta.GetName() < a[j].Meta.GetName()
}
