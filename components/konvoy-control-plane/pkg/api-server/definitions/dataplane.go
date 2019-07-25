package definitions

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
)

var DataplaneWsDefinition = ResourceWsDefinition{
	Name: "Dataplane",
	Path: "dataplanes",
	ResourceFactory: func() model.Resource {
		return &mesh.DataplaneStatusResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.DataplaneStatusResourceList{}
	},
}
