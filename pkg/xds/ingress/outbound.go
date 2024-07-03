package ingress

import (
	"maps"

	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
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
	meshServices []*meshservice_api.MeshServiceResource,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}
	for _, dataplane := range dataplanes {
		for _, inbound := range dataplane.Spec.GetNetworking().GetHealthyInbounds() {
			service := inbound.Tags[mesh_proto.ServiceTag]
			selectors, ok := destinations[service]
			if !ok {
				continue
			}
			if !selectors.Matches(inbound.Tags) {
				continue
			}
			iface := dataplane.Spec.GetNetworking().ToInboundInterface(inbound)
			outbound[service] = append(outbound[service], core_xds.Endpoint{
				Target: iface.DataplaneAdvertisedIP,
				Port:   iface.DataplanePort,
				Tags:   inbound.Tags,
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
					),
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
		if !selectors.Matches(es.Spec.GetTags()) {
			continue
		}

		for _, zoneEgress := range zoneEgresses {
			iface := zoneEgress.Spec.GetNetworking()
			outbound[service] = append(outbound[service], core_xds.Endpoint{
				Target:          iface.Address,
				Port:            iface.Port,
				Tags:            es.Spec.GetTags(),
				Weight:          1,
				ExternalService: &core_xds.ExternalService{},
			})
		}
	}

	// O(dataplane*meshsvc) can be optimized by sharding both by namespace
	// this is copy of topology/outbound.go
	for _, dataplane := range dataplanes {
		dpNetworking := dataplane.Spec.GetNetworking()

		for _, meshSvc := range meshServices {
			tagSelector := mesh_proto.TagSelector(meshSvc.Spec.Selector.DataplaneTags)
			for _, inbound := range dpNetworking.GetHealthyInbounds() {
				if !tagSelector.Matches(inbound.GetTags()) {
					continue
				}
				for _, port := range meshSvc.Spec.Ports {
					switch port.TargetPort.Type {
					case intstr.Int:
						if uint32(port.TargetPort.IntVal) != inbound.Port {
							continue
						}
					case intstr.String:
						if port.TargetPort.StrVal != inbound.Name {
							continue
						}
					}

					inboundTags := maps.Clone(inbound.GetTags())
					serviceName := meshSvc.DestinationName(port.Port)
					if serviceName == inboundTags[mesh_proto.ServiceTag] {
						continue // it was already added by fillDataplaneOutbounds
					}
					inboundInterface := dpNetworking.ToInboundInterface(inbound)

					outbound[serviceName] = append(outbound[serviceName], core_xds.Endpoint{
						Target: inboundInterface.DataplaneAdvertisedIP,
						Port:   inboundInterface.DataplanePort,
						Tags:   inboundTags,
						Weight: 1,
					})
				}
			}
		}
	}

	return outbound
}
