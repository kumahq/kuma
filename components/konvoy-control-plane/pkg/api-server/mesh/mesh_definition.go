package mesh

import (
	api_server "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
)

var MeshWsDefinition = api_server.ResourceWsDefinition{
	Name: "Mesh",
	Path: "meshes",
	ResourceFactory: func() model.Resource {
		return &mesh.MeshResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.MeshResourceList{}
	},
}
