package samples

import (
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	"github.com/kumahq/kuma/v2/pkg/test/xds/builders"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
)

func SampleContext() xds_context.Context {
	return SampleContextWith(xds_context.NewResources())
}

func SampleContextWith(resources xds_context.Resources) xds_context.Context {
	return *builders.Context().
		WithMeshBuilder(samples.MeshDefaultBuilder()).
		WithResources(resources).
		WithEndpointMap(
			builders.EndpointMap().
				AddEndpoint("some-service", builders.Endpoint().WithTags("app", "some-service")),
		).
		AddServiceProtocol("some-service", core_mesh.ProtocolUnknown).
		Build()
}
