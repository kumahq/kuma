package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var HealthCheckWsDefinition = ResourceWsDefinition{
	Name: "HealthCheck",
	Path: "health-checks",
	ResourceFactory: func() model.Resource {
		return mesh.NewHealthCheckResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.HealthCheckResourceList{}
	},
}
