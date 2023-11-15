package builders

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type MeshContextBuilder struct {
	res *xds_context.Context
}

func MeshContext() *MeshContextBuilder {
	return &MeshContextBuilder{
		res: &xds_context.Context{
			Mesh: xds_context.MeshContext{
				Resource: &core_mesh.MeshResource{
					Meta: &test_model.ResourceMeta{
						Name: "default",
					},
				},
			},
			ControlPlane: &xds_context.ControlPlaneContext{
				CLACache: &xds.DummyCLACache{OutboundTargets: map[core_xds.ServiceName][]core_xds.Endpoint{}},
				Zone:     "test-zone",
			},
		},
	}
}

func (mc *MeshContextBuilder) Build() *xds_context.Context {
	return mc.res
}

func (mc *MeshContextBuilder) With(fn func(*xds_context.Context)) *MeshContextBuilder {
	fn(mc.res)
	return mc
}

func (mc *MeshContextBuilder) WithEndpointMap(endpointMap core_xds.EndpointMap) *MeshContextBuilder {
	mc.res.Mesh.EndpointMap = endpointMap
	mc.res.ControlPlane.CLACache.(*xds.DummyCLACache).OutboundTargets = endpointMap
	return mc
}

func (mc *MeshContextBuilder) WithZone(zone string) *MeshContextBuilder {
	mc.res.ControlPlane.Zone = zone
	return mc
}

func (mc *MeshContextBuilder) WithResources(resources xds_context.Resources) *MeshContextBuilder {
	mc.res.Mesh.Resources = resources
	return mc
}

func (mc *MeshContextBuilder) WithMeshName(mesh string) *MeshContextBuilder {
	mc.res.Mesh.Resource.Meta.(*test_model.ResourceMeta).Name = mesh
	return mc
}

func (mc *MeshContextBuilder) WithMesh(mesh *builders.MeshBuilder) *MeshContextBuilder {
	mc.res.Mesh.Resource = mesh.Build()
	return mc
}
