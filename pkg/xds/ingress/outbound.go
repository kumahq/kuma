package ingress

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

func BuildEndpointMap(destinations core_xds.DestinationMap, dataplanes []*core_mesh.DataplaneResource) core_xds.EndpointMap {
	if len(destinations) == 0 {
		return nil
	}
	outbound := core_xds.EndpointMap{}
	for _, dataplane := range dataplanes {
		for _, inbound := range dataplane.Spec.GetNetworking().GetHealthyInbounds() {
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
