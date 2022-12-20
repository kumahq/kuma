package samples

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	test_xds "github.com/kumahq/kuma/pkg/test/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func CreateSimpleMeshContext() xds_context.Context {
	return CreateSimpleMeshContextWith(xds_context.NewResources())
}

func CreateSimpleMeshContextWith(resources xds_context.Resources) xds_context.Context {
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
		ControlPlane: &xds_context.ControlPlaneContext{CLACache: &test_xds.DummyCLACache{}, Zone: "test-zone"},
	}
}
