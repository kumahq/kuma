package samples

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func ServiceInsight(mesh string) *mesh.ServiceInsightResource {
	return ServiceInsightBuilderBuilder().
		WithMesh(mesh).
		AddService("test-service", &mesh_proto.ServiceInsight_Service{
			ServiceType: mesh_proto.ServiceInsight_Service_internal,
			Status:      mesh_proto.ServiceInsight_Service_online,
		}).
		Build()
}

func ServiceInsightBuilderBuilder() *builders.ServiceInsightBuilder {
	return builders.ServiceInsight()
}
