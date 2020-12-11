package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var DataplaneWsDefinition = ResourceWsDefinition{
	Name: "Dataplane",
	Path: "dataplanes",
	ResourceFactory: func() model.Resource {
		return mesh.NewDataplaneResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.DataplaneResourceList{}
	},
}
