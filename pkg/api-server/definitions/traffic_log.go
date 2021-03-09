package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var TrafficLogWsDefinition = ResourceWsDefinition{
	Name: "Traffic Logging",
	Path: "traffic-logs",
	ResourceFactory: func() model.Resource {
		return mesh.NewTrafficLogResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.TrafficLogResourceList{}
	},
}
