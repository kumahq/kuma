package samples

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func ServiceInsight() *builders.ServiceInsightBuilder {
	return builders.ServiceInsight().
		AddService("test-service", &mesh_proto.ServiceInsight_Service{
			ServiceType: mesh_proto.ServiceInsight_Service_internal,
			Status:      mesh_proto.ServiceInsight_Service_online,
		})
}
