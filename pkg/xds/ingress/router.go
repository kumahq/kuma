package ingress

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

func BuildDestinationMap(ingress *core_mesh.DataplaneResource) core_xds.DestinationMap {
	destinations := core_xds.DestinationMap{}
	for _, ingress := range ingress.Spec.Networking.Ingress {
		tags := map[string]string{
			mesh_proto.ServiceTag: ingress.Service,
		}
		for k, v := range ingress.Tags {
			tags[k] = v
		}
		destinations[ingress.Service] = destinations[ingress.Service].Add(tags)
	}
	return destinations
}
