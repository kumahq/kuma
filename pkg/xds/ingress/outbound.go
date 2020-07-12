package ingress

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/xds/envoy"
)

func BuildEndpointMap(destinations core_xds.DestinationMap, dataplanes []*mesh_core.DataplaneResource) core_xds.EndpointMap {
	if len(destinations) == 0 {
		return nil
	}
	outbound := core_xds.EndpointMap{}
	for _, dataplane := range dataplanes {
		if dataplane.Spec.IsIngress() {
			continue
		}
		for _, inbound := range dataplane.Spec.Networking.GetInbound() {
			service := inbound.Tags[mesh_proto.ServiceTag]
			selectors, ok := destinations[service]
			if !ok {
				continue
			}
			withMesh := envoy.Tags(inbound.Tags).WithTags("mesh", dataplane.GetMeta().GetMesh())
			if !selectors.Matches(withMesh) {
				continue
			}
			iface := dataplane.Spec.Networking.ToInboundInterface(inbound)
			outbound[service] = append(outbound[service], core_xds.Endpoint{
				Target: iface.DataplaneIP,
				Port:   iface.DataplanePort,
				Tags:   withMesh,
				Weight: 1,
			})
		}
	}
	return outbound
}
