package ingress

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

func BuildEndpointMap(
	destinations core_xds.DestinationMap,
	dataplanes []*core_mesh.DataplaneResource,
	externalServices []*core_mesh.ExternalServiceResource,
	zoneEgresses []*core_mesh.ZoneEgressResource,
) core_xds.EndpointMap {
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

	for _, es := range externalServices {
		service := es.Spec.Tags[mesh_proto.ServiceTag]
		selectors, ok := destinations[service]
		if !ok {
			continue
		}
		withMesh := envoy.Tags(es.Spec.Tags).WithTags("mesh", es.GetMeta().GetMesh())
		if !selectors.Matches(withMesh) {
			continue
		}

		for _, zoneEgress := range zoneEgresses {
			iface := zoneEgress.Spec.GetNetworking()
			outbound[service] = append(outbound[service], core_xds.Endpoint{
				Target:          iface.Address,
				Port:            iface.Port,
				Tags:            withMesh,
				Weight:          1,
				ExternalService: &core_xds.ExternalService{},
			})
		}
	}

	return outbound
}
