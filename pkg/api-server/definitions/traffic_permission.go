package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var TrafficPermissionWsDefinition = ResourceWsDefinition{
	Name: "Traffic Permission",
	Path: "traffic-permissions",
	ResourceFactory: func() model.Resource {
		return mesh.NewTrafficPermissionResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.TrafficPermissionResourceList{}
	},
}
