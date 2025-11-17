package samples

import (
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	xds_builders "github.com/kumahq/kuma/v2/pkg/test/xds/builders"
)

func HttpEndpointBuilder() *xds_builders.EndpointBuilder {
	return xds_builders.Endpoint().WithTags(mesh_proto.ProtocolTag, core_mesh.ProtocolHTTP)
}

func TcpEndpointBuilder() *xds_builders.EndpointBuilder {
	return xds_builders.Endpoint().WithTags(mesh_proto.ProtocolTag, core_mesh.ProtocolTCP)
}

func GrpcEndpointBuilder() *xds_builders.EndpointBuilder {
	return xds_builders.Endpoint().WithTags(mesh_proto.ProtocolTag, core_mesh.ProtocolGRPC)
}
