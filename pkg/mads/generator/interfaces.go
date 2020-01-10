package generator

import (
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

type Context struct {
	Meshes     []*mesh_core.MeshResource
	Dataplanes []*mesh_core.DataplaneResource
}

type ResourceGenerator interface {
	Generate(Context) ([]*model.Resource, error)
}
