package topology

import (
	"context"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

// GetOutboundTargets resolves all endpoints reachable from a given dataplane.
func GetOutboundTargets(ctx context.Context, dataplane *mesh_core.DataplaneResource, destinations core_xds.DestinationMap, manager core_manager.ReadOnlyResourceManager) (core_xds.EndpointMap, error) {
	if len(destinations) == 0 {
		return nil, nil
	}
	dataplanes := &mesh_core.DataplaneResourceList{}
	if err := manager.List(ctx, dataplanes, core_store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
		return nil, err
	}
	return BuildEndpointMap(destinations, dataplanes.Items), nil
}

// BuildEndpointMap creates a map of all endpoints that match given selectors.
func BuildEndpointMap(destinations core_xds.DestinationMap, dataplanes []*mesh_core.DataplaneResource) core_xds.EndpointMap {
	if len(destinations) == 0 {
		return nil
	}
	outbound := core_xds.EndpointMap{}
	for _, dataplane := range dataplanes {
		for i, inbound := range dataplane.Spec.Networking.GetInbound() {
			service := inbound.Tags[mesh_proto.ServiceTag]
			selectors, ok := destinations[service]
			if !ok {
				continue
			}
			matches := false
			for _, selector := range selectors {
				if selector.Matches(inbound.Tags) {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}
			iface, err := dataplane.Spec.Networking.GetInboundInterfaceByIdx(i)
			if err != nil {
				// skip dataplanes with invalid configuration
				continue
			}
			// TODO(yskopets): do we need to dedup?
			// TODO(yskopets): sort ?
			outbound[service] = append(outbound[service], core_xds.Endpoint{
				Target: iface.DataplaneIP,
				Port:   iface.DataplanePort,
				Tags:   inbound.Tags,
			})
		}
	}
	return outbound
}
