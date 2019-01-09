package generator

import (
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

type Args struct {
	Meshes     []*mesh_core.MeshResource
	Dataplanes []*mesh_core.DataplaneResource
}

type ResourceGenerator interface {
	Generate(Args) ([]*core_xds.Resource, error)
}
