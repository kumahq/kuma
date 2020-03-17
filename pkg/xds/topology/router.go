package topology

import (
	"context"
	"time"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/policy"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

type pseudoMeta struct {
	Name string
}

func (m *pseudoMeta) GetMesh() string {
	return ""
}
func (m *pseudoMeta) GetName() string {
	return m.Name
}
func (m *pseudoMeta) GetNameExtensions() core_model.ResourceNameExtensions {
	return core_model.ResourceNameExtensionsUnsupported
}
func (m *pseudoMeta) GetVersion() string {
	return ""
}
func (m *pseudoMeta) GetCreationTime() time.Time {
	return time.Now()
}
func (m *pseudoMeta) GetModificationTime() time.Time {
	return time.Now()
}

// GetRoutes picks a single the most specific route for each outbound interface of a given Dataplane.
func GetRoutes(ctx context.Context, dataplane *mesh_core.DataplaneResource, manager core_manager.ReadOnlyResourceManager) (core_xds.RouteMap, error) {
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
	policies := make([]policy.ConnectionPolicy, len(routes))
	for i, route := range routes {
		policies[i] = route
	}
	policyMap := policy.SelectOutboundConnectionPolicies(dataplane, policies)

	routeMap := core_xds.RouteMap{}
	for _, oface := range dataplane.Spec.Networking.GetOutbound() {
		policy, exists := policyMap[oface.Service]
		if exists {
			routeMap[oface.Service] = policy.(*mesh_core.TrafficRouteResource)
		} else {
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
