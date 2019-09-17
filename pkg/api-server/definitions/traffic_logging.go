package definitions

import (
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

var TrafficLoggingWsDefinition = ResourceWsDefinition{
	Name: "Traffic Logging",
	Path: "traffic-logging",
	ResourceFactory: func() model.Resource {
		return &mesh.TrafficLoggingResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.TrafficLoggingResourceList{}
	},
}
