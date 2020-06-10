package ingress

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

func BuildDestinationMap(ingress *core_mesh.DataplaneResource) core_xds.DestinationMap {
	destinations := core_xds.DestinationMap{}
	for _, ingress := range ingress.Spec.Networking.Ingress {
		tags := mesh_proto.SingleValueTagSet(ingress.Tags).Add(mesh_proto.ServiceTag, ingress.Service)
		destinations[ingress.Service] = destinations[ingress.Service].Add(mesh_proto.MatchTags(tags))
	}
	return destinations
}
