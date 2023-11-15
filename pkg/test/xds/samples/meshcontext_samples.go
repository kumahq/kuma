package samples

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func SampleMeshContext() xds_context.Context {
	return SampleMeshContextWith(xds_context.NewResources())
}

func SampleMeshContextWith(resources xds_context.Resources) xds_context.Context {
	return *builders.MeshContext().
		WithMesh(samples.MeshDefaultBuilder()).
		WithZone("test-zone").
		WithResources(resources).
		WithEndpointMap(map[core_xds.ServiceName][]core_xds.Endpoint{
			"some-service": {
				{
					Tags: map[string]string{
						"app": "some-service",
					},
				},
			},
		}).
		Build()
}
