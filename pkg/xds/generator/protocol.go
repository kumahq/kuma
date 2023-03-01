package generator

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// protocolStack is a mapping between a protocol and its full protocol stack, e.g.
// HTTP has a protocol stack [HTTP, TCP],
// GRPC has a protocol stack [GRPC, HTTP2, TCP],
// TCP  has a protocol stack [TCP].
var protocolStacks = map[core_mesh.Protocol]core_mesh.ProtocolList{
	core_mesh.ProtocolGRPC:  {core_mesh.ProtocolGRPC, core_mesh.ProtocolHTTP2, core_mesh.ProtocolTCP},
	core_mesh.ProtocolHTTP2: {core_mesh.ProtocolHTTP2, core_mesh.ProtocolTCP},
	core_mesh.ProtocolHTTP:  {core_mesh.ProtocolHTTP, core_mesh.ProtocolTCP},
	core_mesh.ProtocolKafka: {core_mesh.ProtocolKafka, core_mesh.ProtocolTCP},
	core_mesh.ProtocolTCP:   {core_mesh.ProtocolTCP},
}

// getCommonProtocol returns a common protocol between given two.
//
// E.g.,
// a common protocol between HTTP and HTTP2 is HTTP2,
// a common protocol between HTTP and HTTP  is HTTP,
// a common protocol between HTTP and TCP   is TCP,
// a common protocol between GRPC and HTTP2 is HTTP2,
// a common protocol between HTTP and HTTP2 is HTTP.
func getCommonProtocol(one, another core_mesh.Protocol) core_mesh.Protocol {
	if one == another {
		return one
	}
	if one == core_mesh.ProtocolUnknown || another == core_mesh.ProtocolUnknown {
		return core_mesh.ProtocolUnknown
	}
	oneProtocolStack, exist := protocolStacks[one]
	if !exist {
		return core_mesh.ProtocolUnknown
	}
	anotherProtocolStack, exist := protocolStacks[another]
	if !exist {
		return core_mesh.ProtocolUnknown
	}
	for _, firstProtocol := range oneProtocolStack {
		for _, secondProtocol := range anotherProtocolStack {
			if firstProtocol == secondProtocol {
				return firstProtocol
			}
		}
	}
	return core_mesh.ProtocolUnknown
}

// InferServiceProtocol returns a common protocol for a given group of endpoints.
func InferServiceProtocol(endpoints []core_xds.Endpoint) core_mesh.Protocol {
	if len(endpoints) == 0 {
		return core_mesh.ProtocolUnknown
	}
	serviceProtocol := core_mesh.ParseProtocol(endpoints[0].Tags[mesh_proto.ProtocolTag])
	for _, endpoint := range endpoints[1:] {
		endpointProtocol := core_mesh.ParseProtocol(endpoint.Tags[mesh_proto.ProtocolTag])
		serviceProtocol = getCommonProtocol(serviceProtocol, endpointProtocol)
	}
	return serviceProtocol
}
