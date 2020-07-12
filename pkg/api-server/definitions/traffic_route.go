package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
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
