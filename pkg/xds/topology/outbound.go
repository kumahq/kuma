package topology

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

// GetOutboundTargets resolves all endpoints reachable from a given dataplane.
func GetOutboundTargets(destinations core_xds.DestinationMap, dataplanes *mesh_core.DataplaneResourceList) (core_xds.EndpointMap, error) {
	if len(destinations) == 0 {
		return nil, nil
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
		if dataplane.Spec.IsRemoteIngress() && dataplane.Spec.HasAvailableServices() {
			for _, ingress := range dataplane.Spec.Networking.GetIngress().GetAvailableServices() {
				service := ingress.Tags[mesh_proto.ServiceTag]
				selectors, ok := destinations[service]
				if !ok {
					continue
				}
				if !selectors.Matches(ingress.Tags) {
					continue
				}
				outbound[service] = append(outbound[service], core_xds.Endpoint{
					Target: dataplane.Spec.Networking.Address,
					Port:   dataplane.Spec.Networking.Inbound[0].Port,
					Tags:   ingress.Tags,
					Weight: ingress.Instances,
				})
			}
			continue
		}
		for _, inbound := range dataplane.Spec.Networking.GetInbound() {
			service := inbound.Tags[mesh_proto.ServiceTag]
			selectors, ok := destinations[service]
			if !ok {
				continue
			}
			if !selectors.Matches(inbound.Tags) {
				continue
			}
			iface := dataplane.Spec.Networking.ToInboundInterface(inbound)
			// TODO(yskopets): do we need to dedup?
			// TODO(yskopets): sort ?
			outbound[service] = append(outbound[service], core_xds.Endpoint{
				Target: iface.DataplaneIP,
				Port:   iface.DataplanePort,
				Tags:   inbound.Tags,
				Weight: 1,
			})
		}
	}
	return outbound
}
