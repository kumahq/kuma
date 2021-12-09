package route

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func InferServiceProtocol(endpoints []core_xds.Endpoint) core_mesh.Protocol {
	protocol := generator.InferServiceProtocol(endpoints)

	// HTTP is a better default than "unknown".
	if protocol == core_mesh.ProtocolUnknown {
		return core_mesh.ProtocolHTTP
	}

	return protocol
}

func InferForwardingProtocol(destinations []Destination) core_mesh.Protocol {
	var endpoints []core_xds.Endpoint

	for _, d := range destinations {
		endpoints = append(endpoints, core_xds.Endpoint{Tags: d.Destination})
	}

	return InferServiceProtocol(endpoints)
}
