package samples

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func SampleContext() xds_context.Context {
	return SampleContextWith(xds_context.NewResources())
}

func SampleContextWith(resources xds_context.Resources) xds_context.Context {
	return *builders.Context().
		WithMesh(samples.MeshDefaultBuilder()).
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
