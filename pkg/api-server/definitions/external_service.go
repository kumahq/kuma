package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var ExternalServiceWsDefinition = ResourceWsDefinition{
	Name: "ExternalService",
	Path: "external-services",
	ResourceFactory: func() model.Resource {
		return &mesh.ExternalServiceResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.ExternalServiceResourceList{}
	},
}
