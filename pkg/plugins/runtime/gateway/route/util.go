package route

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func InferServiceProtocol(endpoints []core_xds.Endpoint, routeProtocol core_mesh.Protocol) core_mesh.Protocol {
	protocol := generator.InferServiceProtocol(endpoints)

	if protocol == core_mesh.ProtocolUnknown {
		switch routeProtocol {
		case core_mesh.ProtocolHTTP:
			return core_mesh.ProtocolHTTP
		case core_mesh.ProtocolTCP:
			return core_mesh.ProtocolTCP
		default:
			// HTTP is a better default than "unknown".
			return core_mesh.ProtocolHTTP
		}
	}

	return protocol
}

func InferForwardingProtocol(destinations []Destination) core_mesh.Protocol {
	var endpoints []core_xds.Endpoint

	for _, d := range destinations {
		endpoints = append(endpoints, core_xds.Endpoint{Tags: d.Destination})
	}

	return InferServiceProtocol(endpoints, core_mesh.ProtocolHTTP)
}

func HasExternalServiceEndpoint(mesh *core_mesh.MeshResource, endpoints core_xds.EndpointMap, d Destination) bool {
	service := d.Destination[mesh_proto.ServiceTag]

	var firstEndpointExternalService bool
	if endpoints := endpoints[service]; len(endpoints) > 0 {
		firstEndpointExternalService = endpoints[0].IsExternalService()
	}

	// If there is Mesh property ZoneEgress enabled we want always to
	// direct the traffic through them. The condition is, the mesh must
	// have mTLS enabled and traffic through zoneEgress is enabled.
	return firstEndpointExternalService
}

// FilterProtocols returns only the routes a listener of this protocol knows how to configure.
func FilterProtocols(routes []*core_mesh.MeshGatewayRouteResource, protocol mesh_proto.MeshGateway_Listener_Protocol) []*core_mesh.MeshGatewayRouteResource {
	var filtered []*core_mesh.MeshGatewayRouteResource

	for _, route := range routes {
		switch t := route.Spec.GetConf().GetRoute().(type) {
		case *mesh_proto.MeshGatewayRoute_Conf_Http:
			if protocol == mesh_proto.MeshGateway_Listener_HTTP ||
				protocol == mesh_proto.MeshGateway_Listener_HTTPS {
				filtered = append(filtered, route)
			}
		case *mesh_proto.MeshGatewayRoute_Conf_Tcp:
			if protocol == mesh_proto.MeshGateway_Listener_TCP {
				filtered = append(filtered, route)
			}
		default:
			panic(fmt.Sprintf("Route type %T unimplemented", t))
		}
	}

	return filtered
}
