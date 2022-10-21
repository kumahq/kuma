package samples

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func DataplaneBackendBuilder() *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithAddress("192.168.0.1").
		WithServices("backend")
}

func DataplaneBackend() *mesh.DataplaneResource {
	return DataplaneBackendBuilder().Build()
}

func DataplaneWebBuilder() *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithName("web-01").
		WithAddress("192.168.0.2").
		WithServices("web").
		AddOutboundToService("backend")
}

func DataplaneWeb() *mesh.DataplaneResource {
	return DataplaneWebBuilder().Build()
}
