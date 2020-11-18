package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var MeshInsightWsDefinition = ResourceWsDefinition{
	Name:     "Mesh Insight",
	Path:     "mesh-insights",
	ReadOnly: true,
	ResourceFactory: func() model.Resource {
		return &mesh.MeshInsightResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.MeshInsightResourceList{}
	},
}
