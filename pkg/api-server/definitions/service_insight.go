package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var ServiceInsightWsDefinition = ResourceWsDefinition{
	Name: "Service Insight",
	Path: "service-insights",
	ResourceFactory: func() model.Resource {
		return mesh.NewServiceInsightResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.ServiceInsightResourceList{}
	},
}
