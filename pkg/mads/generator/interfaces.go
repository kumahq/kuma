package generator

import (
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

type Args struct {
	Meshes       []*core_mesh.MeshResource
	Dataplanes   []*core_mesh.DataplaneResource
	MeshGateways []*core_mesh.MeshGatewayResource
}

type ResourceGenerator interface {
	Generate(Args) ([]*core_xds.Resource, error)
}
