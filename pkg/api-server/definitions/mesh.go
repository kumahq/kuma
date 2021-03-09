package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var MeshWsDefinition = ResourceWsDefinition{
	Name: "Mesh",
	Path: "meshes",
	ResourceFactory: func() model.Resource {
		return mesh.NewMeshResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.MeshResourceList{}
	},
}
