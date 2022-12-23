package xds

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func CreateSampleMeshContext() xds_context.Context {
	return CreateSampleMeshContextWith(xds_context.NewResources())
}

func CreateSampleMeshContextWith(resources xds_context.Resources) xds_context.Context {
	return xds_context.Context{
		Mesh: xds_context.MeshContext{
			Resource: &core_mesh.MeshResource{
				Meta: &test_model.ResourceMeta{
					Name: "default",
				},
			},
			Resources: resources,
			EndpointMap: map[core_xds.ServiceName][]core_xds.Endpoint{
				"some-service": {
					{
						Tags: map[string]string{
							"app": "some-service",
						},
					},
				},
			},
		},
		ControlPlane: &xds_context.ControlPlaneContext{CLACache: &DummyCLACache{}, Zone: "test-zone"},
	}
}
