package mads

import (
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// TODO: perhaps move to a common generator package
type Args struct {
	Meshes     []*mesh_core.MeshResource
	Dataplanes []*mesh_core.DataplaneResource
}

type ResourceGenerator interface {
	Generate(Args) ([]*core_xds.Resource, error)
}
