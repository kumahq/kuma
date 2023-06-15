package ingress

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

func BuildEndpointMap(
	destinations core_xds.DestinationMap,
	dataplanes []*core_mesh.DataplaneResource,
	externalServices []*core_mesh.ExternalServiceResource,
	zoneEgresses []*core_mesh.ZoneEgressResource,
	gateways []*core_mesh.MeshGatewayResource,
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
			withMesh := tags.Tags(inbound.Tags).WithTags("mesh", dataplane.GetMeta().GetMesh())
			if !selectors.Matches(withMesh) {
				continue
			}
			iface := dataplane.Spec.GetNetworking().ToInboundInterface(inbound)
			outbound[service] = append(outbound[service], core_xds.Endpoint{
				Target: iface.DataplaneIP,
				Port:   iface.DataplanePort,
				Tags:   withMesh,
				Weight: 1,
			})
		}
		if dataplane.Spec.IsBuiltinGateway() {
			gateway := topology.SelectGateway(gateways, dataplane.Spec.Matches)
			if gateway == nil {
				continue
			}

			dpSpec := dataplane.Spec
			dpNetworking := dpSpec.GetNetworking()

			dpGateway := dpNetworking.GetGateway()
			dpTags := dpGateway.GetTags()
			serviceName := dpTags[mesh_proto.ServiceTag]

			for _, listener := range gateway.Spec.GetConf().GetListeners() {
				if !listener.CrossMesh {
					continue
				}
				outbound[serviceName] = append(outbound[serviceName], core_xds.Endpoint{
					Target: dpNetworking.GetAddress(),
					Port:   listener.GetPort(),
					Tags: mesh_proto.Merge[tags.Tags](
						dpTags, gateway.Spec.GetTags(), listener.GetTags(),
					).WithTags("mesh", dataplane.GetMeta().GetMesh()),
					Weight: 1,
				})
			}
		}
	}

	for _, es := range externalServices {
		service := es.Spec.GetTags()[mesh_proto.ServiceTag]
		selectors, ok := destinations[service]
		if !ok {
			continue
		}
		withMesh := tags.Tags(es.Spec.GetTags()).WithTags("mesh", es.GetMeta().GetMesh())
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
