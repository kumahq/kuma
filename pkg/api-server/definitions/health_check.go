package definitions

import (
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

var HealthCheckWsDefinition = ResourceWsDefinition{
	Name: "HealthCheck",
	Path: "health-checks",
	ResourceFactory: func() model.Resource {
		return &mesh.HealthCheckResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.HealthCheckResourceList{}
	},
}
