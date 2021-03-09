package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var MeshInsightWsDefinition = ResourceWsDefinition{
	Name: "Mesh Insight",
	Path: "mesh-insights",
	ResourceFactory: func() model.Resource {
		return mesh.NewMeshInsightResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.MeshInsightResourceList{}
	},
}
