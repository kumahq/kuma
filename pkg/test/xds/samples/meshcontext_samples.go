package samples

import (
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/test/xds/builders"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

func SampleContext() xds_context.Context {
	builder := SampleContextWith(xds_context.NewResources())
	return *builder.Build()
}

func SampleContextWith(resources xds_context.Resources) *builders.ContextBuilder {
	return builders.Context().
		WithMeshBuilder(samples.MeshDefaultBuilder()).
		WithResources(resources).
		WithEndpointMap(
			builders.EndpointMap().
				AddEndpoint("some-service", builders.Endpoint().WithTags("app", "some-service")),
		).
		AddServiceProtocol("some-service", core_meta.ProtocolUnknown)
}
