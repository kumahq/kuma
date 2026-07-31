package dataplane

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
)

// AdditionalInbounds no longer has any source of additional inbounds now that
// the legacy Prometheus metrics inbound has been removed. It stays as a
// no-op so callers don't need to be restructured if a future source is added.
func AdditionalInbounds(dataplane *core_mesh.DataplaneResource, mesh *core_mesh.MeshResource) ([]*mesh_proto.Dataplane_Networking_Inbound, error) {
	return nil, nil
}
