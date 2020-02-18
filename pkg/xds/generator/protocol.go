package generator

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

var (
	// protocolStack is a mapping between a protocol and its full protocol stack, e.g.
	// HTTP has a protocol stack [HTTP, TCP],
	// GRPC has a protocol stack [GRPC, HTTP2, TCP],
	// TCP  has a protocol stack [TCP].
	protocolStacks = map[mesh_core.Protocol]mesh_core.ProtocolList{
		mesh_core.ProtocolHTTP: mesh_core.ProtocolList{mesh_core.ProtocolHTTP, mesh_core.ProtocolTCP},
		mesh_core.ProtocolTCP:  mesh_core.ProtocolList{mesh_core.ProtocolTCP},
	}
)

// getCommonProtocol returns a common protocol between given two.
//
// E.g.,
// a common protocol between HTTP and HTTP  is HTTP,
// a common protocol between HTTP and TCP   is TCP,
// a common protocol between GRPC and HTTP2 is HTTP2,
// a common protocol between HTTP and HTTP2 is TCP.
func getCommonProtocol(one, another mesh_core.Protocol) mesh_core.Protocol {
	if one == another {
		return one
	}
	if one == mesh_core.ProtocolUnknown || another == mesh_core.ProtocolUnknown {
		return mesh_core.ProtocolUnknown
	}
	oneProtocolStack, exist := protocolStacks[one]
	if !exist {
		return mesh_core.ProtocolUnknown
	}
	anotherProtocolStack, exist := protocolStacks[another]
	if !exist {
		return mesh_core.ProtocolUnknown
	}
	for _, firstProtocol := range oneProtocolStack {
		for _, secondProtocol := range anotherProtocolStack {
			if firstProtocol == secondProtocol {
				return firstProtocol
			}
		}
	}
	return mesh_core.ProtocolUnknown
}

// InferServiceProtocol returns a common protocol for a given group of endpoints.
func InferServiceProtocol(endpoints []core_xds.Endpoint) mesh_core.Protocol {
	if len(endpoints) == 0 {
		return mesh_core.ProtocolUnknown
	}
	serviceProtocol := mesh_core.ParseProtocol(endpoints[0].Tags[mesh_proto.ProtocolTag])
	for _, endpoint := range endpoints[1:] {
		endpointProtocol := mesh_core.ParseProtocol(endpoint.Tags[mesh_proto.ProtocolTag])
		serviceProtocol = getCommonProtocol(serviceProtocol, endpointProtocol)
	}
	return serviceProtocol
}
