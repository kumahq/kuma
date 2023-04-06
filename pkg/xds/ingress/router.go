package ingress

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

func BuildDestinationMap(mesh string, ingress *core_mesh.ZoneIngressResource) core_xds.DestinationMap {
	destinations := core_xds.DestinationMap{}
	for _, svc := range ingress.Spec.GetAvailableServices() {
		if mesh != svc.GetMesh() {
			continue
		}
		service := svc.Tags[mesh_proto.ServiceTag]
		destinations[service] = destinations[service].Add(mesh_proto.MatchTags(svc.Tags))
	}
	return destinations
}
