package samples

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func MeshMultiZoneServiceBackendBuilder() *builders.MeshMultiZoneServiceBuilder {
	return builders.MeshMultiZoneService().
		WithServiceLabelSelector(map[string]string{
			mesh_proto.DisplayName: "backend",
		}).
		AddIntPort(builders.FirstInboundPort, "http")
}

func MeshMultiZoneServiceBackend() *meshmzservice_api.MeshMultiZoneServiceResource {
	return MeshMultiZoneServiceBackendBuilder().Build()
}
