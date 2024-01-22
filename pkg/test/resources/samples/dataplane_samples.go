package samples

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func DataplaneBackendBuilder() *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithAddress("192.168.0.1").
		WithServices("backend")
}

func DataplaneBackend() *mesh.DataplaneResource {
	return DataplaneBackendBuilder().Build()
}

func DataplaneWebBuilder() *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithName("web-01").
		WithAddress("192.168.0.2").
		WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http").
		AddOutboundToService("backend")
}

func DataplaneWeb() *mesh.DataplaneResource {
	return DataplaneWebBuilder().Build()
}

func GatewayDataplaneBuilder() *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithName("sample-gateway").
		WithAddress("192.168.0.1").
		WithBuiltInGateway("sample-gateway")
}

func IgnoredDataplaneBackendBuilder() *builders.DataplaneBuilder {
	return DataplaneBackendBuilder().With(func(resource *mesh.DataplaneResource) {
		resource.Spec.Networking.Inbound[0].State = mesh_proto.Dataplane_Networking_Inbound_Ignored
	})
}
