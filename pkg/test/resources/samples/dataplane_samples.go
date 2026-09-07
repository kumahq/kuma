package samples

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
)

// DataplaneBackendBuilder carries the kuma.io/workload label matching
// MeshServiceBackendBuilder's label selector, so the two samples still
// pair up now that MeshService selection is labels-only.
func DataplaneBackendBuilder() *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithAddress("192.168.0.1").
		WithServices("backend").
		WithLabels(map[string]string{metadata.KumaWorkload: "backend"})
}

func DataplaneBackend() *mesh.DataplaneResource {
	return DataplaneBackendBuilder().Build()
}

func DataplaneWebBuilder() *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithName("web-01").
		WithAddress("192.168.0.2").
		WithInboundOfTagsAndProtocol("http").
		WithLabels(map[string]string{metadata.KumaWorkload: "web"}).
		AddOutboundToService("backend")
}

func DataplaneWeb() *mesh.DataplaneResource {
	return DataplaneWebBuilder().Build()
}

func GatewayDataplaneBuilder() *builders.DataplaneBuilder {
	return builders.Dataplane().
		WithName("sample-gateway").
		WithAddress("192.168.0.1").
		WithDelegatedGateway()
}

func IgnoredDataplaneBackendBuilder() *builders.DataplaneBuilder {
	return DataplaneBackendBuilder().With(func(resource *mesh.DataplaneResource) {
		resource.Spec.Networking.Inbound[0].State = mesh_proto.Dataplane_Networking_Inbound_Ignored
	})
}
