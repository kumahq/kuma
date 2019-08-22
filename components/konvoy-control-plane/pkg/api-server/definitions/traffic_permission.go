package definitions

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
)

var TrafficPermissionWsDefinition = ResourceWsDefinition{
	Name: "Traffic Permission",
	Path: "traffic-permission",
	ResourceFactory: func() model.Resource {
		return &mesh.TrafficPermissionResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.TrafficPermissionResourceList{}
	},
}
