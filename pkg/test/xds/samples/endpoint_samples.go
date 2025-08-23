package samples

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
)

func HttpEndpointBuilder() *xds_builders.EndpointBuilder {
	return xds_builders.Endpoint().WithTags(mesh_proto.ProtocolTag, string(core_meta.ProtocolHTTP))
}

func TcpEndpointBuilder() *xds_builders.EndpointBuilder {
	return xds_builders.Endpoint().WithTags(mesh_proto.ProtocolTag, string(core_meta.ProtocolTCP))
}

func GrpcEndpointBuilder() *xds_builders.EndpointBuilder {
	return xds_builders.Endpoint().WithTags(mesh_proto.ProtocolTag, string(core_meta.ProtocolGRPC))
}
