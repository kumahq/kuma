package samples

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func MeshServiceBackendBuilder() *builders.MeshServiceBuilder {
	return builders.MeshService().
		WithDataplaneTagsSelector(map[string]string{
			mesh_proto.ServiceTag: "backend",
		}).
		WithKumaVIP("240.0.0.1").
		AddIntPort(builders.FirstInboundPort, builders.FirstInboundPort, "http")
}

func MeshServiceBackend() *v1alpha1.MeshServiceResource {
	return MeshServiceBackendBuilder().Build()
}

func MeshServiceWebBuilder() *builders.MeshServiceBuilder {
	return builders.MeshService().
		WithName("web").
		WithDataplaneTagsSelector(map[string]string{
			mesh_proto.ServiceTag: "web",
		}).
		WithKumaVIP("240.0.0.2").
		AddIntPort(builders.FirstInboundPort, builders.FirstInboundPort, "http")
}

func MeshServiceWeb() *v1alpha1.MeshServiceResource {
	return MeshServiceBackendBuilder().Build()
}
