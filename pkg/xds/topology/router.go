package topology

import (
	"context"

	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// GetRoutes picks a single the most specific route for each outbound interface of a given Dataplane.
func GetRoutes(ctx context.Context, dataplane *core_mesh.DataplaneResource, manager core_manager.ReadOnlyResourceManager) (core_xds.RouteMap, error) {
	if len(dataplane.Spec.Networking.GetOutbound()) == 0 {
		return nil, nil
	}
	routes := &core_mesh.TrafficRouteResourceList{}
	if err := manager.List(ctx, routes, core_store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
		return nil, err
	}
	return BuildRouteMap(dataplane, routes.Items), nil
}

// BuildRouteMap picks a single the most specific route for each outbound interface of a given Dataplane.
func BuildRouteMap(dataplane *core_mesh.DataplaneResource, routes []*core_mesh.TrafficRouteResource) core_xds.RouteMap {
	policies := make([]policy.ConnectionPolicy, len(routes))
	for i, route := range routes {
		policies[i] = route
	}
	policyMap := policy.SelectOutboundConnectionPolicies(dataplane, policies)

	routeMap := core_xds.RouteMap{}
	for _, oface := range dataplane.Spec.Networking.GetOutbounds(mesh_proto.NonBackendRefFilter) {
		serviceName := oface.GetService()
		outbound := dataplane.Spec.Networking.ToOutboundInterface(oface)
		if policy, exists := policyMap[serviceName]; exists {
			routeMap[outbound] = resolveTrafficRouteWildcards(policy.(*core_mesh.TrafficRouteResource), oface.GetTags())
		}
	}
	return routeMap
}

func resolveTrafficRouteWildcards(routeRes *core_mesh.TrafficRouteResource, outboundTags map[string]string) *core_mesh.TrafficRouteResource {
	route := proto.Clone(routeRes.Spec).(*mesh_proto.TrafficRoute) // we need to clone the Spec so we don't override the resource in Cache.
	if len(route.Conf.Destination) > 0 {
		route.Conf.Destination = handleWildcardTagsFor(outboundTags, route.Conf.Destination)
	}
	for _, split := range route.Conf.Split {
		split.Destination = handleWildcardTagsFor(outboundTags, split.Destination)
	}
	for _, http := range route.Conf.Http {
		if len(http.Destination) > 0 {
			http.Destination = handleWildcardTagsFor(outboundTags, http.Destination)
		}
		for _, split := range http.Split {
			split.Destination = handleWildcardTagsFor(outboundTags, split.Destination)
		}
	}

	return &core_mesh.TrafficRouteResource{
		Meta: routeRes.GetMeta(),
		Spec: route,
	}
}

func handleWildcardTagsFor(outboundTags, routeTags map[string]string) map[string]string {
	resultingTags := map[string]string{}

	for k, v := range routeTags {
		if v != mesh_proto.MatchAllTag {
			resultingTags[k] = v
		}
	}

	for k, v := range outboundTags {
		if _, found := resultingTags[k]; !found {
			resultingTags[k] = v
		}
	}

	return resultingTags
}

// BuildDestinationMap creates a map of selectors to match other dataplanes reachable from a given one
// via given routes.
func BuildDestinationMap(dataplane *core_mesh.DataplaneResource, routes core_xds.RouteMap) core_xds.DestinationMap {
	destinations := core_xds.DestinationMap{}
	for _, oface := range dataplane.Spec.Networking.GetOutbounds(mesh_proto.NonBackendRefFilter) {
		serviceName := oface.GetService()
		outbound := dataplane.Spec.Networking.ToOutboundInterface(oface)
		route, ok := routes[outbound]
		if ok {
			for _, destination := range route.Spec.GetConf().GetSplitWithDestination() {
				service, ok := destination.Destination[mesh_proto.ServiceTag]
				if !ok {
					// ignore destinations without a `service` tag
					// TODO(yskopets): consider adding a metric for this
					continue
				}
				destinations[service] = destinations[service].Add(mesh_proto.MatchTags(destination.Destination))
			}
		} else {
			destinations[serviceName] = destinations[serviceName].Add(mesh_proto.MatchService(serviceName))
		}
	}
	return destinations
}
