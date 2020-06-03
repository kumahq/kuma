package ingress

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

func BuildDestinationMap(ingress *core_mesh.DataplaneResource) core_xds.DestinationMap {
	destinations := core_xds.DestinationMap{}
	for _, inbound := range ingress.Spec.Networking.Inbound {
		service := inbound.GetTags()[mesh_proto.ServiceTag]
		destinations[service] = destinations[service].Add(inbound.GetTags())
	}
	return destinations
}
