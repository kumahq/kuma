package protocol

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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

// GetCommonProtocol returns a common protocol between given two.
//
// E.g.,
// a common protocol between HTTP and HTTP2 is HTTP2,
// a common protocol between HTTP and HTTP  is HTTP,
// a common protocol between HTTP and TCP   is TCP,
// a common protocol between GRPC and HTTP2 is HTTP2,
// a common protocol between HTTP and HTTP2 is HTTP.
func GetCommonProtocol(one, another core_mesh.Protocol) core_mesh.Protocol {
	switch {
	case one == another:
		return one
	case one == "" || another == "":
		return core_mesh.ProtocolUnknown
	case one == core_mesh.ProtocolUnknown || another == core_mesh.ProtocolUnknown:
		return core_mesh.ProtocolUnknown
	default:
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
}
