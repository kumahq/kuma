package definitions

import (
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

var TrafficRouteWsDefinition = ResourceWsDefinition{
	Name: "TrafficRoute",
	Path: "traffic-routes",
	ResourceFactory: func() model.Resource {
		return &mesh.TrafficRouteResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.TrafficRouteResourceList{}
	},
}
