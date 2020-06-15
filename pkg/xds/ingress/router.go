package ingress

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

func BuildDestinationMap(ingress *core_mesh.DataplaneResource) core_xds.DestinationMap {
	destinations := core_xds.DestinationMap{}
	for _, ingress := range ingress.Spec.GetNetworking().GetIngress().GetAvailableServices() {
		service := ingress.Tags[mesh_proto.ServiceTag]
		destinations[service] = destinations[service].Add(mesh_proto.MatchTags(ingress.Tags))
	}
	return destinations
}
