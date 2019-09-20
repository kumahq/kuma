package definitions

import (
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

var TrafficPermissionWsDefinition = ResourceWsDefinition{
	Name: "Traffic Permission",
	Path: "traffic-permissions",
	ResourceFactory: func() model.Resource {
		return &mesh.TrafficPermissionResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.TrafficPermissionResourceList{}
	},
}
